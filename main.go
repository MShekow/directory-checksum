package main

import (
	"flag"
	"fmt"
	"golang-exp1/directory_checksum"
	"log"
	"os"
)

var maxDepth int

func init() {
	flag.IntVar(&maxDepth, "max-depth", 2, "Max directory depth (level) of the listing to be printed")
}

func main() {
	flag.CommandLine.SetOutput(os.Stdout) // ensure that flag.PrintDefaults() does NOT print to stderr by default
	flag.Usage = func() {
		fmt.Println("Usage of Directory Checksum Tool v1.0:")
		fmt.Println("directory-checksum [--max-depth=N] <path>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("You must provide exactly one argument: the absolute or relative path to the directory \n" +
			"to be scanned (may just be a dot for the current working directory)")
	}
	if maxDepth < 0 {
		log.Fatal("max-depth argument must be 0 or larger")
	}

	root := flag.Arg(0)
	directory := directory_checksum.ScanDirectory(root)
	directory.ComputeDirectoryChecksums()
	output := directory.PrintChecksums(".", maxDepth)
	fmt.Print(output)
}
