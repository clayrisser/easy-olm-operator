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

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	easyolmv1alpha1 "gitlab.com/bitspur/easy-olm-operator/api/v1alpha1"
	"gitlab.com/bitspur/easy-olm-operator/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ManualSubscriptionReconciler reconciles a ManualSubscription object
type ManualSubscriptionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=easyolm.bitspur.com,resources=manualsubscriptions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=easyolm.bitspur.com,resources=manualsubscriptions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=easyolm.bitspur.com,resources=manualsubscriptions/finalizers,verbs=update

const manualSubscriptionFinalizer = "manualsubscription.finalizers.easyolm.bitspur.com"

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ManualSubscription object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ManualSubscriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("R manualsubscription")

	var manualSubscription easyolmv1alpha1.ManualSubscription
	if err := r.Get(ctx, req.NamespacedName, &manualSubscription); err != nil {
		logger.Error(err, "GetError", "manualsubscription", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if manualSubscription.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(manualSubscription.ObjectMeta.Finalizers, manualSubscriptionFinalizer) {
			controllerutil.AddFinalizer(&manualSubscription, manualSubscriptionFinalizer)
			if err := r.Update(ctx, &manualSubscription); err != nil {
				logger.Error(err, "UpdateError", "manualsubscription", manualSubscription.Name)
				return ctrl.Result{}, err
			}
		}
	} else {
		if util.ContainsString(manualSubscription.ObjectMeta.Finalizers, manualSubscriptionFinalizer) {
			subscription := &operatorsv1alpha1.Subscription{}
			if err := r.Get(ctx, types.NamespacedName{Name: manualSubscription.Name, Namespace: manualSubscription.Namespace}, subscription); err != nil {
				logger.Error(err, "GetError", "subscription", manualSubscription.Name)
				return ctrl.Result{}, err
			}
			if err := r.Delete(ctx, subscription); err != nil {
				logger.Error(err, "DeleteError", "subscription", manualSubscription.Name)
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(&manualSubscription, manualSubscriptionFinalizer)
			if err := r.Update(ctx, &manualSubscription); err != nil {
				logger.Error(err, "UpdateError", "manualsubscription", manualSubscription.Name)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	subscription := &operatorsv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      manualSubscription.Name,
			Namespace: manualSubscription.Namespace,
		},
		Spec: &operatorsv1alpha1.SubscriptionSpec{
			CatalogSource:          manualSubscription.Spec.Source,
			CatalogSourceNamespace: manualSubscription.Spec.SourceNamespace,
			Channel:                manualSubscription.Spec.Channel,
			InstallPlanApproval:    "Manual",
			Package:                manualSubscription.Spec.Name,
			StartingCSV:            manualSubscription.Spec.StartingCSV,
		},
	}
	manualSubscription.Status.ObservedGeneration = manualSubscription.Generation
	if err := r.Create(ctx, subscription); err != nil {
		if errors.IsAlreadyExists(err) {
			subscription := &operatorsv1alpha1.Subscription{}
			if err := r.Get(ctx, req.NamespacedName, subscription); err != nil {
				logger.Error(err, "GetError", "subscription", subscription.Name)
				meta.SetStatusCondition(&manualSubscription.Status.Conditions,
					metav1.Condition{
						Type:               "Failed",
						Status:             "True",
						Reason:             "GetError",
						Message:            err.Error(),
						LastTransitionTime: metav1.Now(),
					})
				if err := r.Status().Update(ctx, &manualSubscription); err != nil {
					logger.Error(err, "StatusUpdateError", "manualSubscription", manualSubscription.Name)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}
			subscription.Spec.CatalogSource = manualSubscription.Spec.Source
			subscription.Spec.CatalogSourceNamespace = manualSubscription.Spec.SourceNamespace
			subscription.Spec.Channel = manualSubscription.Spec.Channel
			subscription.Spec.InstallPlanApproval = "Manual"
			subscription.Spec.Package = manualSubscription.Spec.Name
			subscription.Spec.StartingCSV = manualSubscription.Spec.StartingCSV
			if err := r.Update(ctx, subscription); err != nil {
				logger.Error(err, "UpdateError", "subscription", subscription.Name)
				meta.SetStatusCondition(&manualSubscription.Status.Conditions,
					metav1.Condition{
						Type:               "Failed",
						Status:             "True",
						Reason:             "UpdateError",
						Message:            err.Error(),
						LastTransitionTime: metav1.Now(),
					})
				if err := r.Status().Update(ctx, &manualSubscription); err != nil {
					logger.Error(err, "StatusUpdateError", "manualSubscription", manualSubscription.Name)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}
			meta.SetStatusCondition(&manualSubscription.Status.Conditions,
				metav1.Condition{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             "UpdateSuccessful",
					Message:            "Subscription successfully updated",
					LastTransitionTime: metav1.Now(),
				})
			meta.RemoveStatusCondition(&manualSubscription.Status.Conditions, "Failed")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "CreateError", "subscription", subscription.Name)
		meta.SetStatusCondition(&manualSubscription.Status.Conditions,
			metav1.Condition{
				Type:               "Failed",
				Status:             "True",
				Reason:             "CreateError",
				Message:            err.Error(),
				LastTransitionTime: metav1.Now(),
			})
		if err := r.Status().Update(ctx, &manualSubscription); err != nil {
			logger.Error(err, "StatusUpdateError", "manualsubscription", manualSubscription.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	manualSubscription.Status.SubscriptionRef = subscription.Name
	meta.SetStatusCondition(&manualSubscription.Status.Conditions,
		metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			Reason:             "CreateSuccessful",
			Message:            "Subscription successfully created",
			LastTransitionTime: metav1.Now(),
		})
	meta.RemoveStatusCondition(&manualSubscription.Status.Conditions, "Failed")
	if err := r.Status().Update(ctx, &manualSubscription); err != nil {
		logger.Error(err, "StatusUpdateError", "manualSubscription", manualSubscription.Name)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManualSubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	maxConcurrentReconciles := 3
	if value := os.Getenv("MAX_CONCURRENT_RECONCILES"); value != "" {
		if val, err := strconv.Atoi(value); err == nil {
			maxConcurrentReconciles = val
		}
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&easyolmv1alpha1.ManualSubscription{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: maxConcurrentReconciles}).
		Complete(r)
}
