package dispatchuser

import (
	"time"
	"strings"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/util/metrics"
	"k8s.io/apimachinery/pkg/runtime/schema"

	ownednamespace_lister "github.com/hantaowang/dispatch/pkg/client/listers/ownednamespace/v1"
	dispatchuser_lister "github.com/hantaowang/dispatch/pkg/client/listers/dispatchuser/v1"
	dispatchuser_informer "github.com/hantaowang/dispatch/pkg/client/informers/externalversions/dispatchuser/v1"
	ownednamespace_informer "github.com/hantaowang/dispatch/pkg/client/informers/externalversions/ownednamespace/v1"
	dispatchuser_api "github.com/hantaowang/dispatch/pkg/apis/dispatchuser/v1"

	"github.com/hantaowang/dispatch/pkg/cmd"
	"log"
	"fmt"
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

	// returns true when the DispatchUser cache is ready
	duListerSynced cache.InformerSynced
	onListerSynced cache.InformerSynced

	// dispatchusers that have been queued up for processing by workers
	queue workqueue.RateLimitingInterface

	// clients to modify resources
	clientsets	cmd.ClientSets
}

// NewNamespaceController creates a new NamespaceController
func NewDispatchUserController(
	duInformer	dispatchuser_informer.DispatchUserInformer,
	onInformer ownednamespace_informer.OwnedNamespaceInformer,
	clientSets cmd.ClientSets,
	) *DispatchUserController {


	if clientSets.OriginalClient != nil &&  clientSets.OriginalClient.CoreV1().RESTClient().GetRateLimiter() != nil {
		metrics.RegisterMetricAndTrackRateLimiterUsage("dispatch-user-controller",
			clientSets.OriginalClient.CoreV1().RESTClient().GetRateLimiter())
	}

	duc := &DispatchUserController{
		GroupVersionKind: dispatchuser_api.SchemeGroupVersion.WithKind("DispatchUser"),
		clientsets:       clientSets,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "dispatchuser"),
	}


	duInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		//AddFunc:    duc.enqueueDispatchUser,
		//UpdateFunc: duc.enqueueDispatchUser,
		//DeleteFunc: duc.enqueueDispatchUser,
	})

	duc.duLister = duInformer.Lister()
	duc.duListerSynced = duInformer.Informer().HasSynced

	duc.onLister = onInformer.Lister()
	duc.onListerSynced = onInformer.Informer().HasSynced

	return duc
}

// Run begins watching and syncing.
func (duc *DispatchUserController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer duc.queue.ShutDown()

	controllerName := strings.ToLower(duc.Kind)
	log.Printf("Starting %v controller", controllerName)
	defer log.Printf("Shutting down %v controller", controllerName)

	if !controller.WaitForCacheSync(duc.Kind, stopCh, duc.duListerSynced, duc.onListerSynced) {
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
	key, quit := duc.queue.Get()
	if quit {
		return false
	}
	defer duc.queue.Done(key)

	err := duc.syncHandler(key.(string))
	if err == nil {
		duc.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("Sync %q failed with %v", key, err))
	duc.queue.AddRateLimited(key)

	return true
}

func (duc *DispatchUserController) syncHandler(x string) error {
	return nil
}
