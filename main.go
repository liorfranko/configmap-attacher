package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/liorfranko/configmap-attacher/kubernetes"
	"github.com/liorfranko/configmap-attacher/options"
	log "github.com/sirupsen/logrus"
)

// TODO wait for replicaset
// TODO Replace kubectl commands CRD client
// TODO Use SetOwnerReferences insteach of Patch
// TODO Add support for running from inside the cluster

func runCmd(str ...string) map[string]interface{} {
	cmd := exec.Command("kubectl", str...)
	log.Infof("Running kubectl command: '%v'", cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("Could not run kubectl, command: ", cmd, err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal("Could not start the CMD, command: ", cmd, err)
	}

	data, err := ioutil.ReadAll(stdout)

	if err != nil {
		log.Fatal("Error while reading the kubectl output, command: ", cmd, err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal("Error while waiting for the kubectl output, command: ", cmd, err)
	}
	var x map[string]interface{}
	json.Unmarshal([]byte(data), &x)
	return x

}
func Runner(configMapPtr string, rolloutPtr string, namespacePtr string, opts *options.Options) {
	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubernetesClient, err := kubernetes.NewClient(opts)
	if err != nil {
		log.Fatal("failed to initialize kubernetes client: '%v'", err)
	}
	newRs, err := kubernetesClient.GetRolloutInfo(namespacePtr, rolloutPtr)
	if err != nil {
		log.Fatal("currentPodHash was not found in rollout: '%v'", err)
	}

	var uid string

	// Extract the UID from the Replicaset object
	x := runCmd("-n", namespacePtr, "get", "replicasets.apps", rolloutPtr+"-"+newRs, "-o", "json")
	if val, ok := x["metadata"]; ok {
		v := val.(map[string]interface{})
		if val2, ok2 := v["uid"]; ok2 {
			uid = fmt.Sprintf("%v", val2)
		} else {
			log.Fatal("uid was not found in metadata.replicaset object")
		}
	} else {
		log.Fatal("metadata was not found in replicaset object")
	}

	kubernetesClient.GetRolloutInfo(namespacePtr, rolloutPtr)
	// Split the configmaps
	configmaps := strings.Split(configMapPtr, ",")
	// Patch each configmap
	for i, configmap := range configmaps {
		log.Debug("Checking that configmap exists: ", i, configmap)
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
		log.Fatal("Error parsing env vars into opts", err)
	}

	// Set log level from environment variable
	level, err := log.ParseLevel(opts.LogLevel)
	if err != nil {
		log.Fatal("Loglevel could not be parsed as one of the known loglevels. See logrus documentation for valid log level inputs. Given input was: '%s'", opts.LogLevel)
	}
	log.SetLevel(level)
	log.Infof("Starting configmap-attacher")
	log.Infof("configmaps: %s, rollout: %s, namespace: %s\n", *configMapPtr, *rolloutPtr, *namespacePtr)
	// Start configmap-attacher
	Runner(*configMapPtr, *rolloutPtr, *namespacePtr, opts)
	log.Infof("Done configmap-attacher")
}
