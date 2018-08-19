package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DispatchUser is a user that can own namespaces
type DispatchUser struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	Spec DispatchUserSpec `json:"spec"`
}

// DispatchUserSpec is the spec for a DispatchUser resource
type DispatchUserSpec struct {
	UserID		string	`json:"userID"`
	Namespaces	[]string	`json:"namespaces"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DispatchUserList is a list of DispatchUser resources
type DispatchUserList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []DispatchUser `json:"items"`
}