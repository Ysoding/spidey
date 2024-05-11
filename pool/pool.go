package pool

import (
	"sync"
)

type Worker interface {
	Work(context interface{})
}

type Task struct {
	context interface{}
	work    Worker
}

type Pool struct {
	tasks chan Task
	wg    sync.WaitGroup
	kill  chan struct{}
}

func New() *Pool {
	pool := &Pool{
		tasks: make(chan Task),
		kill:  make(chan struct{}),
	}

	pool.start()

	return pool
}

func (p *Pool) start() {
	for i := 0; i < 1_00; i++ {
		p.wg.Add(1)
		go p.work()
	}
}

func (p *Pool) work() {
	defer p.wg.Done()
done:
	for {
		select {
		case task := <-p.tasks:
			p.execute(task)
		case <-p.kill:
			break done
		}
	}
}

func (p *Pool) execute(task Task) {
	task.work.Work(task.context)
}

func (p *Pool) Do(context interface{}, work Worker) {
	p.tasks <- Task{
		context: context,
		work:    work,
	}
}

func (p *Pool) Shutdown() {
	close(p.kill)
	p.wg.Wait()
}
