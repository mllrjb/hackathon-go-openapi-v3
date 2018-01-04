package generator

type GenOperation struct {
	Name     string
	Handlers []GenHandler
	Models   []*GenSchema
	Path     string
	Method   string
}

type GenHandler struct {
	Name   string
	Params string
	Body   *GenSchema
}
