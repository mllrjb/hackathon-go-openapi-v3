package generator

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/googleapis/gnostic/OpenAPIv3"
)

// 1. walk components/schemas
// 1.1 $ref => lookup/generate intermediate model (store)
// 1.2 walk properties
//  1.2.1 primitive prop => add to schema model
//  1.2.2 $ref prop => lookup/generate intermediate model (store), add to schema model
// 1.3 walk allOf as schema

type Walker struct {
	document   *openapi_v3.Document
	models     map[string]SchemaModel
	operations []*Operation
}

func NewWalker(document *openapi_v3.Document) Walker {
	return Walker{
		document:   document,
		models:     make(map[string]SchemaModel),
		operations: []*Operation{},
	}
}

func (o *Walker) GetModels() []SchemaModel {
	modelList := make([]SchemaModel, len(o.models))

	idx := 0
	for _, v := range o.models {
		modelList[idx] = v
		idx++
	}

	return modelList
}

func (o *Walker) GetOperations() []*Operation {
	return o.operations
}

func (o *Walker) Traverse() error {
	for _, schema := range o.document.Components.Schemas.AdditionalProperties {
		// walk and resolve all refs
		schemaModel, err := o.resolveSchemaOrRef(schema.Value, schema.Name)
		if err != nil {
			return err
		}
		refPath := fmt.Sprintf("#/components/schemas/%s", schema.Name)
		o.models[refPath] = schemaModel
	}

	for _, path := range o.document.Paths.Path {
		operations, err := o.buildOperationsFromPath(path)
		if err != nil {
			return err
		}
		o.operations = append(o.operations, operations...)
	}

	return nil
}

type handlerParams struct {
	path   string
	method string
}

func (o *Walker) buildOperationsFromPath(path *openapi_v3.NamedPathItem) ([]*Operation, error) {
	operations := []*Operation{}

	if path.Value.Get != nil {
		if op, err := o.buildHandlersFromOp(path.Value.Get, handlerParams{
			path:   path.Name,
			method: http.MethodGet,
		}); err == nil {
			operations = append(operations, op)
		} else {
			return nil, err
		}
	}

	if path.Value.Post != nil {
		if op, err := o.buildHandlersFromOp(path.Value.Post, handlerParams{
			path:   path.Name,
			method: http.MethodPost,
		}); err == nil {
			operations = append(operations, op)
		} else {
			return nil, err
		}
	}

	if path.Value.Put != nil {
		if op, err := o.buildHandlersFromOp(path.Value.Put, handlerParams{
			path:   path.Name,
			method: http.MethodPut,
		}); err == nil {
			operations = append(operations, op)
		} else {
			return nil, err
		}
	}

	if path.Value.Patch != nil {
		if op, err := o.buildHandlersFromOp(path.Value.Patch, handlerParams{
			path:   path.Name,
			method: http.MethodPatch,
		}); err == nil {
			operations = append(operations, op)
		} else {
			return nil, err
		}
	}

	if path.Value.Delete != nil {
		if op, err := o.buildHandlersFromOp(path.Value.Delete, handlerParams{
			path:   path.Name,
			method: http.MethodDelete,
		}); err == nil {
			operations = append(operations, op)
		} else {
			return nil, err
		}
	}

	return operations, nil
}

func (o *Walker) buildHandlersFromOp(op *openapi_v3.Operation, params handlerParams) (*Operation, error) {
	operation := Operation{
		Name:   ToPascalCase(op.OperationId),
		Method: params.method,
		Path:   params.path,
	}

	parameters := []Parameter{}

	if op.Parameters != nil {
		for _, param := range op.Parameters {

			// TODO: handle refs
			if param.GetParameter() != nil {
				p := param.GetParameter()
				p2 := Parameter{
					Name:     p.Name,
					In:       p.In,
					Required: p.Required,
				}

				schemaModel, err := o.resolveSchemaOrRef(p.Schema, "")
				if err != nil {
					return nil, err
				}

				if schemaModel.IsObject() {
					return nil, fmt.Errorf("parameter %v should not be an object", p2.Name)
				}

				p2.Schema = schemaModel

				parameters = append(parameters, p2)
			}
		}
	}

	if op.RequestBody != nil {
		// TODO: references
		for _, mediaType := range op.RequestBody.GetRequestBody().Content.AdditionalProperties {
			request := Request{
				Accept: mediaType.Name,
			}

			schemaOrRef := mediaType.Value.Schema
			// schema reference
			if ref := schemaOrRef.GetReference(); ref != nil {
				schemaModel, err := o.resolveSchemaReference(ref)
				if err != nil {
					return nil, err
				}
				request.Body = schemaModel
			} else {
				schemaModel, err := o.resolveSchema(schemaOrRef.GetSchema(), "")
				if err != nil {
					return nil, err
				}
				request.Body = schemaModel
			}

			operation.Requests = append(operation.Requests, request)
		}
	}

	if len(op.Responses.ResponseOrReference) > 0 {
		for _, response := range op.Responses.ResponseOrReference {
			if resp := response.Value.GetResponse(); resp != nil {
				for _, mediaType := range resp.Content.AdditionalProperties {
					// TODO: if mediaTypeName is empty, don't add an extra "_"
					// TODO: try to lookup response "name" from status code (e.g. 200 => OK)
					r := Response{
						StatusCode:  response.Name,
						ContentType: mediaType.Name,
					}

					schemaOrRef := mediaType.Value.Schema
					// schema reference
					if schemaOrRef != nil {
						if ref := schemaOrRef.GetReference(); ref != nil {
							schemaModel, err := o.resolveSchemaReference(ref)
							if err != nil {
								return nil, err
							}
							r.Body = schemaModel
						} else {
							schemaModel, err := o.resolveSchema(schemaOrRef.GetSchema(), "")
							if err != nil {
								return nil, err
							}
							r.Body = schemaModel
						}
					}

					operation.Responses = append(operation.Responses, r)
				}
			}
		}
	}
	return &operation, nil
}

