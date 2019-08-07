package controller

import (
	"github.com/lbroudoux/project-igniter-operator/pkg/controller/projectignition"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, projectignition.Add)
}
