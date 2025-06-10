package databasemanager

import (
	"context"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func scaleUpDatabase(ctx context.Context, k8sClient client.Client, name string) {
	sts := &appsv1.StatefulSet{}

	// TODO maybe CRD
	databaseStsPayload := types.NamespacedName{
		Name:      "",
		Namespace: "",
	}

	if err := k8sClient.Get(ctx, databaseStsPayload, sts); err != nil {
		zap.L().Error("could not find database", zap.String("databaseName", name), zap.String("namespace", ""), zap.Error(err))
		return
	}

	if *sts.Spec.Replicas > 0 {
		return
	}

	var one int32 = 1
	sts.Spec.Replicas = &one
	k8sClient.Update(ctx, sts)
}

func scaleDownDatabase(ctx context.Context, k8sClient client.Client, name string) {
	sts := &appsv1.StatefulSet{}

	// TODO
	databaseStsPayload := types.NamespacedName{
		Name:      "",
		Namespace: "",
	}

	if err := k8sClient.Get(ctx, databaseStsPayload, sts); err != nil {
		zap.L().Error("could not find database", zap.String("databaseName", name), zap.String("namespace", ""), zap.Error(err))
		return
	}

	if *sts.Spec.Replicas == 0 {
		return
	}

	var zero int32 = 0
	sts.Spec.Replicas = &zero
	k8sClient.Update(ctx, sts)
}

