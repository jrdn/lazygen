package main

import (
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"

	"github.com/jrdn/lazygen/pkg"
)

type LazyGenOptions struct {
	SrcFile string
	Print   bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("lazygen", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	var printOutput bool
	flags.BoolVar(&printOutput, "printOutput", false, "Print to stdout instead of writing the file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("expected exactly one source file, got %d", flags.NArg())
	}

	return lazyGen(LazyGenOptions{
		SrcFile: flags.Arg(0),
		Print:   printOutput,
	})
}

func lazyGen(options LazyGenOptions) error {
	fmt.Printf("Generating from source file: %s\n", options.SrcFile)

	generator, err := pkg.NewParser(options.SrcFile)
	if err != nil {
		return err
	}

	data, err := generator.Generate()
	if err != nil {
		return err
	}
	if options.Print {
		formatted, err := format.Source([]byte(data))
		if err != nil {
			return fmt.Errorf("format generated output: %w", err)
		}
		fmt.Printf("\n%s content:\n\n", generator.Outfile())
		_, err = os.Stdout.Write(formatted)
		return err
	}

	fmt.Printf("Generating to destination file: %s\n", generator.Outfile())
	formatted, err := format.Source([]byte(data))
	if err != nil {
		return fmt.Errorf("format generated output: %w", err)
	}
	if err := ioutil.WriteFile(generator.Outfile(), formatted, os.FileMode(0o644)); err != nil {
		return err
	}

	return nil
}
