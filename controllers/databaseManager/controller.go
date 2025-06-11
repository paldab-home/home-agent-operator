package databasemanager

import (
	"context"
	"paldab/home-agent-operator/config"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var databaseInstancesMemoryStore = &DatabaseInstanceMemory{}

func (r *DatabaseManagerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var sts appsv1.StatefulSet
	err := r.Client.Get(ctx, req.NamespacedName, &sts)

	// if target is sts
	if err == nil {
		databaseName, ok := sts.Labels[config.DATABASE_INSTANCE_NAME_LABEL]

		if !ok {
			return ctrl.Result{}, nil
		}

		handleDatabaseInstanceRegistration(sts, databaseName)

		zap.L().Info("database instances found", zap.Any("instances", databaseInstancesMemoryStore.data), zap.Int("len instances", len(databaseInstancesMemoryStore.data))) // Change
	}

	for _, db := range databaseInstancesMemoryStore.data {
		var pods corev1.PodList

		err := r.Client.List(ctx, &pods, client.MatchingLabels{config.POD_NEEDS_DATABASE_LABEL: db.DatabaseName})
		if err != nil {
			continue
		}

		targetDB := types.NamespacedName{
			Name:      db.StatefulSetName,
			Namespace: db.Namespace,
		}

		if len(pods.Items) > 0 {
			scaleUpDatabase(ctx, r.Client, targetDB)
		} else {
			scaleDownDatabase(ctx, r.Client, targetDB)
		}
	}

	return ctrl.Result{}, nil
}

// Also watch custom CRDS
func (r *DatabaseManagerController) RegisterController(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&corev1.Pod{}).
		Watches(&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
			sts := o.(*appsv1.StatefulSet)
			var databaseName string = ""

			for k, v := range sts.Labels {
				if k == config.DATABASE_INSTANCE_NAME_LABEL {
					databaseName = v
				}
			}

			if databaseName != "" {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      sts.ObjectMeta.Name,
							Namespace: sts.ObjectMeta.Namespace,
						},
					},
				}
			}

			return nil
		})).
		Complete(r)
}

func handleDatabaseInstanceRegistration(sts appsv1.StatefulSet, databaseName string) {
	for _, data := range databaseInstancesMemoryStore.data {
		if data.StatefulSetName == sts.GetName() && data.Namespace == sts.GetNamespace() {
			databaseInstancesMemoryStore.Lock()
			data.DatabaseName = databaseName
			data.Replicas = int(*sts.Spec.Replicas)
			databaseInstancesMemoryStore.Unlock()
			return
		}
	}

	// if database does not exist
	newMemoryEntry := DatabaseMemoryData{
		StatefulSetName: sts.GetName(),
		Namespace:       sts.GetNamespace(),
		DatabaseName:    databaseName,
		Replicas:        int(*sts.Spec.Replicas),
	}

	databaseInstancesMemoryStore.Lock()
	defer databaseInstancesMemoryStore.Unlock()

	// If no databases yet, initialize
	if len(databaseInstancesMemoryStore.data) == 0 {
		databaseInstancesMemoryStore.data = []DatabaseMemoryData{newMemoryEntry}
	} else {
		databaseInstancesMemoryStore.data = append(databaseInstancesMemoryStore.data, newMemoryEntry)
	}
}
