package v1

import (
	"encoding/json"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	ClusterNodesKind         = "ClusterNode"
	ClusterNodesName         = "clusternodes"
	ClusterNodesSingularName = "clusternode"
)

type ClusterNodesGetter interface {
	ClusterNodes(namespace string) ClusterNodeInterface
}

var _ ClusterNodeInterface = &clusternodes{}

type ClusterNodeInterface interface {
	Create(*ClusterNode) (*ClusterNode, error)
	Get(name string, opts metav1.GetOptions) (*ClusterNode, error)
	Update(*ClusterNode) (*ClusterNode, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error
}

type clusternodes struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
	ns         string
}

func newClusterNodes(r rest.Interface, c *dynamic.Client) *clusternodes {
	return &clusternodes{
		r,
		c.Resource(
			&metav1.APIResource{
				Kind:       ClusterNodesKind,
				Name:       ClusterNodesName,
				Namespaced: false,
			},
			"",
		),
		"",
	}
}

func (p *clusternodes) Create(o *ClusterNode) (*ClusterNode, error) {
	up, err := UnstructuredFromClusterNode(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Create(up)
	if err != nil {
		return nil, err
	}

	return ClusterNodeFromUnstructured(up)
}

func (p *clusternodes) Get(name string, opts metav1.GetOptions) (*ClusterNode, error) {
	obj, err := p.client.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return ClusterNodeFromUnstructured(obj)
}

func (p *clusternodes) Update(o *ClusterNode) (*ClusterNode, error) {
	up, err := UnstructuredFromClusterNode(o)
	if err != nil {
		return nil, err
	}

	curp, err := p.Get(o.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get current version for update")
	}
	up.SetResourceVersion(curp.ObjectMeta.ResourceVersion)

	up, err = p.client.Update(up)
	if err != nil {
		return nil, err
	}

	return ClusterNodeFromUnstructured(up)
}

func (p *clusternodes) Delete(name string, options *metav1.DeleteOptions) error {
	return p.client.Delete(name, options)
}

func (p *clusternodes) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := p.restClient.Get().
		Namespace(p.ns).
		Resource(ClusterNodesName).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var prom ClusterNodeList
	return &prom, json.Unmarshal(b, &prom)
}

func (p *clusternodes) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := p.restClient.Get().
		Prefix("watch").
		Namespace(p.ns).
		Resource(ClusterNodesName).
		FieldsSelectorParam(nil).
		Stream()
	if err != nil {
		return nil, err
	}
	return watch.NewStreamWatcher(&clusterDecoder{
		dec:   json.NewDecoder(r),
		close: r.Close,
	}), nil
}

func (p *clusternodes) DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error {
	return p.client.DeleteCollection(dopts, lopts)
}

func ClusterNodeFromUnstructured(r *unstructured.Unstructured) (*ClusterNode, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var p ClusterNode
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	p.TypeMeta.Kind = ClusterNodesKind
	p.TypeMeta.APIVersion = Group + "/" + Version
	return &p, nil
}

func UnstructuredFromClusterNode(p *ClusterNode) (*unstructured.Unstructured, error) {
	p.TypeMeta.Kind = ClusterNodesKind
	p.TypeMeta.APIVersion = Group + "/" + Version
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var r unstructured.Unstructured
	if err := json.Unmarshal(b, &r.Object); err != nil {
		return nil, err
	}
	return &r, nil
}

type clusterNodeDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *clusterNodeDecoder) Close() {
	d.close()
}

func (d *clusterNodeDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object ClusterNode
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
