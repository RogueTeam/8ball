package utils

type JobsPull struct {
	size int
	jobs chan struct{}
}

func (p *JobsPull) Get() {
	<-p.jobs
}

func (p *JobsPull) Put() {
	p.jobs <- struct{}{}
}

func NewJobPull(size int) (j *JobsPull) {
	j = &JobsPull{jobs: make(chan struct{}, size)}
	for range size {
		j.jobs <- struct{}{}
	}
	return j
}
