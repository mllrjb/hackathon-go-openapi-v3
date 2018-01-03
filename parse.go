package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.schq.secious.com/jason-miller/go-openapi-v3/generator"

	"github.com/googleapis/gnostic/OpenAPIv3"
	"github.com/googleapis/gnostic/compiler"
)

const filepath = "examples/CaseAPI/caseapi.yaml"

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

	w := generator.NewWalker(document)

	err = w.Traverse()
	if err != nil {
		fmt.Printf("unable to traverse models %s\n", filepath, err)
		os.Exit(1)
	}

	opModels := w.GetOperations()

	spew.Dump(opModels)

	os.Exit(0)
}
