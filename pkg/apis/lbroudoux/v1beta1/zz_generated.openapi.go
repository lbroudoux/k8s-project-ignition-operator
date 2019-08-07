// +build !

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1beta1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1.ProjectIgnition":       schema_pkg_apis_lbroudoux_v1beta1_ProjectIgnition(ref),
		"github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1.ProjectIgnitionSpec":   schema_pkg_apis_lbroudoux_v1beta1_ProjectIgnitionSpec(ref),
		"github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1.ProjectIgnitionStatus": schema_pkg_apis_lbroudoux_v1beta1_ProjectIgnitionStatus(ref),
	}
}

func schema_pkg_apis_lbroudoux_v1beta1_ProjectIgnition(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ProjectIgnition is the Schema for the projectignitions API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1.ProjectIgnitionSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1.ProjectIgnitionStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/lbroudoux/project-igniter-operator/pkg/apis/lbroudoux/v1beta1.ProjectIgnitionSpec", "github.com/lbroudoux/project-igniter-operator/pkg/apis/apiextensions/v1beta1.ProjectIgnitionStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_lbroudoux_v1beta1_ProjectIgnitionSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ProjectIgnitionSpec defines the desired state of ProjectIgnition",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_lbroudoux_v1beta1_ProjectIgnitionStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ProjectIgnitionStatus defines the observed state of ProjectIgnition",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}
