package parser

type Component struct {
	componentName string
}

func NewComponent(componentName string) Component {
	return Component{
		componentName,
	}
}

func (m *Component) IsComponent() bool {
	return len(m.componentName) > 0
}

func (m *Component) GetComponentName() string {
	return m.componentName
}

type SchemaModel interface {
	IsComponent() bool
	GetComponentName() string
	GetType() string
	IsPrimitive() bool
	IsArray() bool
	IsObject() bool
	IsDiscriminated() bool
	GetDiscriminator() *DiscriminatedSchemaModel
}

type Discriminator struct {
	PropertyName string
	Mapping      map[string]SchemaModel
}

type CommonSchemaModel struct {
	Component
	Title    string
	Type     string
	Nullable bool
}

func (m *CommonSchemaModel) GetType() string {
	return m.Type
}

func (m *CommonSchemaModel) IsPrimitive() bool {
	return m.Type != "object" && m.Type != "array"
}

func (m *CommonSchemaModel) IsObject() bool {
	return m.Type == "object"
}

func (m *CommonSchemaModel) IsArray() bool {
	return m.Type == "array"
}

type DiscriminatedSchemaModel struct {
	DiscriminatorType    string // oneOf | anyOf
	DiscriminatorSchemas []SchemaModel
	Discriminator        *Discriminator
}

func (m *DiscriminatedSchemaModel) IsDiscriminated() bool {
	return len(m.DiscriminatorType) > 0
}

func (m *DiscriminatedSchemaModel) GetDiscriminator() *DiscriminatedSchemaModel {
	return m
}

type StructSchemaModel struct {
	CommonSchemaModel
	DiscriminatedSchemaModel
	Properties map[string]SchemaModel
	Required   []string
}

type ArraySchemaModel struct {
	CommonSchemaModel
	Items    SchemaModel
	MinItems int64
	MaxItems int64
}

func (m *ArraySchemaModel) IsDiscriminated() bool {
	return false
}

func (m *ArraySchemaModel) GetDiscriminator() *DiscriminatedSchemaModel {
	return nil
}

type PrimitiveSchemaModel struct {
	CommonSchemaModel
	DiscriminatedSchemaModel
	Format    string
	MinLength int64
	MaxLength int64
}

type Operation struct {
	Name       string
	Requests   []Request
	Method     string
	Path       string
	Responses  []Response
	Parameters []Parameter
}

type Request struct {
	Component
	Accept string
	Body   SchemaModel
}

type Parameter struct {
	Component
	Name     string
	In       string
	Required bool
	Schema   SchemaModel
}

type Response struct {
	Component
	StatusCode  string
	ContentType string
	Body        SchemaModel
	Headers     map[string]SchemaModel
}
