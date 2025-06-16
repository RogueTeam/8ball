package utils

import "sync/atomic"

type JobsPool struct {
	size    int
	stopped atomic.Bool
	jobs    chan struct{}
}

func (p *JobsPool) C() (ch <-chan struct{}) {
	return p.jobs
}

func (p *JobsPool) Get() {
	if !p.stopped.Load() {
		<-p.jobs
	}
}

func (p *JobsPool) Put() {
	if !p.stopped.Load() {
		p.jobs <- struct{}{}
	}
}

func (p *JobsPool) Stop() {
	if !p.stopped.Load() {
		close(p.jobs)
	}
	p.stopped.Store(true)
}

func NewJobPool(size int) (j *JobsPool) {
	j = &JobsPool{jobs: make(chan struct{}, size)}
	for range size {
		j.jobs <- struct{}{}
	}
	return j
}
