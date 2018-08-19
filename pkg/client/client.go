package client

import (
	"os"
	"log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	netsys_client "github.com/hantaowang/dispatch/pkg/client/clientset/versioned"
)

type ClientSets struct {
	OriginalClient			kubernetes.Interface
	NetsysClient 			netsys_client.Interface
}

// retrieve the Kubernetes cluster client from outside of the cluster
func GetKubernetesClient() ClientSets {
	// construct the path to resolve to `~/.kube/config`
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	// create the config from the path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	// generate the client based off of the config
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	customClient, err := netsys_client.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	log.Println("Successfully constructed k8s client")

	return ClientSets{
		OriginalClient: client,
		NetsysClient: customClient,
	}
}
