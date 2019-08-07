package apis

import (
	"github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1beta1.SchemeBuilder.AddToScheme)
}
