package databasemanager

import (
	"context"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func scaleUpDatabase(ctx context.Context, k8sClient client.Client, databaseSts types.NamespacedName) {
	sts := &appsv1.StatefulSet{}

	if err := k8sClient.Get(ctx, databaseSts, sts); err != nil {
		zap.L().Error("could not find requested database", zap.String("statefulset", databaseSts.Name), zap.String("namespace", databaseSts.Namespace), zap.Error(err))
		return
	}

	if *sts.Spec.Replicas > 0 {
		return
	}

	zap.L().Info("found multiple pods that require "+databaseSts.Name+" scaling up", zap.String("controller", databaseSts.Name), zap.String("namespace", databaseSts.Namespace))
	var one int32 = 1
	sts.Spec.Replicas = &one
	k8sClient.Update(ctx, sts)
}

func scaleDownDatabase(ctx context.Context, k8sClient client.Client, databaseSts types.NamespacedName) {
	sts := appsv1.StatefulSet{}

	if err := k8sClient.Get(ctx, databaseSts, &sts); err != nil {
		zap.L().Error("could not find requested database", zap.String("statefulset", databaseSts.Name), zap.String("namespace", databaseSts.Namespace), zap.Error(err))
		return
	}

	if *sts.Spec.Replicas == 0 {
		return
	}

	zap.L().Info("found no pods that require "+databaseSts.Name+" scaling down", zap.String("controller", databaseSts.Name), zap.String("namespace", databaseSts.Namespace))
	var zero int32 = 0
	sts.Spec.Replicas = &zero
	k8sClient.Update(ctx, &sts)
}
