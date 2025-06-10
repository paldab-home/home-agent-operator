package databasemanager

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *DatabaseManagerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var pods corev1.PodList

	if err := r.Client.List(ctx, &pods, client.MatchingLabels{"database": r.Database}); err != nil {
		return ctrl.Result{}, nil
	}

	if len(pods.Items) > 0 {
		scaleUpDatabase(ctx, r.Client, r.Database)
	} else {
		scaleDownDatabase(ctx, r.Client, r.Database)
	}

	return ctrl.Result{}, nil
}

// Also watch custom CRDS
func (r *DatabaseManagerController) RegisterController(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&corev1.Pod{}).Complete(r)
}
