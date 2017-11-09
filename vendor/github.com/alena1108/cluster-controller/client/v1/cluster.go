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
	ClustersKind         = "Cluster"
	ClustersName         = "clusters"
	ClustersSingularName = "cluster"
)

type ClustersGetter interface {
	Clusters(namespace string) ClusterInterface
}

var _ ClusterInterface = &clusters{}

type ClusterInterface interface {
	Create(*Cluster) (*Cluster, error)
	Get(name string, opts metav1.GetOptions) (*Cluster, error)
	Update(*Cluster) (*Cluster, error)
	Delete(name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (runtime.Object, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error
}

type clusters struct {
	restClient rest.Interface
	client     *dynamic.ResourceClient
	ns         string
}

func newClusters(r rest.Interface, c *dynamic.Client) *clusters {
	return &clusters{
		r,
		c.Resource(
			&metav1.APIResource{
				Kind:       ClustersKind,
				Name:       ClustersName,
				Namespaced: false,
			},
			"",
		),
		"",
	}
}

func (p *clusters) Create(o *Cluster) (*Cluster, error) {
	up, err := UnstructuredFromCluster(o)
	if err != nil {
		return nil, err
	}

	up, err = p.client.Create(up)
	if err != nil {
		return nil, err
	}

	return ClusterFromUnstructured(up)
}

func (p *clusters) Get(name string, opts metav1.GetOptions) (*Cluster, error) {
	obj, err := p.client.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return ClusterFromUnstructured(obj)
}

func (p *clusters) Update(o *Cluster) (*Cluster, error) {
	up, err := UnstructuredFromCluster(o)
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

	return ClusterFromUnstructured(up)
}

func (p *clusters) Delete(name string, options *metav1.DeleteOptions) error {
	return p.client.Delete(name, options)
}

func (p *clusters) List(opts metav1.ListOptions) (runtime.Object, error) {
	req := p.restClient.Get().
		Namespace(p.ns).
		Resource(ClustersName).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		return nil, err
	}
	var prom ClusterList
	return &prom, json.Unmarshal(b, &prom)
}

func (p *clusters) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	r, err := p.restClient.Get().
		Prefix("watch").
		Namespace(p.ns).
		Resource(ClustersName).
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

func (p *clusters) DeleteCollection(dopts *metav1.DeleteOptions, lopts metav1.ListOptions) error {
	return p.client.DeleteCollection(dopts, lopts)
}

func ClusterFromUnstructured(r *unstructured.Unstructured) (*Cluster, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var p Cluster
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	p.TypeMeta.Kind = ClustersKind
	p.TypeMeta.APIVersion = Group + "/" + Version
	return &p, nil
}

func UnstructuredFromCluster(p *Cluster) (*unstructured.Unstructured, error) {
	p.TypeMeta.Kind = ClustersKind
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

type clusterDecoder struct {
	dec   *json.Decoder
	close func() error
}

func (d *clusterDecoder) Close() {
	d.close()
}

func (d *clusterDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var e struct {
		Type   watch.EventType
		Object Cluster
	}
	if err := d.dec.Decode(&e); err != nil {
		return watch.Error, nil, err
	}
	return e.Type, &e.Object, nil
}
