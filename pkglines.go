package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type Package struct {
	Name  string
	Lines int64
}

func plural(num int64, str, p string) string {
	if num == 1 {
		return str
	}

	return str + p
}

var (
	wg     sync.WaitGroup
	filter *Filter
)

func countLines(linesC chan<- Package, pkg *build.Package) {
	defer wg.Done()

	for _, ipath := range pkg.Imports {
		pkg, err := build.Import(ipath, ".", 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to import %q: %v", ipath, err)
			os.Exit(1)
		}

		if filter.Check(pkg) {
			continue
		}

		wg.Add(1)
		go countLines(linesC, pkg)
	}

	var lines int64
	var wg sync.WaitGroup

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

			s := bufio.NewScanner(file)
			for s.Scan() {
				atomic.AddInt64(&lines, 1)
			}
		}(path)
	}

	wg.Wait()

	linesC <- Package{
		Name:  pkg.ImportPath,
		Lines: lines,
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [options] <import path> ...\n", os.Args[0])
		flag.PrintDefaults()
	}

	std := flag.Bool("std", false, "Count lines in standard library packages.")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	linesC := make(chan Package)
	filter = NewFilter(func(pkg *build.Package, prev bool) bool {
		if prev {
			return true
		}

		if !*std && pkg.Goroot {
			return true
		}

		return false
	})

	for _, ipath := range flag.Args() {
		pkg, err := build.Import(ipath, ".", 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to import %q: %v", ipath, err)
			os.Exit(1)
		}

		if filter.Check(pkg) {
			continue
		}

		wg.Add(1)
		go countLines(linesC, pkg)
	}

	go func() {
		wg.Wait()
		close(linesC)
	}()

	var total int64
	for pkg := range linesC {
		fmt.Printf("%v: %v %v.\n", pkg.Name, pkg.Lines, plural(pkg.Lines, "line", "s"))
		total += pkg.Lines
	}

	fmt.Println()
	fmt.Printf("%v %v total.\n", total, plural(total, "line", "s"))
}
