package main

import (
	"sync"
)

type doneCheck struct {
	Name string
	Ret  chan<- bool
}

type Done struct {
	done    map[string]struct{}
	check   chan *doneCheck
	retPool *sync.Pool
}

func NewDone() *Done {
	done := &Done{
		done:  make(map[string]struct{}),
		check: make(chan *doneCheck),

		retPool: &sync.Pool{
			New: func() interface{} {
				return make(chan bool, 1)
			},
		},
	}

	go done.db()

	return done
}

func (d *Done) db() {
	for check := range d.check {
		_, ok := d.done[check.Name]
		check.Ret <- ok

		d.done[check.Name] = struct{}{}
	}
}

func (d *Done) Check(name string) bool {
	ret := d.retPool.Get().(chan bool)
	defer d.retPool.Put(ret)

	d.check <- &doneCheck{
		Name: name,
		Ret:  ret,
	}

	return <-ret
}
