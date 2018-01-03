package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.com/googleapis/gnostic/OpenAPIv3"
)

const templateDir = "./templates"

func generateFiles(document *openapi_v3.Document) {
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

	controllerFile, err := os.Create("output/controllers.go")
	if err != nil {
		fmt.Printf("unable to create output/controllers.go: %v", err)
		os.Exit(1)
	}

	paths := document.GetPaths()
	if paths != nil {
		t.ExecuteTemplate(controllerFile, "controller.tmpl", *paths)
	}

}
