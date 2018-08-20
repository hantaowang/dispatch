package cmd

import (
	"github.com/hantaowang/dispatch/pkg/client"
	"github.com/hantaowang/dispatch/pkg/controller/dispatchuser"

	"github.com/hantaowang/dispatch/pkg/client/informers/externalversions"
	"k8s.io/client-go/informers"
)

func Start(stopCh chan struct{}) {

	clientsets := client.GetKubernetesClient()

	netsysInformerFactory := externalversions.NewSharedInformerFactory(clientsets.NetsysClient, 0)
	originalInformerFactory := informers.NewSharedInformerFactory(clientsets.OriginalClient, 0)

	sharedDispatchUserInformer := netsysInformerFactory.Netsys().V1().DispatchUsers()
	sharedOwnedNamespaceInformer := netsysInformerFactory.Netsys().V1().OwnedNamespaces()

	sharedServiceAccountInformer := originalInformerFactory.Core().V1().ServiceAccounts()

	duc := dispatchuser.NewDispatchUserController(sharedDispatchUserInformer, sharedOwnedNamespaceInformer,
		sharedServiceAccountInformer, clientsets)

	go duc.Run(1, stopCh)

	<- stopCh
}
