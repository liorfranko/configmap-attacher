package v1alpha1

//go:generate controller-gen object paths=$GOFILE

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type RolloutSpec struct {
	Replicas int `json:"replicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Rollout struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RolloutSpec   `json:"spec"`
	Status RolloutStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RolloutList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Rollout `json:"items"`
}

type RolloutStatus struct {
	CurrentPodHash string `json:"currentPodHash,omitempty" protobuf:"bytes,5,opt,name=currentPodHash"`
}
