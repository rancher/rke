package types

import (
	"encoding/json"
	"net/http"
)

type ValuesMap struct {
	Foo map[string]interface{}
}

type RawResource struct {
	ID          string                 `json:"id,omitempty" yaml:"id,omitempty"`
	Type        string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Schema      *Schema                `json:"-" yaml:"-"`
	Links       map[string]string      `json:"links" yaml:"links"`
	Actions     map[string]string      `json:"actions" yaml:"actions"`
	Values      map[string]interface{} `json:",inline"`
	ActionLinks bool                   `json:"-"`
}

func (r *RawResource) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	for k, v := range r.Values {
		data[k] = v
	}
	if r.ID != "" {
		data["id"] = r.ID
	}
	data["type"] = r.Type
	data["links"] = r.Links
	if r.ActionLinks {
		data["actionLinks"] = r.Actions
	} else {
		data["action"] = r.Actions
	}
	return json.Marshal(data)
}

type ActionHandler func(actionName string, action *Action, request *APIContext) error

type RequestHandler func(request *APIContext) error

type Validator func(request *APIContext, data map[string]interface{}) error

type Formatter func(request *APIContext, resource *RawResource)

type ErrorHandler func(request *APIContext, err error)

type ResponseWriter interface {
	Write(apiContext *APIContext, code int, obj interface{})
}

type AccessControl interface {
	CanCreate(schema *Schema) bool
	CanList(schema *Schema) bool
}

type APIContext struct {
	Action             string
	ID                 string
	Type               string
	Link               string
	Method             string
	Schema             *Schema
	Schemas            *Schemas
	Version            *APIVersion
	ResponseFormat     string
	ReferenceValidator ReferenceValidator
	ResponseWriter     ResponseWriter
	QueryOptions       *QueryOptions
	Body               map[string]interface{}
	URLBuilder         URLBuilder
	AccessControl      AccessControl
	SubContext         map[string]interface{}

	Request  *http.Request
	Response http.ResponseWriter
}

func (r *APIContext) WriteResponse(code int, obj interface{}) {
	r.ResponseWriter.Write(r, code, obj)
}

var (
	ASC  = SortOrder("asc")
	DESC = SortOrder("desc")
)

type QueryOptions struct {
	Sort       Sort
	Pagination *Pagination
	Conditions []*QueryCondition
}

type ReferenceValidator interface {
	Validate(resourceType, resourceID string) bool
	Lookup(resourceType, resourceID string) *RawResource
}

type URLBuilder interface {
	Current() string
	Collection(schema *Schema) string
	ResourceLink(resource *RawResource) string
	RelativeToRoot(path string) string
	//Link(resource Resource, name string) string
	//ReferenceLink(resource Resource) string
	//ReferenceByIdLink(resourceType string, id string) string
	Version(version string) string
	ReverseSort(order SortOrder) string
	SetSubContext(subContext string)
}

type Store interface {
	ByID(apiContext *APIContext, schema *Schema, id string) (map[string]interface{}, error)
	List(apiContext *APIContext, schema *Schema, opt *QueryOptions) ([]map[string]interface{}, error)
	Create(apiContext *APIContext, schema *Schema, data map[string]interface{}) (map[string]interface{}, error)
	Update(apiContext *APIContext, schema *Schema, data map[string]interface{}, id string) (map[string]interface{}, error)
	Delete(apiContext *APIContext, schema *Schema, id string) error
}
