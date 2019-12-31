package darktable

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Job struct {
	ctx      context.Context
	Cancel   func()
	Done     chan struct{}
	priority Priority
	size     int
	source   string
	xmp      string
	dest     string
	hq       bool // considerable time difference, with quality change mostly in fine sharpness
}

type JobOpt func(*Job)

func SetHQ(hq bool) JobOpt     { return func(j *Job) { j.hq = hq } }
func SetXMP(xmp string) JobOpt { return func(j *Job) { j.xmp = xmp } }

type Exporter struct {
	ctx  context.Context
	stop func()
	add  chan Job
	next chan Job
	Log  *logrus.Logger
	q    []Job
}

func New() *Exporter {
	return &Exporter{
		ctx:  context.Background(),
		q:    make([]Job, 0, 500),
		add:  make(chan Job, 100), // put new jobs on the queue
		next: make(chan Job),      // next job to be done
	}
}

func (e *Exporter) Start(ctx context.Context) {
	e.ctx, e.stop = context.WithCancel(ctx)
	e.Log.Info("beginning darktable process loop")
	go e.Process()
	go e.sort()
}

func (e *Exporter) CreateJob(src string, dest string, px int, opts ...JobOpt) Job {
	l := e.Log.WithFields(logrus.Fields{
		"src":  src,
		"opts": opts,
		"dest": dest,
		"px":   px,
	})
	l.Trace("creating darktable job instance")
	ctx, cancel := context.WithCancel(e.ctx)
	j := Job{
		ctx:    ctx,
		Cancel: cancel,
		Done:   make(chan struct{}),
		size:   px,
		source: src,
		dest:   dest,
		xmp:    "",
		hq:     false,
	}
	for _, o := range opts {
		if o != nil {
			o(&j)
		}
	}
	return j
}

type Priority int

const (
	PR_LOW       Priority = 10
	PR_NORMAL    Priority = 50
	PR_HIGH      Priority = 70
	PR_IMMEDIATE Priority = 100
)

// add a thumbnail request to the queue
// to be done in priority order
func (e *Exporter) Add(j Job, p Priority) {
	j.priority = p
	e.Log.WithField("job", j).Trace("adding job to queue")
	e.add <- j
}

// empty the queue
func (e *Exporter) Clear() error {

	return nil

	// should this cancel an in-flight generation? separate call for that?
}

func (e *Exporter) Halt() {
	if e.stop != nil {
		e.stop()
	}
}

// this is the only place e.q is allowed to be touched
func (e *Exporter) sort() {
	for {
		select {
		case <-e.ctx.Done():
			e.Log.Info("darktable context done. Exiting queue sorter")
			return
		case job := <-e.add:
			e.q = append(e.q, job)
		default:
			// if there is no queue, wait for an addition instead of churning through this loop
			// and hitting default all the time
			if len(e.q) == 0 {
				job := <-e.add
				e.q = append(e.q, job)
			}

			// otherwise send a job
			priority := Priority(-1)
			idx := 0
			for i, j := range e.q {
				if j.priority > priority {
					idx = i
					priority = j.priority
				}
			}
			e.next <- e.q[idx]
			// and now remove it
			copy(e.q[idx:], e.q[idx+1:])
			e.q[len(e.q)-1] = Job{} // erase with empty value
			e.q = e.q[:len(e.q)-1]
		}
	}
}

// @todo: contexts and cancellations in the queue
// @todo: status and progress

func (e *Exporter) Process() {
	for {
		select {
		case <-e.ctx.Done():
			e.Log.Info("darktable context done. Exiting process loop")
			return
		case job := <-e.next:
			e.Log.WithField("dst", job.dest).WithField("priority", job.priority).Debug("pulling job to process")
			e.do(job)
			job.Done <- struct{}{}
		}
	}
}

func (e *Exporter) do(j Job) {
	e.Log.Trace("starting job")
	defer func() {
		if j.Cancel != nil {
			j.Cancel()
		}
	}()
	maxsize := strconv.Itoa(j.size)

	args := []string{j.source, j.dest, "--width", maxsize, "--height", maxsize, "--hq", strconv.FormatBool(j.hq)}
	if j.xmp != "" { // insert xmp argument if present
		args = append(args, "")
		copy(args[2:], args[1:]) // shift
		args[1] = j.xmp
	}

	e.Log.WithField("args", args).Debug("calling darktable-cli")
	cmd := exec.CommandContext(j.ctx, "darktable-cli", args...)

	var buf bytes.Buffer
	sout, o_err := cmd.StdoutPipe()
	if o_err != nil {
		e.Log.WithError(o_err).Error("error getting stdout of darktable process")
	}
	serr, e_err := cmd.StderrPipe()
	if e_err != nil {
		e.Log.WithError(e_err).Error("error getting stderr of darktable process")
	}

	if err := cmd.Start(); err != nil {
		e.Log.WithError(err).Error("error starting darktable-cli process. Unable to perform job")
		return
	}

	if o_err == nil {
		io.Copy(&buf, sout) // nolint
	}
	if e_err == nil {
		io.Copy(&buf, serr) // nolint
	}

	if buf.Len() > 0 {
		e.Log.Trace(buf.String())
	}

	err := cmd.Wait()
	if err != nil {
		e.Log.WithError(err).Error("darktable-cli exit error")
	} else {
		e.Log.Info("darktable-cli exited successfully")
	}
}
