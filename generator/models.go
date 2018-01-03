package generator

type Operation struct {
	Name     string
	Handlers []Handler
	Models   []Model
}

type Handler struct {
	Name   string
	Params string
	Body   string
}

type Model struct {
	Name       string
	Type       string
	Properties map[string]string
}
