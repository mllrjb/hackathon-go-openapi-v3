package example

// schemas.go
type CaseEditableFieldsV1 struct {
	Name string
}

type CaseEditableFieldsV2 struct {
	Name   string
	Status string
}

type CaseV1 struct {
	Id            string
	Name          string
	CreatedBy     Person
	LastUpdatedBy Person
}

type CaseV2 struct {
	Id            string
	Name          string
	Status        string
	CreatedBy     Person
	LastUpdatedBy Person
}

type Person struct {
	ID   number
	Name string
}
