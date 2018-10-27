package generator

var lifecycleTemplate = `package {{.schema.Version.Version}}

import (
	{{.importPackage}}
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/rancher/norman/lifecycle"
)

type {{.schema.CodeName}}Lifecycle interface {
	Create(obj *{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error)
	Remove(obj *{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error)
	Updated(obj *{{.prefix}}{{.schema.CodeName}}) (*{{.prefix}}{{.schema.CodeName}}, error)
}

type {{.schema.ID}}LifecycleAdapter struct {
	lifecycle {{.schema.CodeName}}Lifecycle
}

func (w *{{.schema.ID}}LifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*{{.prefix}}{{.schema.CodeName}}))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *{{.schema.ID}}LifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*{{.prefix}}{{.schema.CodeName}}))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *{{.schema.ID}}LifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*{{.prefix}}{{.schema.CodeName}}))
	if o == nil {
		return nil, err
	}
	return o, err
}

func New{{.schema.CodeName}}LifecycleAdapter(name string, clusterScoped bool, client {{.schema.CodeName}}Interface, l {{.schema.CodeName}}Lifecycle) {{.schema.CodeName}}HandlerFunc {
	adapter := &{{.schema.ID}}LifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *{{.prefix}}{{.schema.CodeName}}) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
`
