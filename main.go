package main

import (
	"flag"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/liorfranko/configmap-attacher/kubernetes"
	"github.com/liorfranko/configmap-attacher/options"
	log "github.com/sirupsen/logrus"
)

// TODO wait for replicaset
// TODO Use SetOwnerReferences insteach of Patch

func Runner(configMapPtr string, rolloutPtr string, namespacePtr string, opts *options.Options) {
	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubernetesClient, err := kubernetes.NewClient(opts)
	if err != nil {
		log.Fatalf("failed to initialize kubernetes client: '%v'", err)
	}
	newRs, err := kubernetesClient.GetRolloutInfo(namespacePtr, rolloutPtr)
	if err != nil {
		log.Fatalf("currentPodHash was not found in rollout: '%v'", err)
	}
	uid, err := kubernetesClient.GetReplicaSetInfo(namespacePtr, rolloutPtr+"-"+newRs)
	if err != nil {
		log.Fatalf("Replicaset was not found: '%v'", err)
	}

	// Split the configmaps
	configmaps := strings.Split(configMapPtr, ",")
	// Patch each configmap
	for i, configmap := range configmaps {
		log.Debugf("Patching configmap, number: %d, name: %s, namespace: %s, rollout: %s, replicaSet: %s, uid: %s", i, configmap, namespacePtr, rolloutPtr, newRs, uid)
		kubernetesClient.PatchConfigmap(configmap, namespacePtr, rolloutPtr, newRs, uid)
	}
}

func main() {
	// Set and parse CLI options
	configMapPtr := flag.String("configmaps", "", "Configmaps to add the ownerReference, for multiple configmaps use ',' as a separator")
	rolloutPtr := flag.String("rollout", "", "Rollout that will be the ownerReference")
	namespacePtr := flag.String("namespace", "", "The namespace of the rollout and configmap")
	flag.Parse()
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
		log.Fatalf("Error parsing env vars into opts", err)
	}

	// Set log level from environment variable
	level, err := log.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Fatalf("Loglevel could not be parsed as one of the known loglevels. See logrus documentation for valid log level inputs. Given input was: '%s'", opts.LogLevel)
	}
	log.SetLevel(level)
	log.Infof("Starting configmap-attacher version: %v", opts.Version)
	log.Infof("configmaps: %s, rollout: %s, namespace: %s, jobName: %s", *configMapPtr, *rolloutPtr, *namespacePtr, opts.JobName)
	// Start configmap-attacher
	Runner(*configMapPtr, *rolloutPtr, *namespacePtr, opts)
	log.Infof("Done configmap-attacher")
}
