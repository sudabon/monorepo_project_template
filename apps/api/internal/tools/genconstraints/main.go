package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "genconstraints: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("genconstraints", flag.ContinueOnError)
	specPath := fs.String("spec", "", "path to OpenAPI YAML")
	goPath := fs.String("go", "", "path to write generated Go constants")
	tsPath := fs.String("ts", "", "path to write generated TypeScript constants")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *specPath == "" {
		return fmt.Errorf("required flag -spec")
	}
	if *goPath == "" && *tsPath == "" {
		return fmt.Errorf("at least one of -go or -ts is required")
	}
	doc, err := openapi3.NewLoader().LoadFromFile(*specPath)
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}
	constraints := extract(doc)
	if *goPath != "" {
		src, err := generateGo(constraints)
		if err != nil {
			return err
		}
		if err := os.WriteFile(*goPath, []byte(src), 0o644); err != nil {
			return err
		}
	}
	if *tsPath != "" {
		if err := os.WriteFile(*tsPath, []byte(generateTS(constraints)), 0o644); err != nil {
			return err
		}
	}
	return nil
}
