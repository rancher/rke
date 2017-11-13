package client

const (
	StatusCauseType         = "statusCause"
	StatusCauseFieldField   = "field"
	StatusCauseFieldMessage = "message"
	StatusCauseFieldType    = "type"
)

type StatusCause struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
}
