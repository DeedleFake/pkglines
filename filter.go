package main

import (
	"go/build"
	"sync"
)

type filterCheck struct {
	Pkg *build.Package
	Ret chan<- bool
}

type Filter struct {
	filter  func(*build.Package, bool) bool
	done    map[string]struct{}
	check   chan *filterCheck
	retPool *sync.Pool
}

func NewFilter(f func(*build.Package, bool) bool) *Filter {
	filter := &Filter{
		filter: f,

		done:  make(map[string]struct{}),
		check: make(chan *filterCheck),

		retPool: &sync.Pool{
			New: func() interface{} {
				return make(chan bool, 1)
			},
		},
	}

	go filter.db()

	return filter
}

func (d *Filter) db() {
	for check := range d.check {
		_, ok := d.done[check.Pkg.ImportPath]
		check.Ret <- d.filter(check.Pkg, ok)

		d.done[check.Pkg.ImportPath] = struct{}{}
	}
}

func (d *Filter) Check(pkg *build.Package) bool {
	ret := d.retPool.Get().(chan bool)
	defer d.retPool.Put(ret)

	d.check <- &filterCheck{
		Pkg: pkg,
		Ret: ret,
	}

	return <-ret
}
