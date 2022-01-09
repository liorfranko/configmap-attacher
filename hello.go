package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/liorfranko/configmap-attacher/options"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Hello() string {
	return "Hello, world"
}

func main() {
	fmt.Println(Hello())

	configMapPtr := flag.String("configmap", "", "Configmap to add the ownerReference")
	rolloutPtr := flag.String("rollout", "", "Rollout that will be the ownerReference")
	namespacePtr := flag.String("namespace", "", "The namespace of the rollout and configmap")
	dryRunPtr := flag.Bool("dryRun", false, "Run in dry-run mode")
	flag.Parse()

	fmt.Printf("configMapPtr: %s, rolloutPtr: %s, namespacePtr: %s, dryRunPtr: %t\n", *configMapPtr, *rolloutPtr, *namespacePtr, *dryRunPtr)
	if *configMapPtr == "" || *rolloutPtr == "" || *namespacePtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Initialize logrus settings
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	// Parse & validate environment variable
	opts := options.NewOptions()
	err := envconfig.Process("", opts)
	if err != nil {
		log.Fatal("Error parsing env vars into opts", err)
	}

	// Set log level from environment variable
	level, err := log.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Panicf("Loglevel could not be parsed as one of the known loglevels. See logrus documentation for valid log level inputs. Given input was: '%s'", opts.LogLevel)
	}
	log.SetLevel(level)

	// Start configmap-attacher
	log.Infof("Starting configmap-attacher v%v", opts.Version)

	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	configmap, err := clientset.CoreV1().ConfigMaps(*namespacePtr).Get(context.Background(), *configMapPtr, metav1.GetOptions{})
	if err != nil {
		log.Fatalln("failed to get configmap:", err)
	}

	// print configmaps
	log.Println("printing configmpas", configmap)
	// for i, configmap := range configmaps.Items {
	// 	log.Println(i, configmap.GetName())
	// }

	// replicasets, err := clientset.AppsV1().ReplicaSets(*namespacePtr).List(context.Background(), metav1.ListOptions{})
	// if err != nil {
	// 	log.Fatalln("failed to get ReplicaSets:", err)
	// }

	// // print pods
	// log.Println("printing replicasets")
	// for i, replicaset := range replicasets.Items {
	// 	log.Println(i, replicaset.GetObjectMeta())
	// }
}
