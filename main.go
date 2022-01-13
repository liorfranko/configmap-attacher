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
	// Get the rollout using kubectl
	x := runCmd("-n", namespacePtr, "get", "rollout", rolloutPtr, "-o", "json")
	var newRs string
	var uid string
	// Extract the new Replicaset from the rollout object
	if val, ok := x["status"]; ok {
		v := val.(map[string]interface{})
		if val3, ok := v["currentPodHash"]; ok {
			newRs = fmt.Sprintf("%v", val3)
		} else {
			log.Fatal("currentPodHash was not found in rollout status")
		}
	} else {
		log.Fatal("status was not found in rollout object")
	}

	// Extract the UID from the Replicaset object
	x = runCmd("-n", namespacePtr, "get", "replicasets.apps", rolloutPtr+"-"+newRs, "-o", "json")
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

	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubernetesClient, err := kubernetes.NewClient(opts)
	if err != nil {
		log.Fatal("failed to initialize kubernetes client: '%v'", err)
	}
	fmt.Println(kubernetesClient.IsHealthy(), uid)
	// Split the configmaps
	configmaps := strings.Split(configMapPtr, ",")
	// var cm []*corev1.ConfigMap
	for i, configmap := range configmaps {
		log.Debug("Checking that configmap exists: ", i, configmap)
		fmt.Println(kubernetesClient.IsHealthy())
		kubernetesClient.PatchConfigmap(configmap, namespacePtr, rolloutPtr, newRs, uid)
	}
}

func main() {
	// Set and parse CLI options
	configMapPtr := flag.String("configmaps", "", "Configmaps to add the ownerReference, for multiple configmaps use ',' as a separator")
	rolloutPtr := flag.String("rollout", "", "Rollout that will be the ownerReference")
	namespacePtr := flag.String("namespace", "", "The namespace of the rollout and configmap")
	flag.Parse()
	log.Debug("configMapPtr: %s, rolloutPtr: %s, namespacePtr: %s\n", *configMapPtr, *rolloutPtr, *namespacePtr)
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
	log.Infof("Configmap-attacher variables, configmap: %s, rollout: %s, namespace: %s", *configMapPtr, *rolloutPtr, *namespacePtr)
	Runner(*configMapPtr, *rolloutPtr, *namespacePtr, opts)
	log.Infof("Done configmap-attacher")
}
