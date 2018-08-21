package dispatchuser

import (
	"time"
	"fmt"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime/schema"
	netsys_lister "github.com/hantaowang/dispatch/pkg/client/listers/netsysio/v1"
	netsys_informer "github.com/hantaowang/dispatch/pkg/client/informers/externalversions/netsysio/v1"
	netsys_v1 "github.com/hantaowang/dispatch/pkg/apis/netsysio/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	informer_v1 "k8s.io/client-go/informers/rbac/v1"
	lister_v1 "k8s.io/client-go/listers/rbac/v1"

	rbac_v1 "k8s.io/api/rbac/v1"

	"github.com/hantaowang/dispatch/pkg/client"
	"github.com/hantaowang/dispatch/pkg/controller"

)

const (
	dispatchNamespace = "dispatch"
)

// NamespaceController is responsible for performing actions dependent upon a namespace phase
type OwnedNamespaceController struct {
	// GroupVersionKind indicates the controller type.
	// Different instances of this struct may handle different GVKs.
	// For example, this struct can be used (with adapters) to handle ReplicationController.
	schema.GroupVersionKind

	// lister that can list DispatchUsers from a shared cache
	onLister netsys_lister.OwnedNamespaceLister

	// returns true when the DispatchUser cache is ready
	onListerSynced 	cache.InformerSynced

	// clients to modify resources
	clientsets	client.ClientSets

	// Buffered channel of events to be done
	workqueue 	chan OwnedNamespaceEvent
}

type OwnedNamespaceEvent struct {
	action		string
	old			*netsys_v1.OwnedNamespace
	new			*netsys_v1.OwnedNamespace
}

// NewOwnedNamespaceController creates a new OwnedNamespaceController
func NewOwnedNamespaceController(
	onInformer  netsys_informer.OwnedNamespaceInformer,
	clientSets client.ClientSets,
	) *OwnedNamespaceController {

	onc := &OwnedNamespaceController{
		GroupVersionKind: netsys_v1.SchemeGroupVersion.WithKind("OwnedNamespace"),
		clientsets: clientSets,
		workqueue: make(chan OwnedNamespaceEvent, 100),
	}

	onInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {
			onc.workqueue <- OwnedNamespaceEvent{
				action: "add",
				new: obj.(*netsys_v1.OwnedNamespace),
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			onc.workqueue <- OwnedNamespaceEvent{
				action: "update",
				old: oldObj.(*netsys_v1.OwnedNamespace),
				new: newObj.(*netsys_v1.OwnedNamespace),
			}
		},
		DeleteFunc:    func(obj interface{}) {
			onc.workqueue <- OwnedNamespaceEvent{
				action: "delete",
				old: obj.(*netsys_v1.OwnedNamespace),
			}
		},
	})

	onc.onLister = onInformer.Lister()
	onc.onListerSynced = onInformer.Informer().HasSynced

	return onc
}

// Run begins watching and syncing.
func (onc *OwnedNamespaceController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	fmt.Printf("Starting %s controller\n", onc.Kind)
	defer fmt.Printf("Shutting down %v controller\n", onc.Kind)

	for !onc.onListerSynced() {
		time.Sleep(time.Second)
	}

	for i := 0; i < workers; i++ {
		go wait.Until(onc.worker, time.Second, stopCh)
	}

	<-stopCh
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (onc *OwnedNamespaceController) worker() {
	fmt.Printf("Starting a %s worker\n", onc.Kind)
	for onc.processNextWorkItem() {
	}
}

func (onc *OwnedNamespaceController) processNextWorkItem() bool {
	event := <- onc.workqueue

	var err error
	if event.action == "add" {
		err = onc.addHandler(event)
	} else if event.action == "update" {
		// drop
		return true
	} else if event.action == "delete" {
		err = onc.deleteHandler(event)
	} else {
		err = fmt.Errorf("event action not recoginized %s", event.action)
	}

	if err != nil {
		fmt.Printf("Error processing OwnedNamespace %s: %s", event.action, err)
		return false
	}

	return true
}

func (onc *OwnedNamespaceController) addHandler(e OwnedNamespaceEvent) error {
	rb := rbac_v1.RoleBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      controller.NameFunc(e.new.Spec.OwnerID, e.new.Namespace),
			Namespace: e.new.Namespace,
		},
		Subjects: []rbac_v1.Subject{
			{
				Kind: "ServiceAccount",
				Name: e.new.Spec.OwnerID,
				Namespace: dispatchNamespace,
			},
		},
		RoleRef: rbac_v1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err := onc.clientsets.OriginalClient.RbacV1().RoleBindings(e.new.Namespace).Create(&rb)

	return err
}

func (onc *OwnedNamespaceController) deleteHandler(e OwnedNamespaceEvent) error {
	return onc.clientsets.OriginalClient.RbacV1().RoleBindings(e.old.Namespace).Delete(e.old.Spec.OwnerID, nil)
}