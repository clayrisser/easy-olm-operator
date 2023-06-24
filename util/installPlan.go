package util

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ApproveInstallPlan(ctx context.Context, cl client.Client, ns string, name string) error {
	installPlanGVR := schema.GroupVersionResource{
		Group:    "operators.coreos.com",
		Version:  "v1alpha1",
		Resource: "installplans",
	}
	installPlan := &unstructured.Unstructured{}
	installPlan.SetGroupVersionKind(installPlanGVR.GroupVersion().WithKind("InstallPlan"))
	if err := cl.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, installPlan); err != nil {
		return err
	}
	installPlan.Object["spec"] = map[string]interface{}{
		"approved": true,
	}
	if err := cl.Update(ctx, installPlan); err != nil {
		return fmt.Errorf("failed to approve InstallPlan %s/%s: %w", ns, name, err)
	}
	return nil
}
