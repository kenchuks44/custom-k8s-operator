package controllers

import (
	"context"

	appsv1 "example.com/my-operator/api/v1"
	corev1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DeploymentSyncReconciler reconciles a DeploymentSync object
type DeploymentSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile method to sync Deployment
func (r *DeploymentSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("deploymentsync", req.NamespacedName)

	// Fetch the DeploymentSync instance
	deploymentSync := &appsv1.DeploymentSync{}
	if err := r.Get(ctx, req.NamespacedName, deploymentSync); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch the source Deployment
	sourceDeployment := &corev1.Deployment{}
	sourceDeploymentName := types.NamespacedName{
		Namespace: deploymentSync.Spec.SourceNamespace,
		Name:      deploymentSync.Spec.DeploymentName,
	}
	if err := r.Get(ctx, sourceDeploymentName, sourceDeployment); err != nil {
		return ctrl.Result{}, err
	}

	// Create or Update the destination Deployment in the target namespace
	destinationDeployment := &corev1.Deployment{}
	destinationDeploymentName := types.NamespacedName{
		Namespace: deploymentSync.Spec.DestinationNamespace,
		Name:      deploymentSync.Spec.DeploymentName,
	}
	if err := r.Get(ctx, destinationDeploymentName, destinationDeployment); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Deployment in destination namespace", "Namespace", deploymentSync.Spec.DestinationNamespace)
			destinationDeployment = &corev1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentSync.Spec.DeploymentName,
					Namespace: deploymentSync.Spec.DestinationNamespace,
				},
				Spec: sourceDeployment.Spec, // Copy spec from source to destination
			}
			if err := r.Create(ctx, destinationDeployment); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		log.Info("Updating Deployment in destination namespace", "Namespace", deploymentSync.Spec.DestinationNamespace)
		destinationDeployment.Spec = sourceDeployment.Spec // Update spec from source to destination
		if err := r.Update(ctx, destinationDeployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DeploymentSync{}).
		Complete(r)
}
