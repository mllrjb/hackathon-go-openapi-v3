package parser

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/googleapis/gnostic/OpenAPIv3"

	"github.schq.secious.com/jason-miller/go-openapi-v3/utils"
)

// 1. walk components/schemas
// 1.1 $ref => lookup/generate intermediate model (store)
// 1.2 walk properties
//  1.2.1 primitive prop => add to schema model
//  1.2.2 $ref prop => lookup/generate intermediate model (store), add to schema model
// 1.3 walk allOf as schema

type Walker struct {
	document   *openapi_v3.Document
	models     []SchemaModel
	operations []*Operation
}

func NewWalker(document *openapi_v3.Document) Walker {
	return Walker{
		document:   document,
		models:     []SchemaModel{},
		operations: []*Operation{},
	}
}

func (o *Walker) GetModels() []SchemaModel {
	return o.models
}

func (o *Walker) AddModel(schemaModel SchemaModel) {
	existing := o.FindModel(schemaModel.GetComponentName())
	if existing == nil {
		o.models = append(o.models, schemaModel)
	}
}

func (o *Walker) FindModel(refOrName string) SchemaModel {
	for _, m := range o.models {
		if m.GetComponentName() == refOrName || componentSchemaPath(m.GetComponentName()) == refOrName {
			return m
		}
	}
	return nil
}

func componentSchemaPath(name string) string {
	return fmt.Sprintf("#/components/schemas/%s", name)
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
		o.AddModel(schemaModel)
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
		Name:   utils.ToPascalCase(op.OperationId),
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
				if resp.Content == nil {
					operation.Responses = append(operation.Responses, Response{
						StatusCode: response.Name,
					})
				} else {
					for _, mediaType := range resp.Content.AdditionalProperties {
						r := Response{
							StatusCode:  response.Name,
							ContentType: mediaType.Name,
						}
						// TODO: if mediaTypeName is empty, don't add an extra "_"
						// TODO: try to lookup response "name" from status code (e.g. 200 => OK)

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
						schemaModel.Properties[utils.ToPascalCase(k)] = v
					}
				}
			}
		}

		if schema.AnyOf != nil {
			schemaModel.DiscriminatorType = "anyOf"
			err := o.discriminatedSchema(&schemaModel.DiscriminatedSchemaModel, schema, schema.AnyOf)
			if err != nil {
				return nil, err
			}
		} else if schema.OneOf != nil {
			schemaModel.DiscriminatorType = "oneOf"
			err := o.discriminatedSchema(&schemaModel.DiscriminatedSchemaModel, schema, schema.OneOf)
			if err != nil {
				return nil, err
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

	// TODO: not valid with other attributes (like properties, type)
	// e.g. you can't express that something is a type: number AND must be oneOf: string, bool
	// (technically, allOf can produce similarly invalid results, which is why we only validate it for object)
	// 'Discriminated' should probably be its own type (that generates an interface, not an object)
	// Question: does this "jive" w/ the shorthand for a discriminated subtype? idk how we would implement that anyways... seems crazy
	if schema.AnyOf != nil {
		schemaModel.DiscriminatorType = "anyOf"
		err := o.discriminatedSchema(&schemaModel.DiscriminatedSchemaModel, schema, schema.AnyOf)
		if err != nil {
			return nil, err
		}
	} else if schema.OneOf != nil {
		schemaModel.DiscriminatorType = "oneOf"
		err := o.discriminatedSchema(&schemaModel.DiscriminatedSchemaModel, schema, schema.OneOf)
		if err != nil {
			return nil, err
		}
	}

	return &schemaModel, nil
}

func (o *Walker) discriminatedSchema(schemaModel *DiscriminatedSchemaModel, schema *openapi_v3.Schema, dSchemas []*openapi_v3.SchemaOrReference) error {
	if schema.Discriminator != nil {
		// TODO: mapping of types
		d := Discriminator{
			PropertyName: schema.Discriminator.PropertyName,
			Mapping:      make(map[string]SchemaModel),
		}

		if schema.Discriminator.Mapping != nil {
			for _, mapping := range schema.Discriminator.Mapping.AdditionalProperties {
				ref := openapi_v3.Reference{
					XRef: mapping.Value,
				}
				mapModel, err := o.resolveSchemaReference(&ref)
				if err != nil {
					return err
				}

				d.Mapping[mapping.Name] = mapModel
			}
		}
		schemaModel.Discriminator = &d
	}

	for _, ds := range dSchemas {
		dModel, err := o.resolveSchemaOrRef(ds, "")
		if err != nil {
			return err
		}

		schemaModel.DiscriminatorSchemas = append(schemaModel.DiscriminatorSchemas, dModel)

		// TODO: how do inline models work?
		if dModel.IsComponent() && schemaModel.Discriminator != nil {
			var existingMapping SchemaModel
			for _, m := range schemaModel.Discriminator.Mapping {
				if m == dModel {
					existingMapping = m
					break
				}
			}
			if existingMapping == nil {
				schemaModel.Discriminator.Mapping[dModel.GetComponentName()] = dModel
			}
		}
	}
	return nil
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
	existingModel := o.FindModel(ref.XRef)
	if existingModel != nil {
		return existingModel, nil
	}

	for _, schema := range o.document.Components.Schemas.AdditionalProperties {
		// TODO: sub-refs? (e.g. #/components/schema/MyModel/properties/FooBar)
		refPath := componentSchemaPath(schema.Name)
		if strings.EqualFold(ref.XRef, refPath) {
			schemaModel, err := o.resolveSchemaOrRef(schema.Value, schema.Name)
			if err != nil {
				return nil, err
			}

			// TODO: is this safe? what if we already wrote it? should we re-use a ref?
			o.AddModel(schemaModel)

			return schemaModel, nil
		}
	}
	return nil, fmt.Errorf("could not resolve $ref: '%v'", ref.XRef)
}
