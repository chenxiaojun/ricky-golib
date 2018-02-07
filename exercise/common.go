package common

import (
	"errors"
	"time"
	"os"
	"os/signal"
)

var ErrTimeOut = errors.New("执行者执行超时")
var ErrInterrupt = errors.New("执行者被中断")

type Runner struct {
	tasks []func(int) // 要执行的任务
	complete chan error // 用于通知任务全部完成
	timeout <-chan time.Time // 这些任务在多久内完成
	interrupt chan os.Signal // 可以控制强制终止的信号
}

func New(tm time.Duration) *Runner {
	return &Runner {
		complete: make(chan error),
		timeout: time.After(tm),
		interrupt: make(chan os.Signal, 1),
	}
}

// 将需要执行的任务添加到runner
func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

func (r *Runner) run() error {
	for id, task := range r.tasks {
		if r.isInterrupt() {
			return ErrInterrupt
		}
		task(id)
	}
	return nil
}

func (r *Runner) isInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}

/**
	func Notify(c chan<- os.Signal, sig ...os.Signal)
	第一个参数： 表示接收信号的管道
 	第二个参数及后面的参数列表： 表示设置要监听的信号, 如果不设置，表示监听所有信号
		信道c是不可以阻塞的，如果信道缓冲不足的话，可能会丢失信号。 如果不再次转发，可以设置1个缓冲大小就可以了
 */
func(r *Runner) Start() error {
	signal.Notify(r.interrupt, os.Interrupt)

	go func() {
		r.complete <- r.run()
	}()

	select {
	case err := <-r.complete:
		return err
	case <-r.timeout:
		return ErrTimeOut
	}
}

