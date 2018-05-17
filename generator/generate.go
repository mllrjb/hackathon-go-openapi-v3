package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/mllrjb/hackathon-go-openapi-v3/utils"

	"github.com/mllrjb/hackathon-go-openapi-v3/parser"
)

const templateDir = "./templates"
const outputDir = "generated"
const formatSource = false

func ref(gs *GenSchema, currentPackage string) string {
	if gs.IsDefinedElsewhere {
		if currentPackage == gs.Pkg {
			return gs.ReferenceType
		}
		return fmt.Sprintf("%s.%s", gs.Pkg, gs.ReferenceType)
	}

	if gs.IsPrimitive {
		return gs.GoType
	}

	if gs.IsSlice {
		return fmt.Sprintf("[]%s", ref(gs.Items, currentPackage))
	}

	if gs.IsObject {
		if currentPackage == gs.Pkg {
			return gs.ReferenceType
		}
		return fmt.Sprintf("%s.%s", gs.Pkg, gs.ReferenceType)
	}

	return "UNKNOWN_REF_TYPE"
}

func GenerateFiles(walker parser.Walker) {
	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		fmt.Printf("unable to read template directory: %v", err)
		os.Exit(1)
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Printf("unable to create output dir: %v", err)
		os.Exit(1)
	}

	var templateFiles []string
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".tmpl") {
			templateFiles = append(templateFiles, fmt.Sprintf("%v/%v", templateDir, filename))
		} else if strings.HasSuffix(filename, ".go") {
			err = copyFile(fmt.Sprintf("%s/%s", templateDir, filename), fmt.Sprintf("%s/%s", outputDir, filename))
			if err != nil {
				fmt.Printf("error copying .go file: %v", err)
				os.Exit(1)
			}
			fmt.Printf("copied file %v\n", filename)
		}
	}

	funcMap := template.FuncMap{
		"Title":  strings.Title,
		"pascal": utils.ToPascalCase,
		"ref":    ref,
	}

	t := template.New("template").Funcs(funcMap)
	t, err = t.ParseFiles(templateFiles...)
	if err != nil {
		fmt.Printf("unable to parse template files: %v", err)
		os.Exit(1)
	}

	genOps := []*GenOperation{}
	for _, op := range walker.GetOperations() {
		genOp := GenerateOperation(op)
		genOps = append(genOps, &genOp)
	}

	genSchemas := []*GenSchema{}
	for _, schema := range walker.GetModels() {
		gs := GenerateSchemaComponents(schema)

		genSchemas = append(genSchemas, &gs)

		nested := GetAllNestedModels(&gs)
		genSchemas = append(genSchemas, nested...)
	}

	generateOperations(t, genOps)
	generateComponents(t, genSchemas)
	generatePaths(t, genOps)

}

func writeFile(filepath string, bytes []byte) error {
	if formatSource {
		formattedBytes, err := format.Source(bytes)
		if err != nil {
			fmt.Printf("warning: unable to format output for %s: %v\n", filepath, err)
		}
		bytes = formattedBytes
	}

	filedir := path.Dir(filepath)
	err := os.MkdirAll(filedir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create output dir: %v", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("unable to create %s: %v", filepath, err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("unable to write to %s: %v", filepath, err)
	}

	return nil
}

func generateOperations(tmpl *template.Template, genOps []*GenOperation) {
	otmpl := tmpl.Lookup("operation.tmpl")
	if otmpl == nil {
		fmt.Println("could not find operation template")
		os.Exit(1)
	}

	for _, genOp := range genOps {
		var buf bytes.Buffer
		err := otmpl.Execute(&buf, genOp)
		if err != nil {
			fmt.Printf("error processing operation: %v\n", err)
			os.Exit(1)
		}

		err = writeFile(fmt.Sprintf("%s/operation/%s.go", outputDir, genOp.Name), buf.Bytes())
		if err != nil {
			fmt.Printf("unable to write operation: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("wrote file to %v for op %v\n", genOp, genOp.Name)
	}
}

func generateComponents(tmpl *template.Template, genSchemas []*GenSchema) {
	ctmpl := tmpl.Lookup("components.tmpl")
	if ctmpl == nil {
		fmt.Println("could not find components template")
		os.Exit(1)
	}

	for _, model := range genSchemas {
		var buf bytes.Buffer
		err := ctmpl.Execute(&buf, model)
		if err != nil {
			fmt.Printf("error processing component: %v\n", err)
			os.Exit(1)
		}

		err = writeFile(fmt.Sprintf("%s/component/%s.go", outputDir, model.ReceiverName), buf.Bytes())
		if err != nil {
			fmt.Printf("unable to write component: %v\n", err)
			os.Exit(1)
		}
	}
}

func generatePaths(tmpl *template.Template, genOps []*GenOperation) {
	ptmpl := tmpl.Lookup("pathRouting.tmpl")
	if ptmpl == nil {
		fmt.Println("could not find operation template")
		os.Exit(1)
	}

	var buf bytes.Buffer
	err := ptmpl.Execute(&buf, genOps)
	if err != nil {
		fmt.Printf("error processing operation: %v\n", err)
		os.Exit(1)
	}

	// fmt.Println(buf.String())
	formattedBytes := buf.Bytes()
	// formattedBytes, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Printf("error formatting pathRouting: %v\n", err)
		os.Exit(1)
	}

	opPath := fmt.Sprintf("%s/pathRouting.go", outputDir)
	opFile, err := os.Create(opPath)
	if err != nil {
		fmt.Printf("unable to create %s: %v", opPath, err)
		os.Exit(1)
	}

	_, err = opFile.Write(formattedBytes)
	if err != nil {
		fmt.Printf("unable to write to %s: %v", opPath, err)
		os.Exit(1)
	}
}

func copyFile(src string, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
