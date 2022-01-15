package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	apiClient *kubernetes.Clientset
}

// NewClient creates a new client to get data from kubernetes masters
func NewClient(opts *options.Options) (*Client, error) {
	// Get right config to connect to kubernetes
	var config *rest.Config
	if opts.IsInCluster {
		log.Info("Creating InCluster config to communicate with Kubernetes master")
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		// Try to read currently set kubernetes config from your local kube config
		log.Info("Looking for Kubernetes config to communicate with Kubernetes master")
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

	clientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	rollouts, err := clientSet.Rollouts("devops-apps-01").List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rollouts found: %+v\n", rollouts)

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes main client: '%v'", err)
	}

	return &Client{
		apiClient: client,
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

func (c *Client) PatchConfigmap(configmap string, namespace string, rollout string, newRs string, uid string) {
	// trueVar := true
	// newOwnerReferences := []metav1.OwnerReference{
	// 	{
	// 		Kind:       "ReplicaSet",
	// 		Name:       (rolloutPtr + "-" + newRs),
	// 		APIVersion: "apps/v1",
	// 		UID:        types.UID(uid),
	// 		Controller: &trueVar,
	// 	},
	// }
	// fmt.Println(newOwnerReferences)
	// configmap.ObjectMeta.SetOwnerReferences(newOwnerReferences)
	// fmt.Println(new2OwnerReference)
	// Patch the configmap and set the replicaset as the owner using ownerReferences
	_, err := c.apiClient.CoreV1().ConfigMaps(namespace).Get(context.Background(), configmap, metav1.GetOptions{})
	if err != nil {
		log.Fatalln("failed to get configmap:", configmap, err)
	}
	patch := fmt.Sprintf(`{"metadata":{"ownerReferences":[{"apiVersion":"apps/v1","blockOwnerDeletion":true,"controller":true,"kind":"ReplicaSet","name":"%s-%s","uid":"%s"}]}}`, rollout, newRs, uid)
	out, err := c.apiClient.CoreV1().ConfigMaps(namespace).Patch(context.Background(), configmap, types.MergePatchType, []byte(patch), v1.PatchOptions{})
	if err != nil {
		log.Fatal("Could not patch the configmap:", configmap, err)
	}
	log.Debug("Configmap %s has been patched, output is: ", configmap, out)
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
