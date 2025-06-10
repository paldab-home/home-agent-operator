package databasemanager

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

type DatabaseManagerController struct {
	Client client.Client
	Scheme *runtime.Scheme
}

type DatabaseMemoryData struct {
	StatefulSetName string
	Namespace       string
	DatabaseName    string // Actual reference like mysql (label key of operator.paldab.io/database-name)
	Replicas        int
}

type DatabaseInstanceMemory struct {
	sync.RWMutex
	data []DatabaseMemoryData
}
