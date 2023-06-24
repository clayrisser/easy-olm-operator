package controllers

import (
	"context"
	"os"
	"strconv"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	cachev1alpha1 "gitlab.com/bitspur/easy-olm-operator/api/v1alpha1"
	"gitlab.com/bitspur/easy-olm-operator/util"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type InstallPlanReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=operators.coreos.com,resources=installplans,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=operators.coreos.com,resources=installplans/status,verbs=get

func (r *InstallPlanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("R installplan")

	installPlan := &operatorsv1alpha1.InstallPlan{}
	if err := r.Get(ctx, req.NamespacedName, installPlan); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "GetError", "installplan", req.NamespacedName)
		return ctrl.Result{}, err
	}

	manualSubscriptionList := &cachev1alpha1.ManualSubscriptionList{}
	if err := r.List(ctx, manualSubscriptionList); err != nil {
		logger.Error(err, "ListError", "manualsubscription")
		return ctrl.Result{}, err
	}
	if len(manualSubscriptionList.Items) == 0 {
		return ctrl.Result{}, nil
	}
	for _, manualSubscription := range manualSubscriptionList.Items {
		if manualSubscription.Spec.StartingCSV == installPlan.Spec.ClusterServiceVersionNames[0] {
			if !installPlan.Spec.Approved {
				if err := util.ApproveInstallPlan(ctx, r.Client, installPlan.Name, installPlan.Namespace); err != nil {
					logger.Error(err, "ApproveInstallPlanError", "installplan", installPlan.Name)
					return ctrl.Result{}, err
				}
				break
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *InstallPlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	maxConcurrentReconciles := 3
	if value := os.Getenv("MAX_CONCURRENT_RECONCILES"); value != "" {
		if val, err := strconv.Atoi(value); err == nil {
			maxConcurrentReconciles = val
		}
	}
	shouldReconcile := predicate.NewPredicateFuncs(func(object client.Object) bool {
		installPlan, ok := object.(*operatorsv1alpha1.InstallPlan)
		if !ok {
			return false
		}
		return !installPlan.Spec.Approved
	})
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{MaxConcurrentReconciles: maxConcurrentReconciles}).
		For(&operatorsv1alpha1.InstallPlan{}, builder.WithPredicates(shouldReconcile)).
		Complete(r)
}
