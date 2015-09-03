package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sync"
)

type Package struct {
	Name  string
	Lines int
}

func plural(num int, str, p string) string {
	if num == 1 {
		return str
	}

	return str + p
}

var (
	wg   sync.WaitGroup
	done = NewDone()
)

func countLines(linesC chan<- Package, pkg *build.Package) {
	defer wg.Done()

	for _, ipath := range pkg.Imports {
		if done.Check(ipath) {
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

	for _, file := range pkg.GoFiles {
		path := filepath.Join(pkg.Dir, file)

		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			file, err := os.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open %q: %v", path, err)
				os.Exit(1)
			}
			defer file.Close()
		}(path)
	}
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

	for _, ipath := range flag.Args() {
		if done.Check(ipath) {
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
