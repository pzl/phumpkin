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
	priority int
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
	next chan Job
	Log  *logrus.Logger
	q    []Job //@todo -- needs to be synchronized between processor thread and adder thread
}

func New() *Exporter {
	return &Exporter{
		ctx:  context.Background(),
		q:    make([]Job, 0, 50),
		next: make(chan Job, 600),
	}
}

func (e *Exporter) Start(ctx context.Context) {
	e.ctx, e.stop = context.WithCancel(ctx)
	e.Log.Info("beginning darktable process loop")
	go e.Process()
}

// add a thumbnail request to the queue
// to be done in priority order
func (e *Exporter) Add() error {
	return nil
}

// force immediate generation of request, skipping queue
// returns Done channel
func (e *Exporter) Immediate(src string, dest string, px int, opts ...JobOpt) (Job, error) {
	l := e.Log.WithFields(logrus.Fields{
		"src":  src,
		"opts": opts,
		"dest": dest,
		"px":   px,
	})
	l.Trace("creating immediate darktable job")
	ctx, cancel := context.WithCancel(e.ctx)
	done := make(chan struct{})

	j := Job{
		ctx:      ctx,
		Cancel:   cancel,
		Done:     done,
		priority: 100,
		size:     px,
		source:   src,
		dest:     dest,
		xmp:      "",
		hq:       false,
	}
	for _, o := range opts {
		if o != nil {
			o(&j)
		}
	}

	e.next <- j

	return j, nil
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

// @todo: contexts and cancellations in the queue
// @todo: status and progress

func (e *Exporter) Process() {
	for {
		select {
		case <-e.ctx.Done():
			e.Log.Info("darktable process loop cancelled. exiting")
			return
		case job := <-e.next:
			e.Log.WithField("dst", job.dest).Debug("pulling job to process")
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
