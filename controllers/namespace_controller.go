/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"os"
	"strconv"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=namespaces/status,verbs=get
//+kubebuilder:rbac:groups=operators.coreos.com,resources=operatorgroups,verbs=get;list;watch;create

// Reconcile is part of the main Kubernetes reconciliation loop
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("R namespace")

	var namespace corev1.Namespace
	if err := r.Get(ctx, req.NamespacedName, &namespace); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "GetError", "namespace", req.NamespacedName)
			return ctrl.Result{}, err
		}
	}

	var operatorGroupList operatorsv1.OperatorGroupList
	if err := r.List(ctx, &operatorGroupList, client.InNamespace(namespace.Name)); err != nil {
		logger.Error(err, "ListError", "namespace", req.Namespace)
		return ctrl.Result{}, err
	}
	if len(operatorGroupList.Items) > 0 {
		for _, operatorGroup := range operatorGroupList.Items {
			if operatorGroup.Namespace == namespace.Name {
				return ctrl.Result{}, nil
			}
		}
	}

	var operatorGroup operatorsv1.OperatorGroup
	if err := r.Get(ctx, req.NamespacedName, &operatorGroup); err != nil {
		operatorGroup := &operatorsv1.OperatorGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespace.Name,
				Namespace: namespace.Name,
			},
			Spec: operatorsv1.OperatorGroupSpec{
				TargetNamespaces: []string{namespace.Name},
			},
		}
		if err := r.Create(ctx, operatorGroup); err != nil {
			if apierrors.IsAlreadyExists(err) {
				return ctrl.Result{}, nil
			}
			logger.Error(err, "CreateError", "operatorgroup", operatorGroup.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	maxConcurrentReconciles := 3
	if value := os.Getenv("MAX_CONCURRENT_RECONCILES"); value != "" {
		if val, err := strconv.Atoi(value); err == nil {
			maxConcurrentReconciles = val
		}
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: maxConcurrentReconciles}).
		Complete(r)
}
