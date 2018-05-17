package generator

import (
	"fmt"

	"github.com/mllrjb/hackathon-go-openapi-v3/parser"
	"github.com/mllrjb/hackathon-go-openapi-v3/utils"
)

type NestedItems []struct {
	name string
}

type resolvedType struct {
	// e.g. "components"
	Pkg string

	// e.g. string, struct, slice
	GoType string

	// e.g. MyObject
	ReferenceType string
}

type GenSchema struct {
	resolvedType
	ReceiverName       string
	IsDefinedElsewhere bool
	IsPrimitive        bool
	IsObject           bool
	IsSlice            bool
	Properties         []*GenSchema
	Items              *GenSchema
}

// TODO: map primitive types?
func getPrimitiveType(t string, format string) string {
	switch t {
	case "integer":
		if len(format) > 0 {
			return format
		}
		return "int64"
	}
	return t
}

func getResolvedType(m parser.SchemaModel, pkg string) resolvedType {
	if m.IsComponent() {
		if m.IsPrimitive() {
			p := m.(*parser.PrimitiveSchemaModel)
			return resolvedType{
				Pkg:           "component",
				GoType:        getPrimitiveType(m.GetType(), p.Format),
				ReferenceType: m.GetComponentName(),
			}
		}

		if m.IsArray() {
			return resolvedType{
				Pkg:           "component",
				GoType:        "slice",
				ReferenceType: m.GetComponentName(),
			}
		}

		if m.IsObject() {
			return resolvedType{
				Pkg:           "component",
				GoType:        "struct",
				ReferenceType: m.GetComponentName(),
			}
		}
	}

	if m.IsPrimitive() {
		p := m.(*parser.PrimitiveSchemaModel)
		return resolvedType{
			Pkg:           pkg,
			GoType:        getPrimitiveType(m.GetType(), p.Format),
			ReferenceType: "",
		}
	}

	if m.IsArray() {
		return resolvedType{
			Pkg:           pkg,
			GoType:        "slice",
			ReferenceType: "",
		}
	}

	if m.IsObject() {
		return resolvedType{
			Pkg:           pkg,
			GoType:        "struct",
			ReferenceType: "",
		}
	}

	return resolvedType{}
}

func GenerateSchema(m parser.SchemaModel, receiverName string, pkg string) GenSchema {
	if m.IsDiscriminated() {
		fmt.Sprintf("discriminated: %s\v", m.GetComponentName())
	}

	resolvedType := getResolvedType(m, pkg)
	if m.IsPrimitive() {
		p := m.(*parser.PrimitiveSchemaModel)

		if p.IsComponent() {
			return GenSchema{
				resolvedType:       resolvedType,
				ReceiverName:       receiverName,
				IsDefinedElsewhere: true,
			}
		}
		// generate "type {name} {type}"
		return GenSchema{
			resolvedType:       resolvedType,
			ReceiverName:       receiverName,
			IsDefinedElsewhere: false,
			IsPrimitive:        true,
		}
	}
	if m.IsObject() {
		p := m.(*parser.StructSchemaModel)

		if p.IsComponent() {
			return GenSchema{
				resolvedType:       resolvedType,
				ReceiverName:       receiverName,
				IsDefinedElsewhere: true,
			}
		}
		// generate "type {name} struct"
		// eventually, we have to generate a type to refer to
		resolvedType.ReferenceType = receiverName
		gs := GenSchema{
			resolvedType:       resolvedType,
			ReceiverName:       receiverName,
			IsDefinedElsewhere: false,
			IsObject:           true,
		}

		for propName, prop := range p.Properties {
			gsp := GenerateSchema(prop, propName, pkg)
			gs.Properties = append(gs.Properties, &gsp)
		}

		return gs
	}
	if m.IsArray() {
		p := m.(*parser.ArraySchemaModel)

		if p.IsComponent() {
			return GenSchema{
				resolvedType:       resolvedType,
				ReceiverName:       receiverName,
				IsDefinedElsewhere: true,
			}
		} else if p.Items.IsComponent() {
			gsi := GenSchema{
				resolvedType:       getResolvedType(p.Items, pkg),
				ReceiverName:       receiverName,
				IsDefinedElsewhere: true,
				IsSlice:            false,
			}
			// generate "type {name} []{item.name}"
			return GenSchema{
				resolvedType:       resolvedType,
				ReceiverName:       receiverName,
				IsDefinedElsewhere: false,
				IsSlice:            true,
				Items:              &gsi,
			}
		} else {
			itemReceiverName := fmt.Sprintf("%s%s", receiverName, utils.ToPascalCase(p.Items.GetType()))
			gsi := GenerateSchema(p.Items, itemReceiverName, pkg)
			// generate "type {name}Slice []{name}{item.type}"
			gs := GenSchema{
				resolvedType:       resolvedType,
				ReceiverName:       receiverName,
				IsDefinedElsewhere: false,
				IsSlice:            true,
				Items:              &gsi,
			}

			return gs
		}
	}

	return GenSchema{}
}

