package main

import (
	"flag"
	"fmt"
	"github.com/MShekow/directory-checksum/directory_checksum"
	"github.com/go-errors/errors"
	"github.com/spf13/afero"
	"log"
	"os"
)

const version = "1.4"

var maxDepth int

func init() {
	flag.IntVar(&maxDepth, "max-depth", 2, "Max directory depth (level) of the listing to be printed")
}

func main() {
	flag.CommandLine.SetOutput(os.Stdout) // ensure that flag.PrintDefaults() does NOT print to stderr by default
	flag.Usage = func() {
		fmt.Printf("Usage of Directory Checksum Tool %s:\n\n", version)
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
	directory, err := directory_checksum.ScanDirectory(root, afero.NewOsFs())
	if err != nil {
		if errorWithStacktrace, ok := err.(*errors.Error); ok {
			fmt.Println("Unable to scan the directory:")
			fmt.Println(errorWithStacktrace.ErrorStack())
		} else {
			fmt.Printf("Unable to scan the directory: %v\n", err)
		}
		os.Exit(1)
	}
	_, err = directory.ComputeDirectoryChecksums()
	if err != nil {
		if errorWithStacktrace, ok := err.(*errors.Error); ok {
			fmt.Println("Unexpected error while computing directory checksum:")
			fmt.Println(errorWithStacktrace.ErrorStack())
		} else {
			fmt.Printf("Unexpected error while computing directory checksums: %v\n", err)
		}
		os.Exit(1)
	}
	output := directory.PrintChecksums(maxDepth)
	fmt.Print(output)
}
