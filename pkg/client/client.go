package client

import (
	"os"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	netsys_client "github.com/hantaowang/dispatch/pkg/client/clientset/versioned"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

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
		panic(fmt.Sprintf("GetClusterConfig config: %v", err))

	}

	fmt.Println("Successfully constructed config")

	// generate the client based off of the config
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("GetClusterConfig originalClient: %v", err))
	}

	fmt.Println("Successfully constructed k8s client")


	customClient, err := netsys_client.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("GetClusterConfig customClient: %v", err))
	}

	fmt.Println("Successfully constructed custom client")

	return ClientSets{
		OriginalClient: client,
		NetsysClient: customClient,
	}
}
