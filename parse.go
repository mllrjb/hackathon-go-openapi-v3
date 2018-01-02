package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/googleapis/gnostic/OpenAPIv3"
	"github.com/googleapis/gnostic/compiler"
)

func main() {
	filepath := "examples/CaseAPI/caseapi.yaml"
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

	generateFiles(document)

	data, err := json.Marshal(document)
	if err != nil {
		fmt.Printf("unable to marshal document to JSON %s %s\n", filepath, err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", data)

	os.Exit(0)
}
