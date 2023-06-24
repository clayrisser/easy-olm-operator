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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "gitlab.com/bitspur/easy-olm-operator/api/v1alpha1"
	"gitlab.com/bitspur/easy-olm-operator/util"
)

// CrdRefReconciler reconciles a CrdRef object
type CrdRefReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.bitspur.com,resources=crdrefs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.bitspur.com,resources=crdrefs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.bitspur.com,resources=crdrefs/finalizers,verbs=update

const crdRefFinalizer = "crdref.finalizers.cache.bitspur.com"

var crdAutoDelete = os.Getenv("CRD_AUTO_DELETE") == "1"

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CrdRef object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CrdRefReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("R crdref")

	var crdRef cachev1alpha1.CrdRef
	if err := r.Get(ctx, req.NamespacedName, &crdRef); err != nil {
		logger.Error(err, "GetError", "crdref", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if crdRef.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(crdRef.ObjectMeta.Finalizers, crdRefFinalizer) {
			controllerutil.AddFinalizer(&crdRef, crdRefFinalizer)
			if err := r.Update(ctx, &crdRef); err != nil {
				logger.Error(err, "UpdateError", "crdref", crdRef.Name)
				return ctrl.Result{}, err
			}
		}
	} else {
		if util.ContainsString(crdRef.ObjectMeta.Finalizers, crdRefFinalizer) {
			if crdAutoDelete {
				var crdRefs cachev1alpha1.CrdRefList
				if err := r.List(ctx, &crdRefs); err != nil {
					logger.Error(err, "ListError", "crdref")
					return ctrl.Result{}, err
				}
				deleteCrd := true
				for _, cr := range crdRefs.Items {
					if cr.Status.CrdRef == crdRef.Status.CrdRef {
						deleteCrd = false
						break
					}
				}
				if deleteCrd {
					crd, err := util.GetCrd(ctx, r.Client, crdRef.Spec.Crd)
					if err != nil {
						logger.Error(err, "GetError", "crd", crdRef.Spec.Crd)
						return ctrl.Result{}, err
					}
					if err := r.Delete(ctx, crd); err != nil {
						logger.Error(err, "DeleteError", "crd", crdRef.Status.CrdRef)
						return ctrl.Result{}, err
					}
				}
			}
			controllerutil.RemoveFinalizer(&crdRef, crdRefFinalizer)
			if err := r.Update(ctx, &crdRef); err != nil {
				logger.Error(err, "UpdateError", "crdref", crdRef.Name)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	crdRef.Status.ObservedGeneration = crdRef.Generation
	crd, err := util.GetCrd(ctx, r.Client, crdRef.Spec.Crd)
	if err != nil {
		logger.Error(err, "FetchCrdError", "crd", crdRef.Spec.Crd)
		return ctrl.Result{}, err
	}
	if err = r.Client.Get(ctx, client.ObjectKey{Namespace: crd.GetNamespace(), Name: crd.GetName()}, crd); err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.Client.Create(ctx, crd); err != nil {
				logger.Error(err, "CreateError", "crd", crd.GetName())
				meta.SetStatusCondition(&crdRef.Status.Conditions,
					metav1.Condition{
						Type:               "Failed",
						Status:             "True",
						Reason:             "CreateError",
						Message:            err.Error(),
						LastTransitionTime: metav1.Now(),
					})
				if err := r.Status().Update(ctx, &crdRef); err != nil {
					logger.Error(err, "StatusUpdateError", "crdref", crdRef.Name)
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}
			crdRef.Status.CrdRef = crd.GetName()
			meta.SetStatusCondition(&crdRef.Status.Conditions,
				metav1.Condition{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             "CreateSuccessful",
					Message:            "CrdRef successfully created",
					LastTransitionTime: metav1.Now(),
				})
			meta.RemoveStatusCondition(&crdRef.Status.Conditions, "Failed")
			if err := r.Status().Update(ctx, &crdRef); err != nil {
				logger.Error(err, "StatusUpdateError", "crdref", crdRef.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		logger.Error(err, "GetError", "crd", crd.GetName())
		meta.SetStatusCondition(&crdRef.Status.Conditions,
			metav1.Condition{
				Type:               "Failed",
				Status:             "True",
				Reason:             "GetError",
				Message:            err.Error(),
				LastTransitionTime: metav1.Now(),
			})
		if err := r.Status().Update(ctx, &crdRef); err != nil {
			logger.Error(err, "StatusUpdateError", "crdref", crdRef.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CrdRefReconciler) SetupWithManager(mgr ctrl.Manager) error {
	maxConcurrentReconciles := 3
	if value := os.Getenv("MAX_CONCURRENT_RECONCILES"); value != "" {
		if val, err := strconv.Atoi(value); err == nil {
			maxConcurrentReconciles = val
		}
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.CrdRef{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: maxConcurrentReconciles}).
		Complete(r)
}
