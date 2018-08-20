package dispatchuser

import (
	"time"
	"strings"
	"log"
	"fmt"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime/schema"
	netsys_lister "github.com/hantaowang/dispatch/pkg/client/listers/netsysio/v1"
	netsys_informer "github.com/hantaowang/dispatch/pkg/client/informers/externalversions/netsysio/v1"
	netsys_v1 "github.com/hantaowang/dispatch/pkg/apis/netsysio/v1"

	informer_v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/kubernetes/pkg/controller"

	"github.com/hantaowang/dispatch/pkg/client"
)

const (
	namespace = "dispatch"
)

// NamespaceController is responsible for performing actions dependent upon a namespace phase
type DispatchUserController struct {
	// GroupVersionKind indicates the controller type.
	// Different instances of this struct may handle different GVKs.
	// For example, this struct can be used (with adapters) to handle ReplicationController.
	schema.GroupVersionKind

	// lister that can list DispatchUsers from a shared cache
	duLister netsys_lister.DispatchUserLister
	onLister netsys_lister.OwnedNamespaceLister

	// returns true when the DispatchUser cache is ready
	duListerSynced cache.InformerSynced
	onListerSynced cache.InformerSynced
	saListerSynced cache.InformerSynced

	// service account control
	saControl	ServiceAccountControl

	// clients to modify resources
	clientsets	client.ClientSets

	// Buffered channel of events to be done
	workqueue 	chan DispatchUserEvent
}

type DispatchUserEvent struct {
	action		string
	old			netsys_v1.DispatchUser
	new			netsys_v1.DispatchUser
}

// NewNamespaceController creates a new NamespaceController
func NewDispatchUserController(
	duInformer	netsys_informer.DispatchUserInformer,
	onInformer  netsys_informer.OwnedNamespaceInformer,
	saInformer	informer_v1.ServiceAccountInformer,
	clientSets client.ClientSets,
	) *DispatchUserController {

	duc := &DispatchUserController{
		GroupVersionKind: netsys_v1.SchemeGroupVersion.WithKind("DispatchUser"),
		clientsets: clientSets,
		workqueue: make(chan DispatchUserEvent, 100),
	}

	duInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {
			duc.workqueue <- DispatchUserEvent{
				action: "add",
				new: obj.(netsys_v1.DispatchUser),
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			duc.workqueue <- DispatchUserEvent{
				action: "update",
				old: oldObj.(netsys_v1.DispatchUser),
				new: newObj.(netsys_v1.DispatchUser),
			}
		},
		DeleteFunc:    func(obj interface{}) {
			duc.workqueue <- DispatchUserEvent{
				action: "delete",
				old: obj.(netsys_v1.DispatchUser),
			}
		},
	})

	duc.duLister = duInformer.Lister()
	duc.duListerSynced = duInformer.Informer().HasSynced

	duc.onLister = onInformer.Lister()
	duc.onListerSynced = onInformer.Informer().HasSynced

	duc.saControl = RealServiceAccountControl{
		saLister: saInformer.Lister().ServiceAccounts(namespace),
		client: clientSets.OriginalClient,
	}

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
		log.Printf("Error processing DispatchUser %s: %s", event.action, err)
		return false
	}

	return true
}

func (duc *DispatchUserController) addHandler(e DispatchUserEvent) error {
	_, err := duc.saControl.Create("dispatch:" + e.new.Spec.UserID)
	return err
}

func (duc *DispatchUserController) updateHandler(e DispatchUserEvent) error {
	return nil
}

func (duc *DispatchUserController) deleteHandler(e DispatchUserEvent) error {
	return nil
}