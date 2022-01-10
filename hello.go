package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/liorfranko/configmap-attacher/options"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Patch struct {
	Metadata Metadata
}

type Metadata struct {
	Labels          map[string]string `json:"labels"`
	ownerReferences []ownerReferences `json:"ownerReferences"`
}
type ownerReferences struct {
	apiVersion         string `json:"apiVersion"`
	blockOwnerDeletion bool   `json:"blockOwnerDeletion"`
	controller         bool   `json:"controller"`
	kind               string `json:"kind"`
	name               string `json:"name"`
	uid                string `json:"uid"`
}

func Hello() string {
	return "Hello, world"
}

func runCmd(str ...string) map[string]interface{} {
	cmd := exec.Command("kubectl", str...)
	log.Infof("Running kubectl command: '%v'", cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadAll(stdout)

	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	var x map[string]interface{}
	json.Unmarshal([]byte(data), &x)
	return x

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
	log.Infof("Configmap-attacher variables, configmap: %s, rollout: %s, namespace: %s", *configMapPtr, *rolloutPtr, *namespacePtr)

	// // Bootstrap k8s configuration from local 	Kubernetes config file
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

	// configmap, err := clientset.CoreV1().ConfigMaps(*namespacePtr).Get(context.Background(), *configMapPtr, metav1.GetOptions{})
	// if err != nil {
	// 	log.Fatalln("failed to get configmap:", err)
	// }

	// // print configmaps
	// log.Println("printing configmpas", configmap)

	// x := runCmd("-n", *namespacePtr, "get", "rollout", *rolloutPtr, "-o", "json")
	// var newRs string
	// var uid string
	// if val, ok := x["status"]; ok {
	// 	v := val.(map[string]interface{})
	// 	if val2, ok2 := v["currentPodHash"]; ok2 {
	// 		newRs = fmt.Sprintf("%v", val2)
	// 	} else {
	// 		log.Fatal("currentPodHash was not found in rollout status")
	// 	}
	// } else {
	// 	log.Fatal("status was not found in rollout object")
	// }
	// x = runCmd("-n", *namespacePtr, "get", "replicasets.apps", *rolloutPtr+"-"+newRs, "-o", "json")
	// if val, ok := x["metadata"]; ok {
	// 	v := val.(map[string]interface{})
	// 	if val2, ok2 := v["uid"]; ok2 {
	// 		uid = fmt.Sprintf("%v", val2)
	// 	} else {
	// 		log.Fatal("uid was not found in metadata.replicaset object")
	// 	}
	// } else {
	// 	log.Fatal("metadata was not found in replicaset object")
	// }
	// fmt.Println(newRs)
	// fmt.Println(uid)
	newRs := "7796f4cd8f"
	uid := "07a80790-f2fc-46a3-8598-3d7bf110c143"
	fmt.Println(newRs)
	fmt.Println(uid)
	c := ownerReferences{
		apiVersion:         "apps/v1",
		blockOwnerDeletion: true,
		controller:         true,
		kind:               "ReplicaSet",
		name:               (*rolloutPtr + "-" + newRs),
		uid:                uid,
	}
	fmt.Println(c)
	patch3 := struct {
		Metadata struct {
			ownerReferences [1]ownerReferences `json:"ownerReferences"`
		} `json:"metadata"`
	}{}
	patch3.Metadata.ownerReferences[0] = c
	patchJson, _ := json.Marshal(patch3)
	out, err2 := clientset.CoreV1().ConfigMaps(*namespacePtr).Patch(context.Background(), *configMapPtr, types.MergePatchType, patchJson, v1.PatchOptions{})
	if err2 != nil {
		panic(err2.Error())
	}
	fmt.Println(out)
}
