package v1beta1

import (
	"github.com/redhat-cop/operator-utils/pkg/util/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProjectIgnitionSpec defines the desired state of ProjectIgnition
// +k8s:openapi-gen=true
type ProjectIgnitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	ProjectName                string                `json:"projectName"`
	Namespaces                 NamespacesSpec        `json:"namespaces"`
	OpenShiftMultiProjectQuota MultiProjectQuotaSpec `json:"openShiftMultiProjectQuota,omitempty"`
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// NamespacesSpec defines the desired state of Namespaces
// +k8s:openapi-gen=true
type NamespacesSpec struct {
	UseOpenShiftProject       bool `json:"useOpenShiftProject"`
	AddStageNumber            bool `json:"addStageNumber"`
	AddStageNameInDisplayName bool `json:"addStageNameInDisplayName"`
	// +kubebuilder:validation:MinItems=1
	Definitions []DefinitionSpec `json:"definitions"`
}

// DefinitionSpec defines the desired state of Definitions
// +k8s:openapi-gen=true
type DefinitionSpec struct {
	Name         string            `json:"name"`
	Annotations  []string          `json:"annotations,omitempty"`
	Labels       []LabelSpec       `json:"labels,omitempty"`
	Finalizers   []string          `json:"finalizers,omitempty"`
	RoleBindings []RoleBindingSpec `json:"roleBindings,omitempty"`
	Quotas       []string          `json:"quotas,omitempty"`
}

// LabelSpec defines the desired state of Labels
// +k8s:openapi-gen=true
type LabelSpec struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RoleBindingSpec defines the desired state of Roles
// +k8s:openapi-gen=true
type RoleBindingSpec struct {
	Role  string `json:"role"`
	User  string `json:"user,omitempty"`
	Group string `json:"group,omitempty"`
}

// MultiProjectQuotaSpec defines the desired state of MultiProjectQuota
// +k8s:openapi-gen=true
type MultiProjectQuotaSpec struct {
	ProjectAnnotationSelector string `json:"projectAnnotationSelector"`
	ProjectLabelSelector      string `json:"projectLabelSelector"`
	Quota                     string `json:"quota"`
}

// ProjectIgnitionStatus defines the observed state of ProjectIgnition
// +k8s:openapi-gen=true
type ProjectIgnitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	Namespaces           []string `json:"namespaces"`
	RoleBindings         []string `json:"roleBindings"`
	apis.ReconcileStatus `json:",inline"`
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProjectIgnition is the Schema for the projectignitions API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ProjectIgnition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectIgnitionSpec   `json:"spec,omitempty"`
	Status ProjectIgnitionStatus `json:"status,omitempty"`
}

// GetReconcileStatus - Applying https://github.com/redhat-cop/operator-utils
func (m *ProjectIgnition) GetReconcileStatus() apis.ReconcileStatus {
	return m.Status.ReconcileStatus
}

// SetReconcileStatus - Applying https://github.com/redhat-cop/operator-utils
func (m *ProjectIgnition) SetReconcileStatus(reconcileStatus apis.ReconcileStatus) {
	m.Status.ReconcileStatus = reconcileStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProjectIgnitionList contains a list of ProjectIgnition
type ProjectIgnitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectIgnition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectIgnition{}, &ProjectIgnitionList{})
}
