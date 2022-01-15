package v1alpha1

import (
	"context"

	"github.com/liorfranko/configmap-attacher/api/types/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type RolloutInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.RolloutList, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.Rollout, error)
	Create(*v1alpha1.Rollout) (*v1alpha1.Rollout, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	// ...
}

type rolloutClient struct {
	restClient rest.Interface
	ns         string
}

func (c *rolloutClient) List(opts metav1.ListOptions) (*v1alpha1.RolloutList, error) {
	result := v1alpha1.RolloutList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("rollouts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *rolloutClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.Rollout, error) {
	result := v1alpha1.Rollout{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("rollouts").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *rolloutClient) Create(rollout *v1alpha1.Rollout) (*v1alpha1.Rollout, error) {
	result := v1alpha1.Rollout{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("rollouts").
		Body(rollout).
		Do(context.Background()).
		Into(&result)

	return &result, err
}

func (c *rolloutClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("rollouts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(context.Background())
}