func GenerateSchemaComponents(m parser.SchemaModel) GenSchema {
	if m.IsDiscriminated() {
		fmt.Printf("discriminated: %s\n", m.GetComponentName())
	}
	resolvedType := getResolvedType(m, "component")
	if m.IsPrimitive() {
		p := m.(*parser.PrimitiveSchemaModel)

		// generate "type {name} {type}"
		//
		return GenSchema{
			resolvedType:       resolvedType,
			ReceiverName:       p.GetComponentName(),
			IsDefinedElsewhere: p.IsComponent(),
			IsPrimitive:        true,
		}
	}
	if m.IsObject() {
		p := m.(*parser.StructSchemaModel)

		// generate "type {name} struct"
		gs := GenSchema{
			resolvedType:       resolvedType,
			ReceiverName:       p.GetComponentName(),
			IsDefinedElsewhere: p.IsComponent(),
			IsObject:           true,
		}

		for propName, prop := range p.Properties {
			gsp := GenerateSchema(prop, propName, "component")
			gs.Properties = append(gs.Properties, &gsp)
		}

		return gs
	}
	if m.IsArray() {
		p := m.(*parser.ArraySchemaModel)

		if p.Items.IsComponent() {
			gsi := GenSchema{
				resolvedType:       getResolvedType(p.Items, "component"),
				ReceiverName:       p.GetComponentName(),
				IsDefinedElsewhere: p.Items.IsComponent(),
				IsSlice:            false,
			}
			// generate "type {name} []{item.name}"
			return GenSchema{
				resolvedType:       resolvedType,
				ReceiverName:       p.GetComponentName(),
				IsDefinedElsewhere: p.IsComponent(),
				IsSlice:            true,
				Items:              &gsi,
			}
		}
		itemReceiverName := fmt.Sprintf("%s%s", p.GetComponentName(), utils.ToPascalCase(p.Items.GetType()))
		gsi := GenerateSchema(p.Items, itemReceiverName, "component")
		// generate "type {name}Slice []{name}{item.type}"
		gs := GenSchema{
			resolvedType:       resolvedType,
			ReceiverName:       p.GetComponentName(),
			IsDefinedElsewhere: p.IsComponent(),
			IsSlice:            true,
			Items:              &gsi,
		}

		return gs
	}

	return GenSchema{}
}

func GetAllNestedModels(gs *GenSchema) []*GenSchema {
	// TODO: not accurate
	if gs.IsObject {
		return []*GenSchema{}
	}

	if gs.IsPrimitive {
		return []*GenSchema{}
	}

	if gs.IsSlice {
		if gs.Items.IsPrimitive {
			return []*GenSchema{}
		}

		if gs.Items.IsObject {
			if !gs.Items.IsDefinedElsewhere {
				return []*GenSchema{gs.Items}
			}
		}

		if gs.Items.IsSlice {
			return GetAllNestedModels(gs.Items)
		}
	}

	return []*GenSchema{}
}
