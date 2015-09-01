package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [import path] ...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}
