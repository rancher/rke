package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterRoleTemplateLifecycle interface {
	Initialize(obj *ClusterRoleTemplate) error
	Remove(obj *ClusterRoleTemplate) error
	Updated(obj *ClusterRoleTemplate) error
}

type clusterRoleTemplateLifecycleAdapter struct {
	lifecycle ClusterRoleTemplateLifecycle
}

func (w *clusterRoleTemplateLifecycleAdapter) Initialize(obj runtime.Object) error {
	return w.lifecycle.Initialize(obj.(*ClusterRoleTemplate))
}

func (w *clusterRoleTemplateLifecycleAdapter) Finalize(obj runtime.Object) error {
	return w.lifecycle.Remove(obj.(*ClusterRoleTemplate))
}

func (w *clusterRoleTemplateLifecycleAdapter) Updated(obj runtime.Object) error {
	return w.lifecycle.Updated(obj.(*ClusterRoleTemplate))
}

func NewClusterRoleTemplateLifecycleAdapter(name string, client ClusterRoleTemplateInterface, l ClusterRoleTemplateLifecycle) ClusterRoleTemplateHandlerFunc {
	adapter := &clusterRoleTemplateLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, adapter, client.ObjectClient())
	return func(key string, obj *ClusterRoleTemplate) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
