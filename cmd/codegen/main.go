package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codegen/internal/core"
	"github.com/codegen/pkg/gen"
)

func main() {
	locFlag := flag.String("c", "", "location of the .codegen folder")
	flag.Parse()

	spec, err := core.NewSpec(*locFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := gen.Execute(spec); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
