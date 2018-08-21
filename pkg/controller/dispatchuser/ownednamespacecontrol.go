package dispatchuser

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	netsys_v1 "github.com/hantaowang/dispatch/pkg/apis/netsysio/v1"
	lister_v1 "github.com/hantaowang/dispatch/pkg/client/listers/netsysio/v1"
	netsys_client "github.com/hantaowang/dispatch/pkg/client/clientset/versioned"
	core_v1 "k8s.io/api/core/v1"

	"github.com/hantaowang/dispatch/pkg/controller"

	"fmt"
)

type OwnedNamespaceControl interface {
	ListForUser(owner string)				([]*netsys_v1.OwnedNamespace, error)
	Get(owner, namespace string)			(*netsys_v1.OwnedNamespace, error)
	Create(owner, namespace string)			(*netsys_v1.OwnedNamespace, error)
	Delete(owner, namespace string) 		error
}

type RealOwnedNamespaceControl struct {
	onLister			lister_v1.OwnedNamespaceNamespaceLister
	netsys_client		netsys_client.Interface
	original_client		kubernetes.Interface
}

func (ronc RealOwnedNamespaceControl) ListForUser(owner string) ([]*netsys_v1.OwnedNamespace, error) {
	m := map[string]string{
		"ownerID": owner,
	}
	s := labels.Set(m).AsSelector()
	return ronc.onLister.List(s)
}

func (ronc RealOwnedNamespaceControl) Get(owner, namespace string) (*netsys_v1.OwnedNamespace, error) {
	return ronc.onLister.Get(controller.NameFunc(owner, namespace))
}

func (ronc RealOwnedNamespaceControl) Create(owner, namespace string) (*netsys_v1.OwnedNamespace, error) {

	_, err := ronc.original_client.CoreV1().Namespaces().Get(namespace, meta_v1.GetOptions{})
	if errors.IsNotFound(err) {
		nSpec := &core_v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: namespace}}
		_, err = ronc.original_client.CoreV1().Namespaces().Create(nSpec)
	}
	if err != nil {
		return nil, err
	}

	if _, err := ronc.Get(owner, namespace); err != nil {
		if errors.IsNotFound(err) {
			on := netsys_v1.OwnedNamespace{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: controller.NameFunc(owner, namespace),
					Namespace: dispatchNamespace,
					Labels: map[string]string{
						"ownerID": owner,
					},
				},
				Spec: netsys_v1.OwnedNamespaceSpec{
					OwnerID: owner,
					Namespace: namespace,
				},
			}
			return ronc.netsys_client.NetsysV1().OwnedNamespaces(dispatchNamespace).Create(&on)
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("already exists")
	}
}

func (ronc RealOwnedNamespaceControl) Delete(owner, namespace string) error {
	if _, err := ronc.Get(owner, namespace); err != nil && errors.IsNotFound(err){
		return ronc.netsys_client.NetsysV1().OwnedNamespaces(dispatchNamespace).Delete(controller.NameFunc(owner, namespace), nil)
	} else {
		return err
	}
}

