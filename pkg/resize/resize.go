package resize

import (
	"context"

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

type Resizer struct {
	ctx  context.Context
	stop func()
	add  chan Job
	next chan Job
	log  *logrus.Logger
	q    []Job
}

func New() *Resizer {
	return &Resizer{
		ctx:  context.Background(),
		q:    make([]Job, 0, 500),
		add:  make(chan Job, 100), // put new jobs on the queue
		next: make(chan Job),      // next job to be done
	}
}

func (r *Resizer) Start(ctx context.Context) {
	r.log = ctx.Value("log").(*logrus.Logger)
	r.ctx, r.stop = context.WithCancel(ctx)
	r.log.Info("beginning resizer process loop")
	go r.Process()
	go r.sort()
}

func (r *Resizer) CreateJob(src string, dest string, px int, opts ...JobOpt) Job {
	l := r.log.WithFields(logrus.Fields{
		"src":  src,
		"opts": opts,
		"dest": dest,
		"px":   px,
	})
	l.Trace("creating resizer job instance")
	ctx, cancel := context.WithCancel(r.ctx)
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
func (r *Resizer) Add(j Job, p Priority) {
	j.priority = p
	r.log.WithField("job", j).Trace("adding job to queue")
	r.add <- j
}

// empty the queue
func (r *Resizer) Clear() error {

	return nil

	// should this cancel an in-flight generation? separate call for that?
}

func (r *Resizer) Halt() {
	if r.stop != nil {
		r.stop()
	}
}

// this is the only place r.q is allowed to be touched
func (r *Resizer) sort() {
	for {
		select {
		case <-r.ctx.Done():
			r.log.Info("resizer context done. Exiting queue sorter")
			return
		case job := <-r.add:
			r.q = append(r.q, job)
		default:
			// if there is no queue, wait for an addition instead of churning through this loop
			// and hitting default all the time
			if len(r.q) == 0 {
				job := <-r.add
				r.q = append(r.q, job)
			}

			// otherwise send a job
			priority := Priority(-1)
			idx := 0
			for i, j := range r.q {
				if j.priority > priority {
					idx = i
					priority = j.priority
				}
			}
			r.next <- r.q[idx]
			// and now remove it
			copy(r.q[idx:], r.q[idx+1:])
			r.q[len(r.q)-1] = Job{} // erase with empty value
			r.q = r.q[:len(r.q)-1]
		}
	}
}

// @todo: contexts and cancellations in the queue
// @todo: status and progress

func (r *Resizer) Process() {
	for {
		select {
		case <-r.ctx.Done():
			r.log.Info("resizer context done. Exiting process loop")
			return
		case job := <-r.next:
			r.log.WithField("dst", job.dest).WithField("priority", job.priority).Debug("pulling job to process")
			r.runDarktable(job)
			job.Done <- struct{}{}
		}
	}
}
