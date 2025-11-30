package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/james-hester/piccolo/internal"
)

func main() {
	asm := flag.Bool("S", false, "Print assembly output")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: piccolo [options] <input.pic>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputFile := flag.Arg(0)
	content, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	tokens, err := internal.Lex(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lexing error: %v\n", err)
		os.Exit(1)
	}

	prog, err := internal.Parse(tokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parsing error: %v\n", err)
		os.Exit(1)
	}

	ops, syms, err := internal.Compile(prog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
		os.Exit(1)
	}

	if *asm {
		for _, op := range ops {
			fmt.Println(op.Assembly())
		}
		return
	}

	words, err := internal.Assemble(ops, syms)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly error: %v\n", err)
		os.Exit(1)
	}

	outputFile := strings.TrimSuffix(inputFile, filepath.Ext(inputFile)) + ".hex"
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := internal.WriteHex(f, words); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing hex file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully compiled %s to %s\n", inputFile, outputFile)
}
