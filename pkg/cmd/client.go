package cmd

import (
	"os"
	"log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ownednamespace_client "github.com/hantaowang/dispatch/pkg/client/clientset/versioned/typed/ownednamespace/v1"
	dispatchuser_client "github.com/hantaowang/dispatch/pkg/client/clientset/versioned/typed/dispatchuser/v1"
)

type ClientSets struct {
	OriginalClient			kubernetes.Interface
	OwnedNamespaceClient 	ownednamespace_client.NetsysV1Interface
	DispatchUserClient		dispatchuser_client.NetsysV1Interface
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

	ownedNamespaceClient, err := ownednamespace_client.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	dispatchUserClient, err := dispatchuser_client.NewForConfig(config)
	if err != nil {
		log.Fatalf("getClusterConfig: %v", err)
	}

	log.Println("Successfully constructed k8s client")

	return ClientSets{
		OriginalClient: client,
		OwnedNamespaceClient: ownedNamespaceClient,
		DispatchUserClient: dispatchUserClient,
	}
}
