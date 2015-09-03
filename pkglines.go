package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"sync"
)

type Package struct {
	Name  string
	Lines int
}

type CheckDone struct {
	Name string
	Ret  chan<- bool
}

func plural(num int, str, p string) string {
	if num == 1 {
		return str
	}

	return str + p
}

var (
	wg        sync.WaitGroup
	checkDone = make(chan *CheckDone)
)

func countLines(linesC chan<- Package, pkg *build.Package) {
	defer wg.Done()
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [import path] ...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	linesC := make(chan Package)

	go func() {
		done := make(map[string]struct{}, flag.NArg())
		for check := range checkDone {
			_, ok := done[check.Name]
			check.Ret <- ok

			done[check.Name] = struct{}{}
		}
	}()

	checkC := make(chan bool)
	for _, ipath := range flag.Args() {
		checkDone <- &CheckDone{
			Name: ipath,
			Ret:  checkC,
		}
		if <-checkC {
			continue
		}

		pkg, err := build.Import(ipath, ".", 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to import %q: %v", ipath, err)
			os.Exit(1)
		}

		wg.Add(1)
		go countLines(linesC, pkg)
	}

	go func() {
		wg.Wait()
		close(linesC)
	}()

	var total int
	for pkg := range linesC {
		fmt.Printf("%v: %v %v\n", pkg.Name, pkg.Lines, plural(pkg.Lines, "line", "s"))
		total += pkg.Lines
	}
	fmt.Printf("%v %v total.", total, plural(total, "line", "s"))
}
