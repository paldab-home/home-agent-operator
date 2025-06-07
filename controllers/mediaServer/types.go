package mediaserver

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

type MediaServerController struct {
	Client client.Client
	Scheme *runtime.Scheme
}

type VolumeInfo struct {
	PvcName         string   `json:"pvcName"`
	Namespace       string   `json:"namespace"`
	SizeBytes       int64    `json:"sizeBytes"`
	SizeGB          float64  `json:"sizeGB"`
	ActualSizeBytes *int64   `json:"actualSizeBytes"`
	ActualSizeGB    *float64 `json:"actualSizeGB"`
	UpdatedAt       string   `json:"updatedAt"`
}

type VolumeInfoMemStore struct {
	sync.RWMutex
	volumeInfo VolumeInfo
}
