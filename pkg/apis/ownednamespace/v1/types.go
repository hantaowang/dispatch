package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OwnedNamespace is a namespace that has been claimed by a User
type OwnedNamespace struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	Spec OwnedNamespaceSpec `json:"spec"`
}

// OwnedNamespaceSpec is the spec for a OwnedNamespace resource
type OwnedNamespaceSpec struct {
	OwnerID		string	`json:"ownerID"`
	Namespace	string	`json:"namespace"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OwnedNamespaceList is a list of OwnedNamespace resources
type OwnedNamespaceList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []OwnedNamespace `json:"items"`
}