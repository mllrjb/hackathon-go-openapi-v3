package generator

import (
	"errors"
	"fmt"
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
	document *openapi_v3.Document
	models   map[string]*SchemaModel
}

func NewWalker(document *openapi_v3.Document) Walker {
	return Walker{
		document: document,
		models:   make(map[string]*SchemaModel),
	}
}

func (o *Walker) GetModels() []SchemaModel {
	modelList := make([]SchemaModel, len(o.models))

	idx := 0
	for _, v := range o.models {
		modelList[idx] = *v
		idx++
	}

	return modelList
}

func (o *Walker) Traverse() error {
	for _, schema := range o.document.Components.Schemas.AdditionalProperties {
		// walk and resolve all refs
		schemaModel, err := o.resolveSchemaOrRef(schema.Value)
		if err != nil {
			return err
		}
		o.models[schema.Name] = schemaModel

		// if a name was not already set, use the 'schemas' key as the name
		if len(schemaModel.Name) == 0 {
			schemaModel.Name = schema.Name
		}
	}
	return nil
}

func (o *Walker) resolveSchemaOrRef(schemaOrRef *openapi_v3.SchemaOrReference) (*SchemaModel, error) {
	// schema reference
	if ref := schemaOrRef.GetReference(); ref != nil {
		return o.resolveSchemaReference(ref)
	}

	// actual schema!
	if schema := schemaOrRef.GetSchema(); schema != nil {
		return o.resolveSchema(schema)
	}

	return nil, errors.New("not sure what happened...")
}

func (o *Walker) resolveSchema(schema *openapi_v3.Schema) (*SchemaModel, error) {
	schemaModel := SchemaModel{
		Name:     schema.Title,
		Required: []string{},
	}

	if schema.Properties != nil {
		properties, err := o.buildProperties(schema.Properties)
		if err != nil {
			return nil, err
		}
		schemaModel.Properties = &properties
	} else {
		properties := make(map[string]interface{})
		schemaModel.Properties = &properties
	}

	if schema.AllOf != nil {
		for _, allOf := range schema.AllOf {
			allOfModel, err := o.resolveSchemaOrRef(allOf)
			if err != nil {
				return nil, err
			}
			if allOfModel == nil {
			}
			if len(*allOfModel.Properties) > 0 {
				for k, v := range *allOfModel.Properties {
					// TODO: check overrides and warn?
					(*schemaModel.Properties)[k] = v
				}
			}
		}
	}

	// TODO: support for oneOf has polymorphic implications
	// TODO: support for anyOf is like oneOf, but potentially w/o a discriminator, first match
	return &schemaModel, nil
}

func (o *Walker) buildProperties(props *openapi_v3.Properties) (map[string]interface{}, error) {
	properties := make(map[string]interface{})
	for _, prop := range props.AdditionalProperties {
		if ref := prop.Value.GetReference(); ref != nil {
			model, err := o.resolveSchemaReference(ref)
			if err != nil {
				return nil, err
			}
			properties[prop.Name] = StructProperty{
				Name: prop.Name,
				Ref:  model,
			}
		}

		if schema := prop.Value.GetSchema(); schema != nil {
			switch schema.Type {
			case "object":
			default:
				properties[prop.Name] = PrimitiveProperty{
					Name:   prop.Name,
					Type:   schema.Type,
					Format: schema.Format,
				}
			}
		}
	}
	return properties, nil
}

func (o *Walker) resolveSchemaReference(ref *openapi_v3.Reference) (*SchemaModel, error) {
	existingModel, ok := o.models[ref.XRef]
	if ok {
		return existingModel, nil
	}

	// TODO: better way to resolve refs? (this is pretty strict)
	name := strings.TrimPrefix(ref.XRef, "#/components/schemas/")

	for _, schema := range o.document.Components.Schemas.AdditionalProperties {
		if strings.EqualFold(schema.Name, name) {
			model, err := o.resolveSchemaOrRef(schema.Value)
			if err != nil {
				return nil, err
			}

			// TODO: is this safe? what if we already wrote it?
			o.models[ref.XRef] = model

			// set a name if one was not set (i.e. by title)
			if len(model.Name) == 0 {
				model.Name = name
			}
			return model, nil
		}
	}
	return nil, fmt.Errorf("could not resolve $ref: '%v'", ref.XRef)
}
