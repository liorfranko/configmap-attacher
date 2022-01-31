package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/liorfranko/configmap-attacher/api/types/v1alpha1"
	clientV1alpha1 "github.com/liorfranko/configmap-attacher/clientset/v1alpha1"
	"github.com/liorfranko/configmap-attacher/options"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Auth required for out of cluster connections
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Auth required for out of cluster connections
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	apiClient  *kubernetes.Clientset
	argoClient *clientV1alpha1.ExampleV1Alpha1Client
}

// NewClient creates a new client to get data from kubernetes masters
func NewClient(opts *options.Options) (*Client, error) {
	// Get right config to connect to kubernetes
	var config *rest.Config
	if opts.IsInCluster {
		log.Infof("Creating InCluster config to communicate with Kubernetes master")
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		// Try to read currently set kubernetes config from your local kube config
		log.Infof("Looking for Kubernetes config to communicate with Kubernetes master")
		kubeConfigPath, err := getKubeConfigPath()
		if err != nil {
			return nil, err
		}
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("read kubeconfig: %v", err)
		}
	}

	v1alpha1.AddToScheme(scheme.Scheme)

	argoClient, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes argo client: '%v'", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes main client: '%v'", err)
	}

	return &Client{
		apiClient:  client,
		argoClient: argoClient,
	}, nil
}

// IsHealthy returns whether the kubernetes client is able to get a list of all pods
func (c *Client) IsHealthy() bool {
	_, err := c.apiClient.CoreV1().Pods(metav1.NamespaceSystem).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("kubernetes client is not healthy")
		return false
	}

	return true
}

func (c *Client) GetRolloutInfo(namespace string, rolloutName string) (string, error) {
	rollout, err := c.argoClient.Rollouts(namespace).Get(rolloutName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return rollout.Status.CurrentPodHash, nil
}

func (c *Client) GetReplicaSetInfo(namespace string, replicaset string) (types.UID, error) {
	out, err := c.apiClient.AppsV1().ReplicaSets(namespace).Get(context.Background(), replicaset, metav1.GetOptions{})

	if err != nil {
		return "", err
	}

	return out.ObjectMeta.GetUID(), err
	// c.ReplicaSets(namespace).Get(replicaSetName, metav1.GetOptions{})
}

func (c *Client) PatchConfigmap(configmap string, namespace string, rollout string, newRs string, uid types.UID) {
	// Creating the OwnerReferences for patching
	trueVar := true
	newOwnerReferences := []metav1.OwnerReference{
		{
			Kind:               "ReplicaSet",
			Name:               (rollout + "-" + newRs),
			APIVersion:         "apps/v1",
			UID:                types.UID(uid),
			Controller:         &trueVar,
			BlockOwnerDeletion: &trueVar,
		},
	}

	// Creating the OwnerReferences for patching using Patch command
	patch := fmt.Sprintf(`{"metadata":{"ownerReferences":[{"apiVersion":"apps/v1","blockOwnerDeletion":true,"controller":true,"kind":"ReplicaSet","name":"%s-%s","uid":"%s"}]}}`, rollout, newRs, uid)

	log.Debugf("newOwnerReferences is: %s", newOwnerReferences)
	log.Debugf("patch is: %s", patch)

	// Get the configmap object
	configmapObj, err := c.apiClient.CoreV1().ConfigMaps(namespace).Get(context.Background(), configmap, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("failed to get configmap: %s, err: %s", configmap, err)
	}

	log.Debugf("Going to patch configmap '%s' using SetOwnerReferences ", configmapObj)
	// Patch the configmap using SetOwnerReferences
	configmapObj.ObjectMeta.SetOwnerReferences(newOwnerReferences)
	log.Debugf("Configmap %s has been patched using SetOwnerReferences", configmap)

	configmapObj, err = c.apiClient.CoreV1().ConfigMaps(namespace).Get(context.Background(), configmap, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get configmap: %s, after patching using SetOwnerReferences, err: %s", configmap, err)
	}

	// Get the configmap's OwnerReferences
	obj := configmapObj.ObjectMeta.GetOwnerReferences()
	log.Debugf("OwnerRegerences after patching using SetOwnerReferences is: %s", obj)

	// Compare the configmap's OwnerReferences to what it needs to be
	if !reflect.DeepEqual(obj, newOwnerReferences) {
		log.Debugf("Patching failed using SetOwnerReferences, Wanted: %s, Got: %s", newOwnerReferences, obj)
		// Patch the configmap using Patch command
		log.Debugf("Going to patch configmap %s using Patch command", configmap)
		out, err := c.apiClient.CoreV1().ConfigMaps(namespace).Patch(context.Background(), configmap, types.MergePatchType, []byte(patch), v1.PatchOptions{})
		if err != nil {
			log.Fatalf("Failed to patch the configmap: %s, using Patch command, err: %s", configmap, err)
		}
		log.Debugf("Configmap %s has been patched using Patch command, output is: %s", configmap, out)

		configmapObj, err = c.apiClient.CoreV1().ConfigMaps(namespace).Get(context.Background(), configmap, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Failed to get configmap: %s, after patching with .Patch, err: %s", configmap, err)
		}

		obj = configmapObj.ObjectMeta.GetOwnerReferences()
		log.Debugf("OwnerRegerences after patching with .Patch is: %s", obj)
		if !reflect.DeepEqual(obj, newOwnerReferences) {
			log.Fatalf("Patching failed using Patch command, Wanted: %s, Got: %s", newOwnerReferences, obj)
		}
	}
}

// getKubeConfigPath returns the filepath to the local kubeConfig file or fails if it couldn't find it
func getKubeConfigPath() (string, error) {
	home := os.Getenv("HOME")

	// Mac OS
	if home != "" {
		configPath := filepath.Join(home, ".kube", "config")
		_, err := os.Stat(configPath)
		if err == nil {
			return configPath, nil
		}
	}

	// Windows
	home = os.Getenv("USERPROFILE")
	if home != "" {
		configPath := filepath.Join(home, ".kube", "config")
		_, err := os.Stat(configPath)
		if err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("couldn't find home directory to look for the kube config")
}
