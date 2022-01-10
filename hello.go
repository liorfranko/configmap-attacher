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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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

	configmap, err := clientset.CoreV1().ConfigMaps(*namespacePtr).Get(context.Background(), *configMapPtr, metav1.GetOptions{})
	if err != nil {
		log.Fatalln("failed to get configmap:", err)
	}

	// print configmaps
	log.Println("printing configmpas", configmap)

	x := runCmd("-n", *namespacePtr, "get", "rollout", *rolloutPtr, "-o", "json")
	var newRs string
	var uid string
	if val, ok := x["status"]; ok {
		v := val.(map[string]interface{})
		if val2, ok2 := v["currentPodHash"]; ok2 {
			newRs = fmt.Sprintf("%v", val2)
		} else {
			log.Fatal("currentPodHash was not found in rollout status")
		}
	} else {
		log.Fatal("status was not found in rollout object")
	}
	x = runCmd("-n", *namespacePtr, "get", "replicasets.apps", *rolloutPtr+"-"+newRs, "-o", "json")
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
	fmt.Println(uid)
	// type ownerReferences struct {
	// 	apiVersion         string `json:"apiVersion"`
	// 	blockOwnerDeletion bool   `json:"blockOwnerDeletion"`
	// 	controller         bool   `json:"controller"`
	// 	kind               string `json:"kind"`
	// 	name               string `json:"name"`
	// 	uid                string `json:"uid"`
	// }
	// type Metadata struct {
	// 	Labels          map[string]interface{} `json:"labels"`
	// 	ownerReferences []ownerReferences      `json:"ownerReferences"`
	// }
	// patch3 := Metadata{
	// 	Labels: {"dsds": "dsdsds"},
	// 	ownerReferences: []ownerReferences{
	// 		c,
	// 	},
	// }

	// patch3.Metadata.Labels = map[string]string{}
	// patch3.Metadata.Labels["hi"] = ""
	// patch3.Metadata.ownerReferences[0] = {
	// 	"fdfd":"fdfd",
	// 	"sdsds":"dsdsdgh"
	// }
	// patch3.Metadata.ownerReferences[0] = ownerReferences{
	// 	apiVersion:         "apps/v1",
	// 	blockOwnerDeletion: true,
	// 	controller:         true,
	// 	kind:               "ReplicaSet",
	// 	name:               "sleep-264446269",
	// 	uid:                "6c078900-71f8-4bf5-8a7d-063a72c7a12b",
	// }
	patch := fmt.Sprintf(`{"metadata":{"ownerReferences":[{"apiVersion":"apps/v1","blockOwnerDeletion":true,"controller":true,"kind":"ReplicaSet","name":"%s-%s","uid":"%s"}]}}`, *rolloutPtr, newRs, uid)
	patch2 := []byte(`{"metadata":{"ownerReferences":[{"apiVersion":"apps/v1","blockOwnerDeletion":true,"controller":true,"kind":"ReplicaSet","name":"sleep-7796f4cd8f","uid":"07a80790-f2fc-46a3-8598-3d7bf110c143"}]}}}`)
	fmt.Println([]byte(patch))
	fmt.Println(patch2)
	patchJson, _ := json.Marshal(patch2)
	fmt.Println(patchJson)
	out, err2 := clientset.CoreV1().ConfigMaps(*namespacePtr).Patch(context.Background(), *configMapPtr, types.MergePatchType, []byte(patch), v1.PatchOptions{})
	if err2 != nil {
		panic(err2.Error())
	}
	log.Infof("Configmap %s has been patched, output: %s ", *configMapPtr, out)
}
