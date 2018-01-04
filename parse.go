package main

import (
	"fmt"
	"os"

	"github.schq.secious.com/jason-miller/go-openapi-v3/generator"

	"github.schq.secious.com/jason-miller/go-openapi-v3/parser"

	"github.com/googleapis/gnostic/OpenAPIv3"
	"github.com/googleapis/gnostic/compiler"
)

const filepath = "examples/demo/requests.yaml"

func main() {
	bytes, err := compiler.ReadBytesForFile(filepath)
	if err != nil {
		fmt.Printf("unable to read bytes from %s %s\n", filepath, err)
		os.Exit(1)
	}
	info, err := compiler.ReadInfoFromBytes(filepath, bytes)
	if err != nil {
		fmt.Printf("unable to read info from %s %s\n", filepath, err)
		os.Exit(1)
	}

	document, err := openapi_v3.NewDocument(info, compiler.NewContext("$root", nil))
	if err != nil {
		fmt.Printf("unable to parse document %s %s\n", filepath, err)
		os.Exit(1)
	}

	w := parser.NewWalker(document)

	err = w.Traverse()
	if err != nil {
		fmt.Printf("unable to traverse models %s %s\n", filepath, err)
		os.Exit(1)
	}

	generator.GenerateFiles(w)

	// opModels := w.GetOperations()

	// spew.Dump(opModels)

	os.Exit(0)
}
