package dispatchuser

import (
	"time"
	"strings"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	ownednamespace_lister "github.com/hantaowang/dispatch/pkg/client/listers/ownednamespace/v1"
	dispatchuser_lister "github.com/hantaowang/dispatch/pkg/client/listers/dispatchuser/v1"
	dispatchuser_informer "github.com/hantaowang/dispatch/pkg/client/informers/externalversions/dispatchuser/v1"
	ownednamespace_informer "github.com/hantaowang/dispatch/pkg/client/informers/externalversions/ownednamespace/v1"
	dispatchuser_api "github.com/hantaowang/dispatch/pkg/apis/dispatchuser/v1"

	"github.com/hantaowang/dispatch/pkg/cmd"

	"log"
	"fmt"

	informer_v1 "k8s.io/client-go/informers/core/v1"
	lister_v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/controller"

	dispatchuser "github.com/hantaowang/dispatch/pkg/apis/dispatchuser/v1"
)

// NamespaceController is responsible for performing actions dependent upon a namespace phase
type DispatchUserController struct {
	// GroupVersionKind indicates the controller type.
	// Different instances of this struct may handle different GVKs.
	// For example, this struct can be used (with adapters) to handle ReplicationController.
	schema.GroupVersionKind

	// lister that can list DispatchUsers from a shared cache
	duLister dispatchuser_lister.DispatchUserLister
	onLister ownednamespace_lister.OwnedNamespaceLister
	saLister lister_v1.ServiceAccountLister

	// returns true when the DispatchUser cache is ready
	duListerSynced cache.InformerSynced
	onListerSynced cache.InformerSynced
	saListerSynced cache.InformerSynced

	// clients to modify resources
	clientsets	cmd.ClientSets

	// Buffered channel of events to be done
	workqueue 	chan DispatchUserEvent
}

type DispatchUserEvent struct {
	action		string
	old			dispatchuser.DispatchUser
	new			dispatchuser.DispatchUser
}

// NewNamespaceController creates a new NamespaceController
func NewDispatchUserController(
	duInformer	dispatchuser_informer.DispatchUserInformer,
	onInformer ownednamespace_informer.OwnedNamespaceInformer,
	saInformer	informer_v1.ServiceAccountInformer,
	clientSets cmd.ClientSets,
	) *DispatchUserController {

	duc := &DispatchUserController{
		GroupVersionKind: dispatchuser_api.SchemeGroupVersion.WithKind("DispatchUser"),
		clientsets: clientSets,
		workqueue: make(chan DispatchUserEvent, 100),
	}

	duInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {
			duc.workqueue <- DispatchUserEvent{
				action: "add",
				new: obj.(dispatchuser.DispatchUser),
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			duc.workqueue <- DispatchUserEvent{
				action: "update",
				old: oldObj.(dispatchuser.DispatchUser),
				new: newObj.(dispatchuser.DispatchUser),
			}
		},
		DeleteFunc:    func(obj interface{}) {
			duc.workqueue <- DispatchUserEvent{
				action: "delete",
				old: obj.(dispatchuser.DispatchUser),
			}
		},
	})

	duc.duLister = duInformer.Lister()
	duc.duListerSynced = duInformer.Informer().HasSynced

	duc.onLister = onInformer.Lister()
	duc.onListerSynced = onInformer.Informer().HasSynced

	duc.saLister = saInformer.Lister()
	duc.saListerSynced = saInformer.Informer().HasSynced

	return duc
}

// Run begins watching and syncing.
func (duc *DispatchUserController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	controllerName := strings.ToLower(duc.Kind)
	log.Printf("Starting %v controller", controllerName)
	defer log.Printf("Shutting down %v controller", controllerName)

	if !controller.WaitForCacheSync(duc.Kind, stopCh, duc.duListerSynced, duc.onListerSynced, duc.saListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(duc.worker, time.Second, stopCh)
	}

	<-stopCh
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (duc *DispatchUserController) worker() {
	for duc.processNextWorkItem() {
	}
}

func (duc *DispatchUserController) processNextWorkItem() bool {
	event := <- duc.workqueue

	var err error
	if event.action == "add" {
		err = duc.addHandler(event)
	} else if event.action == "update" {
		err = duc.updateHandler(event)
	} else if event.action == "delete" {
		err = duc.deleteHandler(event)
	} else {
		err = fmt.Errorf("event action not recoginized %s", event.action)
	}

	if err != nil {
		return false
	}

	return true
}
