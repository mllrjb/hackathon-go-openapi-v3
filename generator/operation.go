package generator

import (
	"fmt"
	"strings"

	"github.schq.secious.com/jason-miller/go-openapi-v3/parser"
	"github.schq.secious.com/jason-miller/go-openapi-v3/utils"
)

func MediaTypeToTitle(mediaType string) string {
	// TODO: handle */* and non application/json
	mediaType = strings.TrimPrefix(mediaType, "application/")
	mediaType = strings.TrimSuffix(mediaType, "json")
	return utils.ToPascalCase(mediaType)
}

func GenerateOperation(op *parser.Operation) GenOperation {
	gOp := GenOperation{
		Name:   op.Name,
		Path:   op.Path,
		Method: op.Method,
	}
	paramsName := fmt.Sprintf("%sParameters", op.Name)
	handlerBase := fmt.Sprintf("%sHandler", op.Name)

	if len(op.Requests) == 0 {
		// generic handler for no request body
		gOp.Handlers = []GenHandler{
			GenHandler{
				Name:   handlerBase,
				Params: paramsName,
			},
		}
	} else {
		for _, r := range op.Requests {
			var handlerBodyName string
			mediaTypeTitle := MediaTypeToTitle(r.Accept)
			handlerName := fmt.Sprintf("%s_%s", handlerBase, mediaTypeTitle)
			if r.IsComponent() {
				// TODO:??
			} else {
				handlerBodyName = fmt.Sprintf("%s%s", op.Name, mediaTypeTitle)

				gs := GenerateSchema(r.Body, handlerBodyName, "operation")
				nested := GetAllNestedModels(&gs)

				// ignore top level slices, since we just use their type directly
				// (should never have a top level primitive for a request either)
				if gs.IsObject {
					gOp.Models = append(gOp.Models, &gs)
				}

				gOp.Models = append(gOp.Models, nested...)

				gOp.Handlers = append(gOp.Handlers, GenHandler{
					Name:   handlerName,
					Params: paramsName,
					Body:   &gs,
				})
			}
		}
	}

	return gOp
}

// operation:

// for each request
// what schema models do i need to instantiate?
// do: schema model

// for each response
// what schema models do i need to instantiate?
// do: schema model

// schema model:

// is it a primitive?
// declare a "type {name} {type}"

// is it an object?
// declare a "type {name} struct"
// recurse properties

// is it an array?
// recurse items
