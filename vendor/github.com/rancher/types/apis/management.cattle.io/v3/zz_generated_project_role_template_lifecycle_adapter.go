package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ProjectRoleTemplateLifecycle interface {
	Initialize(obj *ProjectRoleTemplate) error
	Remove(obj *ProjectRoleTemplate) error
	Updated(obj *ProjectRoleTemplate) error
}

type projectRoleTemplateLifecycleAdapter struct {
	lifecycle ProjectRoleTemplateLifecycle
}

func (w *projectRoleTemplateLifecycleAdapter) Initialize(obj runtime.Object) error {
	return w.lifecycle.Initialize(obj.(*ProjectRoleTemplate))
}

func (w *projectRoleTemplateLifecycleAdapter) Finalize(obj runtime.Object) error {
	return w.lifecycle.Remove(obj.(*ProjectRoleTemplate))
}

func (w *projectRoleTemplateLifecycleAdapter) Updated(obj runtime.Object) error {
	return w.lifecycle.Updated(obj.(*ProjectRoleTemplate))
}

func NewProjectRoleTemplateLifecycleAdapter(name string, client ProjectRoleTemplateInterface, l ProjectRoleTemplateLifecycle) ProjectRoleTemplateHandlerFunc {
	adapter := &projectRoleTemplateLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, adapter, client.ObjectClient())
	return func(key string, obj *ProjectRoleTemplate) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
