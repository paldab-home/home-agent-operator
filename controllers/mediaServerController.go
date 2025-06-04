package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"paldab/home-agent-operator/config"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"

	longhornv1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	apiGroup = "/media"
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

var volumeInfoMemStore = &VolumeInfoMemStore{}

func (r *MediaServerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	zap.L().Debug("watching resource", zap.String("resourceName", req.String()))
	var longhornVolume longhornv1beta2.Volume

	if err := r.Client.Get(ctx, req.NamespacedName, &longhornVolume); err != nil {
		return ctrl.Result{}, nil
	}

	ks := longhornVolume.Status.KubernetesStatus
	if ks.PVName == "" || ks.Namespace != config.MEDIASERVER_NAMESPACE {
		return ctrl.Result{}, nil
	}

	pvcPayload := types.NamespacedName{
		Namespace: ks.Namespace,
		Name:      ks.PVCName,
	}

	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.Client.Get(context.TODO(), pvcPayload, pvc); err != nil {
		zap.L().Error("failed to fetch associated PVC or wrong namespace", zap.String("pvName", ks.PVName), zap.String("namespace", ks.Namespace), zap.Any("payload", pvcPayload), zap.Error(err))
		return ctrl.Result{RequeueAfter: config.ERR_RETRY_TIMEOUT}, nil
	}

	if pvc.Labels["media"] != "true" {
		zap.L().Info("PVC is not a media PVC", zap.String("pvc", pvc.Name), zap.String("namespace", pvc.GetNamespace()))
		return ctrl.Result{}, nil
	}

	isLonghornStorage := *pvc.Spec.StorageClassName != "" && strings.Contains(strings.ToLower(*pvc.Spec.StorageClassName), "longhorn")

	if !isLonghornStorage {
		return ctrl.Result{}, nil
	}

	if pvc.GetName() == "" {
		zap.L().Error("failed to find any media pvc")
	}

	sizeToGB := bytesToGB(longhornVolume.Spec.Size)
	actualSizeToGB := bytesToGB(longhornVolume.Status.ActualSize)

	volumeInfo := VolumeInfo{
		PvcName:         pvc.GetName(),
		Namespace:       config.MEDIASERVER_NAMESPACE,
		SizeBytes:       longhornVolume.Spec.Size,
		ActualSizeBytes: &longhornVolume.Status.ActualSize,
		SizeGB:          sizeToGB,
		ActualSizeGB:    &actualSizeToGB,
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}

	// Store in memory
	volumeInfoMemStore.Lock()
	volumeInfoMemStore.volumeInfo = volumeInfo
	volumeInfoMemStore.Unlock()

	return ctrl.Result{}, nil
}

func (r *MediaServerController) RegisterController(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&longhornv1beta2.Volume{}).
		Watches(&corev1.PersistentVolumeClaim{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			pvc := o.(*corev1.PersistentVolumeClaim)
			if pvc.Labels["media"] != "true" || pvc.Spec.VolumeName == "" {
				return nil
			}

			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      pvc.Spec.VolumeName,
						Namespace: pvc.Namespace,
					},
				}}
		})).
		Complete(r)
}

func (r *MediaServerController) RegisterApiEndpoints(mux *http.ServeMux) {
	mux.HandleFunc(apiGroup+"/volume-usage", func(w http.ResponseWriter, r *http.Request) {
		volumeInfoMemStore.RLock()
		defer volumeInfoMemStore.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(volumeInfoMemStore.volumeInfo)
	})
}

func bytesToGB(bytes int64) float64 {
	const bytesInGB = 1024 * 1024 * 1024
	return float64(bytes) / float64(bytesInGB)
}
