package mediaserver

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"paldab/home-agent-operator/config"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	longhornv1beta2 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta2"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	apiGroup = "/media"
)

var volumeInfoMemStore = &VolumeInfoMemStore{}

func (r *MediaServerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// zap.L().Debug("watching resource", zap.String("resourceName", req.String()))
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
		return ctrl.Result{}, nil
	}

	isLonghornStorage := *pvc.Spec.StorageClassName != "" && strings.Contains(strings.ToLower(*pvc.Spec.StorageClassName), "longhorn")

	if !isLonghornStorage {
		return ctrl.Result{}, nil
	}

	if pvc.GetName() == "" {
		zap.L().Error("failed to find any media pvc")
	}

	volumeInfo := extractVolumeInfo(pvc, longhornVolume)

	// Store in memory
	volumeInfoMemStore.Lock()
	volumeInfoMemStore.volumeInfo = volumeInfo
	volumeInfoMemStore.Unlock()

	zap.L().Info("Media PVC found", zap.String("pvc", volumeInfoMemStore.volumeInfo.PvcName), zap.String("namespace", volumeInfoMemStore.volumeInfo.Namespace))

	return ctrl.Result{}, nil
}

func extractVolumeInfo(pvc *corev1.PersistentVolumeClaim, longhornVolume longhornv1beta2.Volume) VolumeInfo {
	sizeToGB := bytesToGB(longhornVolume.Spec.Size)
	actualSizeToGB := math.Round(bytesToGB(longhornVolume.Status.ActualSize))

	return VolumeInfo{
		PvcName:         pvc.GetName(),
		Namespace:       config.MEDIASERVER_NAMESPACE,
		SizeBytes:       longhornVolume.Spec.Size,
		ActualSizeBytes: &longhornVolume.Status.ActualSize,
		SizeGB:          sizeToGB,
		ActualSizeGB:    &actualSizeToGB,
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}
}

func (r *MediaServerController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("MediaServerManager").
		For(&longhornv1beta2.Volume{}).
		Watches(&corev1.PersistentVolumeClaim{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			pvc := o.(*corev1.PersistentVolumeClaim)
			if pvc.Labels["media"] != "true" || pvc.Spec.VolumeName == "" {
				return nil
			}

			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      pvc.Spec.VolumeName,
						Namespace: pvc.GetNamespace(),
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
