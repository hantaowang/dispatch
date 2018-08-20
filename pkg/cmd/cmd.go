package cmd

import (
	"github.com/hantaowang/dispatch/pkg/client"
	"github.com/hantaowang/dispatch/pkg/controller/dispatchuser"

	"github.com/hantaowang/dispatch/pkg/client/informers/externalversions"
	"k8s.io/client-go/informers"
	"fmt"
)

func Start(stopCh chan struct{}) {

	clientsets := client.GetKubernetesClient()

	fmt.Println("Creating Informer Factories")

	netsysInformerFactory := externalversions.NewSharedInformerFactory(clientsets.NetsysClient, 0)
	originalInformerFactory := informers.NewSharedInformerFactory(clientsets.OriginalClient, 0)

	fmt.Println("Creating Informers")
	sharedDispatchUserInformer := netsysInformerFactory.Netsys().V1().DispatchUsers()
	sharedOwnedNamespaceInformer := netsysInformerFactory.Netsys().V1().OwnedNamespaces()
	sharedServiceAccountInformer := originalInformerFactory.Core().V1().ServiceAccounts()

	fmt.Println("Starting Informers")
	go sharedServiceAccountInformer.Informer().Run(stopCh)
	go sharedOwnedNamespaceInformer.Informer().Run(stopCh)
	go sharedDispatchUserInformer.Informer().Run(stopCh)

	fmt.Println("Creating Controller")
	duc := dispatchuser.NewDispatchUserController(sharedDispatchUserInformer, sharedOwnedNamespaceInformer,
		sharedServiceAccountInformer, clientsets)

	go duc.Run(1, stopCh)

	<- stopCh
}
