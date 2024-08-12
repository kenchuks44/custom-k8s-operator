package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "example.com/my-operator/api/v1"
)

// DeploymentSyncReconciler reconciles a DeploymentSync object
type DeploymentSyncReconciler struct {
        client.Client
        Scheme *runtime.Scheme
}

// Reconcile method to sync Deployment
func (r *DeploymentSyncReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
    ctx := context.Background()
    log := r.Log.WithValues("deploymentsync", req.NamespacedName)
// Fetch the DeploymentSync instance
    deploymentSync := &appsv1.DeploymentSync{}
    if err := r.Get(ctx, req.NamespacedName, deploymentSync); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    // Fetch the source Deployment
    sourceDeployment := &corev1.Deployment{}
    sourceDeployment := types.NamespacedName{
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
                Data: sourceDeployment.Data, // Copy data from source to destination
            }
            if err := r.Create(ctx, destinationDeployment); err != nil {
                return ctrl.Result{}, err
            }
        } else {
            return ctrl.Result{}, err
        }
    } else {
        log.Info("Updating Deployment in destination namespace", "Namespace", deploymentSync.Spec.DestinationNamespace)
        destinationDeployment.Data = sourceDeployment.Data // Update data from source to destination
        if err := r.Update(ctx, destinationDeployment); err != nil {
            return ctrl.Result{}, err
        }
    }
    return ctrl.Result{}, nil
}