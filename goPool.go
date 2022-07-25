package easyPool

import (
	"errors"
	"sync"
)

type task struct {
	proccess func([]interface{}) interface{}
	callback func(interface{})
	args     []interface{}
}

type goPool struct {
	taskChannel    chan *task
	maxWaitingNum  int
	blockThreshold int
	passThreshold  int
	blockState     bool
	openState      bool
	waitingTask    int
	wg             sync.WaitGroup

	waitingtaskLock sync.Mutex
}

func NewPool(maxWaitingNum, blockThreshold, passThreshold int) *goPool {
	pool := &goPool{
		taskChannel:    make(chan *task, maxWaitingNum),
		maxWaitingNum:  maxWaitingNum,
		blockThreshold: blockThreshold,
		passThreshold:  passThreshold,
		blockState:     false,
		openState:      true,
		waitingTask:    0,
	}

	go pool.scheduler()

	return pool
}

func (p *goPool) Add(proccess func([]interface{}) interface{}, callBack func(interface{}), args ...interface{}) error {
	if !p.openState {
		return errors.New("pool closed")
	}
	for p.blockState {
	}

	p.taskChannel <- &task{
		proccess: proccess,
		callback: callBack,
		args:     args,
	}

	p.waitingtaskLock.Lock()
	p.waitingTask++
	if p.waitingTask >= p.blockThreshold {
		p.blockState = true
	}
	p.waitingtaskLock.Unlock()

	return nil
}

func (p *goPool) Close() {
	p.openState = false
	p.wg.Wait()
}

func (p *goPool) scheduler() {
	for {
		if !p.openState && p.waitingTask == 0 {
			return
		}

		select {
		case task := <-p.taskChannel:
			go func() {
				task.callback(task.proccess(task.args))
				p.wg.Done()
			}()
			p.waitingtaskLock.Lock()
			p.waitingTask--
			if p.passThreshold >= p.waitingTask {
				p.blockState = false
			}
			p.waitingtaskLock.Unlock()

			p.wg.Add(1)
		}
	}
}
