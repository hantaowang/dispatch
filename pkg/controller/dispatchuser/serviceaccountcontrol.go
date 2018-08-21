package dispatchuser

import (
	"k8s.io/api/core/v1"
	lister_v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
)

type ServiceAccountControl interface {
	List()				([]*v1.ServiceAccount, error)
	Get(name string) 	(*v1.ServiceAccount, error)
	Create(name string)	(*v1.ServiceAccount, error)
	Delete(name string) error
}

type RealServiceAccountControl struct {
	saLister		lister_v1.ServiceAccountNamespaceLister
	client			kubernetes.Interface
}

func (rsac RealServiceAccountControl) List() ([]*v1.ServiceAccount, error) {
	return rsac.saLister.List(labels.Everything())
}

func (rsac RealServiceAccountControl) Get(name string) (*v1.ServiceAccount, error) {
	return rsac.saLister.Get(name)
}

func (rsac RealServiceAccountControl) Create(name string) (*v1.ServiceAccount, error) {
	if _, err := rsac.Get(name); err != nil {
		if errors.IsNotFound(err) {
			sa := v1.ServiceAccount{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      name,
					Namespace: dispatchNamespace,
				},
			}
			return rsac.client.CoreV1().ServiceAccounts(dispatchNamespace).Create(&sa)
		} else {
			return nil, err
		}
	}else {
		return nil, fmt.Errorf("already exists")
	}
}

func (rsac RealServiceAccountControl) Delete(name string) error {
	if _, err := rsac.Get(name); err != nil && errors.IsNotFound(err){
		return rsac.client.CoreV1().ServiceAccounts(dispatchNamespace).Delete(name, nil)
	} else {
		return err
	}
}