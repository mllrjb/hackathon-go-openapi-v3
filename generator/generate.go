package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.schq.secious.com/jason-miller/go-openapi-v3/parser"
	"github.schq.secious.com/jason-miller/go-openapi-v3/utils"
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

		genOp := ConvertOperation(op)

		err = ctmpl.Execute(&buf, genOp)
		if err != nil {
			fmt.Printf("error processing operation: %v\n", err)
			os.Exit(1)
		}

		// formattedBytes := buf.Bytes()
		formattedBytes, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Println("error formatting operation: %v\n", err)
			os.Exit(1)
		}

		opPath := fmt.Sprintf("output/%s.go", op.Name)
		opFile, err := os.Create(opPath)
		if err != nil {
			fmt.Printf("unable to create output/%s.go: %v", opPath, err)
			os.Exit(1)
		}

		opFile.Write(formattedBytes)
	}

}

func MediaTypeToTitle(mediaType string) string {
	// TODO: handle */* and non application/json
	mediaType = strings.TrimPrefix(mediaType, "application/")
	mediaType = strings.TrimSuffix(mediaType, "json")
	return utils.ToPascalCase(mediaType)
}

func ConvertOperation(op *parser.Operation) Operation {
	gOp := Operation{
		Name: op.Name,
	}
	paramsName := fmt.Sprintf("%sParameters", op.Name)

	if len(op.Requests) == 0 {
		gOp.Handlers = []Handler{
			Handler{
				Name:   op.Name,
				Params: paramsName,
			},
		}
	} else {
		for _, r := range op.Requests {
			var handlerBodyName string
			mediaTypeTitle := MediaTypeToTitle(r.Accept)
			handlerName := fmt.Sprintf("%sHandler_%s", op.Name, mediaTypeTitle)
			if r.IsComponent() {
				if r.Body.IsComponent() {
					handlerBodyName = r.Body.GetComponentName()
				} else {
					handlerBodyName = fmt.Sprintf("%sBody", r.GetComponentName())
				}
			} else {
				handlerBodyName = fmt.Sprintf("%s%sBody", op.Name, mediaTypeTitle)

				if r.Body.IsObject() {
					model := Model{
						Name: handlerBodyName,
						// TODO: slice (primitive not valid)
						Type:       "struct",
						Properties: make(map[string]string),
					}

					structModel := r.Body.(*parser.StructSchemaModel)
					for k, p := range structModel.Properties {
						if p.IsPrimitive() {
							p2 := p.(*parser.PrimitiveSchemaModel)

							// TODO: is a component?
							model.Properties[k] = p2.GetType()
							// } else if p.IsArray() {
							// 	p2 := p.(*parser.ArraySchemaModel)
						} else {
							p2 := p.(*parser.StructSchemaModel)
							model.Properties[k] = p2.GetComponentName()
						}
					}
					gOp.Models = append(gOp.Models, model)
				}
			}

			handler := Handler{
				Name:   handlerName,
				Params: paramsName,
				Body:   handlerBodyName,
			}

			gOp.Handlers = append(gOp.Handlers, handler)
		}
	}

	return gOp
}
