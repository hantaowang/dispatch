package cmd

import (
	"github.com/hantaowang/dispatch/pkg/client"
	"github.com/hantaowang/dispatch/pkg/controller/dispatchuser"
	dispatchuser_informer "github.com/hantaowang/dispatch/pkg/client/informers/externalversions/dispatchuser/v1"
	ownednamespace_informer "github.com/hantaowang/dispatch/pkg/client/informers/externalversions/ownednamespace/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/hantaowang/dispatch/pkg/client/informers/externalversions"
	"k8s.io/client-go/informers"
)

func Start(stopCh chan struct{}) {

	clientsets := client.GetKubernetesClient()

	netsysInformerFactory := externalversions.NewSharedInformerFactory(clientsets.NetsysClient, 0)
	originalInformerFactory := informers.NewSharedInformerFactory(clientsets.OriginalClient, 0)

	sharedDispatchUserInformer := netsysInformerFactory.DispatchUser().V1().DispatchUsers()
	sharedOwnedNamespaceInformer := netsysInformerFactory.OwnedNamespace().V1().OwnedNamespaces()

	sharedServiceAccountInformer := originalInformerFactory.Core().V1().ServiceAccounts()

	duc := dispatchuser.NewDispatchUserController(sharedDispatchUserInformer, sharedOwnedNamespaceInformer,
		sharedServiceAccountInformer, clientsets)

	go duc.Run(1, stopCh)

	<- stopCh
}