func (o *Walker) resolveSchemaOrRef(schemaOrRef *openapi_v3.SchemaOrReference, componentName string) (SchemaModel, error) {
	// schema reference
	if ref := schemaOrRef.GetReference(); ref != nil {
		return o.resolveSchemaReference(ref)
	}

	// actual schema!
	if schema := schemaOrRef.GetSchema(); schema != nil {
		return o.resolveSchema(schema, componentName)
	}

	return nil, errors.New("not sure what happened...")
}

func (o *Walker) resolveSchema(schema *openapi_v3.Schema, componentName string) (SchemaModel, error) {
	if schema.Type == "object" {
		schemaModel := StructSchemaModel{
			CommonSchemaModel: CommonSchemaModel{
				Component: NewComponent(componentName),
				Title:     schema.Title,
				Type:      schema.Type,
				Nullable:  schema.Nullable,
			},
			Required:   schema.Required,
			Properties: make(map[string]SchemaModel),
		}
		if schema.Properties != nil {
			properties, err := o.buildProperties(schema.Properties)
			if err != nil {
				return nil, err
			}
			schemaModel.Properties = properties
		}

		if schema.AllOf != nil {
			for _, allOf := range schema.AllOf {
				allOfModel, err := o.resolveSchemaOrRef(allOf, "")
				if err != nil {
					return nil, err
				}
				if allOfModel.IsObject() {
					structModel := (allOfModel).(*StructSchemaModel)
					for k, v := range structModel.Properties {
						// TODO: check overrides and warn?
						schemaModel.Properties[k] = v
					}
				}
			}
		}

		return &schemaModel, nil
	}

	if schema.Type == "array" {
		itemModel, err := o.resolveSchemaOrRef(schema.Items.SchemaOrReference[0], "")
		if err != nil {
			return nil, err
		}
		schemaModel := ArraySchemaModel{
			CommonSchemaModel: CommonSchemaModel{
				Component: NewComponent(componentName),
				Title:     schema.Title,
				Type:      schema.Type,
				Nullable:  schema.Nullable,
			},
			Items:    itemModel,
			MinItems: schema.MinItems,
			MaxItems: schema.MaxItems,
		}

		return &schemaModel, nil
	}

	schemaModel := PrimitiveSchemaModel{
		CommonSchemaModel: CommonSchemaModel{
			Component: NewComponent(componentName),
			Title:     schema.Title,
			Type:      schema.Type,
			Nullable:  schema.Nullable,
		},
		Format:    schema.Format,
		MinLength: schema.MinLength,
		MaxLength: schema.MaxLength,
	}

	// TODO: support for oneOf has polymorphic implications
	// TODO: support for anyOf is like oneOf, but potentially w/o a discriminator, first match
	return &schemaModel, nil
}

// TODO: for $ref that points to a simple object, just copy it (don't have a pointer reference)
func (o *Walker) buildProperties(props *openapi_v3.Properties) (map[string]SchemaModel, error) {
	properties := make(map[string]SchemaModel)
	for _, prop := range props.AdditionalProperties {
		model, err := o.resolveSchemaOrRef(prop.Value, "")
		if err != nil {
			return nil, err
		}
		properties[prop.Name] = model
	}
	return properties, nil
}

func (o *Walker) resolveSchemaReference(ref *openapi_v3.Reference) (SchemaModel, error) {
	existingModel, ok := o.models[ref.XRef]
	if ok {
		return existingModel, nil
	}

	for _, schema := range o.document.Components.Schemas.AdditionalProperties {
		// TODO: sub-refs? (e.g. #/components/schema/MyModel/properties/FooBar)
		refPath := fmt.Sprintf("#/components/schemas/%s", schema.Name)
		if strings.EqualFold(ref.XRef, refPath) {
			model, err := o.resolveSchemaOrRef(schema.Value, schema.Name)
			if err != nil {
				return nil, err
			}

			// TODO: is this safe? what if we already wrote it?
			o.models[ref.XRef] = model

			return model, nil
		}
	}
	return nil, fmt.Errorf("could not resolve $ref: '%v'", ref.XRef)
}
