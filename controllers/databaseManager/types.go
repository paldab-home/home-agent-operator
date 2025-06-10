package databasemanager

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DatabaseManagerController struct {
	Client   client.Client
	Scheme   *runtime.Scheme
	Database string
}
