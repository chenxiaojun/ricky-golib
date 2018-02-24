package exercise

import (
	"sync"
	"io"
	"errors"
	"log"
)

// 建立一个资源池已经关闭的错误
var ErrPoolClosed = errors.New("资源池已经关闭")

// 建立一个资源池结构体
type Pool struct {
	m sync.Mutex // 这是一个互斥锁
	res chan io.Closer // res是一个有缓冲的通道，用于保存共享的资源。 通道的类型是io.Closer接口
	factory func() (io.Closer, error) // 用于创建一个新的资源
	closed bool // 判断资源池是否被关闭
}

// 定义一个可以创建连接池的方法
func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("size的值太小了。")
	}
	return &Pool {
		factory: fn,
		res: make(chan io.Closer, size),
	}, nil
}

// 从资源池里获取资源
func (p *Pool) Acquire() (io.Closer, error) {
	select {
	case r,ok := <-p.res:
		log.Println("Acquire: 共享资源")
		if !ok {
			return nil, ErrPoolClosed
		}
		return r,nil
	default:
		log.Println("Acquire: 新生成资源")
		return p.factory()
	}
}

// 关闭资源池，释放资源
func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	// 关闭通道，不让写入
	close(p.res)

	// 关闭通道里的资源
	for r := range p.res {
		r.Close()
	}
}

// 释放资源
func (p *Pool) Release(r io.Closer) {
	// 保证该操作和Close方法的操作是安全的
	p.m.Lock()
	defer p.m.Unlock()

	// 资源池都关闭了，就剩这一个没有释放的了，释放即可
	if p.closed {
		r.Close()
		return
	}

	select {
	case p.res <- r:
		log.Println("资源释放到池子里了")
	default:
		log.Println("资源池满了，释放这个资源吧")
		r.Close()
	}
}








































