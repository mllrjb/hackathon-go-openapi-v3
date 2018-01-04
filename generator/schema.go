package generator

import (
	"fmt"

	"github.schq.secious.com/jason-miller/go-openapi-v3/parser"
	"github.schq.secious.com/jason-miller/go-openapi-v3/utils"
)

type GenSchema struct {
	ReceiverName       string
	IsDefinedElsewhere bool
	IsPrimitive        bool
	IsObject           bool
	IsSlice            bool
	GoType             string
	Properties         []*GenSchema
	Items              *GenSchema
}

func GenerateSchema(m parser.SchemaModel, receiverName string) GenSchema {
	if m.IsPrimitive() {
		p := m.(*parser.PrimitiveSchemaModel)

		if p.IsComponent() {
			return GenSchema{
				ReceiverName:       receiverName,
				IsDefinedElsewhere: true,
				GoType:             p.GetComponentName(),
			}
		}
		// generate "type {name} {type}"
		return GenSchema{
			ReceiverName:       receiverName,
			IsDefinedElsewhere: false,
			IsPrimitive:        true,
			// TODO: map types (via type + format?)
			GoType: p.Type,
		}
	}
	if m.IsObject() {
		p := m.(*parser.StructSchemaModel)

		if p.IsComponent() {
			return GenSchema{
				ReceiverName:       receiverName,
				IsDefinedElsewhere: true,
				GoType:             p.GetComponentName(),
			}
		}
		// generate "type {name} struct"
		gs := GenSchema{
			ReceiverName:       receiverName,
			IsDefinedElsewhere: false,
			IsObject:           true,
			GoType:             receiverName,
		}

		for propName, prop := range p.Properties {
			gsp := GenerateSchema(prop, propName)
			gs.Properties = append(gs.Properties, &gsp)
		}

		return gs
	}
	if m.IsArray() {
		p := m.(*parser.ArraySchemaModel)

		if p.IsComponent() {
			return GenSchema{
				ReceiverName:       receiverName,
				IsDefinedElsewhere: true,
				GoType:             p.GetComponentName(),
			}
		} else if p.Items.IsComponent() {
			gsi := GenSchema{
				ReceiverName:       receiverName,
				IsDefinedElsewhere: true,
				IsSlice:            false,
				GoType:             p.Items.GetComponentName(),
			}
			// generate "type {name} []{item.name}"
			return GenSchema{
				ReceiverName:       receiverName,
				IsDefinedElsewhere: false,
				IsSlice:            true,
				GoType:             fmt.Sprintf("[]%s", gsi.GoType),
				Items:              &gsi,
			}
		} else {
			itemReceiverName := fmt.Sprintf("%s%s", receiverName, utils.ToPascalCase(p.Items.GetType()))
			gsi := GenerateSchema(p.Items, itemReceiverName)
			// generate "type {name}Slice []{name}{item.type}"
			gs := GenSchema{
				ReceiverName:       receiverName,
				IsDefinedElsewhere: false,
				IsSlice:            true,
				GoType:             fmt.Sprintf("[]%s", gsi.GoType),
				Items:              &gsi,
			}

			return gs
		}
	}

	return GenSchema{}
}

func GenerateSchemaComponents(m parser.SchemaModel) GenSchema {
	if m.IsPrimitive() {
		p := m.(*parser.PrimitiveSchemaModel)

		// generate "type {name} {type}"
		return GenSchema{
			ReceiverName:       p.GetComponentName(),
			IsDefinedElsewhere: p.IsComponent(),
			IsPrimitive:        true,
			// TODO: map types (via type + format?)
			GoType: p.Type,
		}
	}
	if m.IsObject() {
		p := m.(*parser.StructSchemaModel)

		// generate "type {name} struct"
		gs := GenSchema{
			ReceiverName:       p.GetComponentName(),
			IsDefinedElsewhere: p.IsComponent(),
			IsObject:           true,
			GoType:             p.GetComponentName(),
		}

		for propName, prop := range p.Properties {
			gsp := GenerateSchema(prop, propName)
			gs.Properties = append(gs.Properties, &gsp)
		}

		return gs
	}
	if m.IsArray() {
		p := m.(*parser.ArraySchemaModel)

		if p.Items.IsComponent() {
			gsi := GenSchema{
				ReceiverName:       p.GetComponentName(),
				IsDefinedElsewhere: p.Items.IsComponent(),
				IsSlice:            false,
				GoType:             p.Items.GetComponentName(),
			}
			// generate "type {name} []{item.name}"
			return GenSchema{
				ReceiverName:       p.GetComponentName(),
				IsDefinedElsewhere: p.IsComponent(),
				IsSlice:            true,
				GoType:             fmt.Sprintf("[]%s", gsi.GoType),
				Items:              &gsi,
			}
		}
		itemReceiverName := fmt.Sprintf("%s%s", p.GetComponentName(), utils.ToPascalCase(p.Items.GetType()))
		gsi := GenerateSchema(p.Items, itemReceiverName)
		// generate "type {name}Slice []{name}{item.type}"
		gs := GenSchema{
			ReceiverName:       p.GetComponentName(),
			IsDefinedElsewhere: p.IsComponent(),
			IsSlice:            true,
			GoType:             fmt.Sprintf("[]%s", gsi.GoType),
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
