package main

import (
	"encoding/json"
	"fmt"
	"os"

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

	// generateFiles(document)

	w := generator.NewWalker(document)

	err = w.Traverse()
	if err != nil {
		fmt.Printf("unable to traverse models %s\n", filepath, err)
		os.Exit(1)
	}

	schemaModels := w.GetModels()

	for _, value := range schemaModels {
		data, _ := json.Marshal(value)
		fmt.Printf("%s\n", data)
	}

	os.Exit(0)
}
