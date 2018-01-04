package generator

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.schq.secious.com/jason-miller/go-openapi-v3/parser"
)

const templateDir = "./templates"

func GenerateFiles(walker parser.Walker) {
	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		fmt.Printf("unable to read template directory: %v", err)
		os.Exit(1)
	}

	var templateFiles []string
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".tmpl") {
			templateFiles = append(templateFiles, fmt.Sprintf("%v/%v", templateDir, filename))
		}
		//TODO: move .go extension files directly to generated
	}

	funcMap := template.FuncMap{
		"Title": strings.Title,
	}

	t := template.New("template").Funcs(funcMap)
	t, err = t.ParseFiles(templateFiles...)
	if err != nil {
		fmt.Printf("unable to parse template files: %v", err)
		os.Exit(1)
	}

	err = os.MkdirAll("output", os.ModePerm)
	if err != nil {
		fmt.Printf("unable to create output dir: %v", err)
		os.Exit(1)
	}

	operations := walker.GetOperations()

	ctmpl := t.Lookup("operation.tmpl")
	if ctmpl == nil {
		fmt.Println("could not find controller template")
		os.Exit(1)
	}
	for _, op := range operations {
		var buf bytes.Buffer

		genOp := GenerateOperation(op)

		err = ctmpl.Execute(&buf, genOp)
		if err != nil {
			fmt.Printf("error processing operation: %v\n", err)
			os.Exit(1)
		}

		formattedBytes := buf.Bytes()
		// formattedBytes, err := format.Source(buf.Bytes())
		// if err != nil {
		// 	fmt.Println("error formatting operation: %v\n", err)
		// 	os.Exit(1)
		// }

		opPath := fmt.Sprintf("output/%s.go", op.Name)
		opFile, err := os.Create(opPath)
		if err != nil {
			fmt.Printf("unable to create output/%s.go: %v", opPath, err)
			os.Exit(1)
		}

		opFile.Write(formattedBytes)
	}

}
