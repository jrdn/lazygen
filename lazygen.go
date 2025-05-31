package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/jrdn/lazygen/pkg"
)

type LazyGenOptions struct {
	SrcFile string
	Print   bool
}

func main() {
	var printOutput bool
	flag.BoolVar(&printOutput, "printOutput", false, "Print to stdout instead of writing the file")
	flag.Parse()

	options := LazyGenOptions{
		SrcFile: flag.Arg(0),
		Print:   printOutput,
	}

	err := lazyGen(options)
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	}
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
		fmt.Printf("\n%s content:\n\n", generator.Outfile())
		fmt.Println(data)
	} else {
		fmt.Printf("Generating to destination file: %s\n", generator.Outfile())
		if err := ioutil.WriteFile(generator.Outfile(), []byte(data), os.FileMode(0o644)); err != nil {
			return err
		}
	}

	cmd := exec.Command("gofmt", "-w", generator.Outfile())
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
