package generator

import "github.schq.secious.com/jason-miller/go-openapi-v3/parser"
import "fmt"
import "github.schq.secious.com/jason-miller/go-openapi-v3/utils"

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
				IsSlice:            true,
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
