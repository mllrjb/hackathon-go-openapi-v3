package generator

type GenOperation struct {
	Name     string
	Handlers []GenHandler
	Models   []*GenSchema
}

type GenHandler struct {
	Name   string
	Params string
	Body   *GenSchema
}
