package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cookieo9/packager/lib/packager"
	"golang.org/x/tools/imports"
)

var (
	out = flag.String("output", "", "output file name ('-' for stdout)")
	pkg = flag.String("package", ".", "package to parse")

	processor packager.Processor
)

func init() {
	flag.StringVar(&processor.Local, "local", "defaultValue", "variable holding local instance")
	flag.StringVar(&processor.Allow, "allow", "", "regexp of allowed method names")
	flag.StringVar(&processor.Block, "block", "^$", "regexp of blocke method names")
}

func outfile() string {
	if *out == "" || *out == "-" {
		return processor.Local + ".funcs.go"
	}
	return *out
}

func output(code []byte) error {
	if *out == "-" {
		_, err := os.Stdout.Write(code)
		return err
	}
	return ioutil.WriteFile(outfile(), code, 0644)
}

func format(filename string, data []byte) ([]byte, error) {
	res, err := imports.Process(filename, data, nil)
	if err != nil {
		return nil, fmt.Errorf("can't format output: %w", err)
	}
	return res, nil
}

func action() error {
	pkgs, err := packager.LoadPackage(*pkg)
	if err != nil {
		return err
	}
	log.Printf("Load: %+v", pkgs)

	b := new(strings.Builder)

	fmt.Fprintf(b, "// Code generated by packager %q -- DO NOT EDIT\n", os.Args[1:])
	fmt.Fprintln(b, "// +build !packager")
	fmt.Fprintln(b)

	if err := processor.Process(b, pkgs); err != nil {
		return err
	}

	if formatted, err := format(outfile(), []byte(b.String())); err != nil {
		return err
	} else {
		return output(formatted)
	}
}

func main() {
	flag.Parse()

	if err := action(); err != nil {
		log.Fatal(err)
	}
}
