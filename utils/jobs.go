package utils

type JobsPull struct {
	jobs chan struct{}
}

func (p *JobsPull) Get() {
	<-p.jobs
}

func (p *JobsPull) Put() {
	p.jobs <- struct{}{}
}

func NewJobPull(size int) (j *JobsPull) {
	return &JobsPull{jobs: make(chan struct{}, size)}
}
