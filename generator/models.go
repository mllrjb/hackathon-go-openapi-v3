package generator

type SchemaModel struct {
	Name       string
	Required   []string
	Properties *map[string]interface{}
}

type Property struct {
	Name string
	Type string
}

type PrimitiveProperty struct {
	Name   string
	Type   string
	Format string
}

type StructProperty struct {
	Name string
	Ref  *SchemaModel `json:"-"`
}

type ResponseModel struct {
	Ref  string
	Name string
}
