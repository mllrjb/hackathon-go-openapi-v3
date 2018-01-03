package example

type UpdateCaseHandler_VndLogRhythmCaseV1 interface {
	Handle(params UpdateCaseParams, content CaseEditableFieldsV1) Responder
}

type UpdateCaseHandler_VndLogRhythmCaseV2 interface {
	Handle(params UpdateCaseParams, content CaseEditableFieldsV2) Responder
}

// headers, query, path params
type UpdateCaseParams struct {
	ID      string
	Headers interface{}
	Path    string
	Query   string
}

type UpdateCase_Ok_Headers struct{}

// Accept: application/json
type UpdateCase_Ok struct {
	Body    CaseV2
	Headers UpdateCase_Ok_Headers
}

// Accept: application/vnd.logrhythm.case.v1+json
type UpdateCase_Ok_VndLogRhythmCaseV1 struct {
	Body    CaseV1
	Headers UpdateCase_Ok_Headers
}

// Accept: application/vnd.logrhythm.case.v2+json
type UpdateCase_Ok_VndLogRhythmCaseV2 struct {
	Body    CaseV2
	Headers UpdateCase_Ok_Headers
}

// schemas.go
// responses.go
// params.go
// {operation}.go
