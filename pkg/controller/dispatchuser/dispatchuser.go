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

	informer_v1 "k8s.io/client-go/informers/core/v1"

	"github.com/hantaowang/dispatch/pkg/client"
)

const (
	dispatchNamespace = "dispatch"
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

	// resource controls
	saControl	ServiceAccountControl
	onControl	OwnedNamespaceControl

	// clients to modify resources
	clientsets	client.ClientSets

	// Buffered channel of events to be done
	workqueue 	chan DispatchUserEvent
}

type DispatchUserEvent struct {
	action		string
	old			*netsys_v1.DispatchUser
	new			*netsys_v1.DispatchUser
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
			if obj.(*netsys_v1.DispatchUser).Namespace != dispatchNamespace {
				return
			}
			duc.workqueue <- DispatchUserEvent{
				action: "add",
				new: obj.(*netsys_v1.DispatchUser),
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if newObj.(*netsys_v1.DispatchUser).Namespace != dispatchNamespace {
				return
			}
			duc.workqueue <- DispatchUserEvent{
				action: "update",
				old: oldObj.(*netsys_v1.DispatchUser),
				new: newObj.(*netsys_v1.DispatchUser),
			}
		},
		DeleteFunc:    func(obj interface{}) {
			if obj.(*netsys_v1.DispatchUser).Namespace != dispatchNamespace {
				return
			}
			duc.workqueue <- DispatchUserEvent{
				action: "delete",
				old: obj.(*netsys_v1.DispatchUser),
			}
		},
	})

	duc.duLister = duInformer.Lister()
	duc.duListerSynced = duInformer.Informer().HasSynced

	duc.onLister = onInformer.Lister()
	duc.onListerSynced = onInformer.Informer().HasSynced

	duc.saControl = RealServiceAccountControl{
		saLister: saInformer.Lister().ServiceAccounts(dispatchNamespace),
		client: clientSets.OriginalClient,
	}

	duc.onControl = RealOwnedNamespaceControl{
		onLister: onInformer.Lister().OwnedNamespaces(dispatchNamespace),
		original_client: clientSets.OriginalClient,
		netsys_client: clientSets.NetsysClient,
	}

	duc.saListerSynced = saInformer.Informer().HasSynced

	return duc
}

// Run begins watching and syncing.
func (duc *DispatchUserController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	fmt.Printf("Starting %s controller\n", duc.Kind)
	defer fmt.Printf("Shutting down %v controller\n", duc.Kind)

	for !(duc.duListerSynced() && duc.onListerSynced() && duc.saListerSynced()) {
		time.Sleep(time.Second)
	}

	for i := 0; i < workers; i++ {
		go wait.Until(duc.worker, time.Second, stopCh)
	}

	<-stopCh
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (duc *DispatchUserController) worker() {
	fmt.Printf("Starting a %s worker\n", duc.Kind)
	for duc.processNextWorkItem() {
	}
}

func (duc *DispatchUserController) processNextWorkItem() bool {
	event := <- duc.workqueue

	var err error
	if event.action == "add" {
		err = duc.addHandler(event)
	} else if event.action == "update" {
		err = duc.syncOwnedNamespaces(event.new)
	} else if event.action == "delete" {
		err = duc.deleteHandler(event)
	} else {
		err = fmt.Errorf("event action not recoginized %s", event.action)
	}

	if err != nil {
		fmt.Printf("Error processing DispatchUser %s: %s", event.action, err)
		return false
	}

	return true
}

func (duc *DispatchUserController) addHandler(e DispatchUserEvent) error {
	_, err := duc.saControl.Create(e.new.Spec.UserID)
	if err != nil && err.Error() != "already exists" {
		return err
	}
	return duc.syncOwnedNamespaces(e.new)
}

func (duc *DispatchUserController) syncOwnedNamespaces(u *netsys_v1.DispatchUser) error {
	currentNamespaces, err := duc.onControl.ListForUser(u.Spec.UserID)
	if err != nil {
		return err
	}
	currentSet := make(map[string]bool, len(currentNamespaces))
	futureSet := make(map[string]bool, len(u.Spec.Namespaces))

	for _, n := range currentNamespaces {
		currentSet[n.Spec.Namespace] = true
	}
	for _, n := range u.Spec.Namespaces {
		futureSet[n] = true
	}

	for k := range currentSet {
		if _, ok := futureSet[k]; !ok {
			err = duc.onControl.Delete(u.Spec.UserID, k)
			if err != nil {
				return err
			}
		}
	}

	for k := range futureSet {
		if _, ok := currentSet[k]; !ok {
			_, err = duc.onControl.Create(u.Spec.UserID, k)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (duc *DispatchUserController) deleteHandler(e DispatchUserEvent) error {
	err := duc.saControl.Delete(e.new.Spec.UserID)
	if err != nil {
		return err
	}
	currentNamespaces, err := duc.onControl.ListForUser(e.old.Spec.UserID)
	for _, n := range currentNamespaces {
		err = duc.onControl.Delete(e.old.Spec.UserID, n.Spec.Namespace)
		if err != nil {
			return err
		}
	}
	return nil
}