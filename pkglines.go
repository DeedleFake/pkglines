package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
)

type Package struct {
	Name  string
	Lines int
}

func countLines(linesC chan Package, pkg *build.Package) {
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

	linesC := make(chan Package, flag.NArg())

	done := make(map[string]struct{}, flag.NArg())
	for _, ipath := range flag.Args() {
		if _, ok := done[ipath]; ok {
			continue
		}
		done[ipath] = struct{}{}

		pkg, err := build.Import(ipath, ".", 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to import %q: %v", ipath, err)
			os.Exit(1)
		}

		go countLines(linesC, pkg)
	}
}
