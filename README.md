## Deploying Custom Kubernetes Operator
In this project, we take a deep dive into the world of Controllers, Operators and Custom Resource Definitions(CRDs) to enable us create a custom Kubernetes operator. We would be making use of Kubebuilder, a framework for building Kubernetes API using CRDs to create the foundational structure for our operator, including a sample API and controller. Then, we will define our custom resource, generate the custom resource definition, setup the controller logic, build and deploy the custom operator, "DeploymentSync". This operator is designed to seamlessly synchronize deployments across namespaces within a cluster.

## What are Kubernetes Operators?
Kubernetes Operators entails design patterns for extending Kubernetes functionality to manage complex, stateful applications. They automate the deployment, scaling, and management of applications on Kubernetes by using custom controllers and custom resources. An operator essentially adds an endpoint to the Kubernetes API, called a custom resource(CR), along with a control plane component (controller) that monitors and maintains resources of the new type. In other words, Operators are software extensions that use custom resources to manage applications and their components.

- Custom Resources (CRs): Custom Resources are extensions of the Kubernetes API that allow users to create and manage their own resource types.
- Custom Controllers: Custom Controllers watch and manage the lifecycle of Custom Resources. They implement the control logic needed to maintain the desired state of the application, including deployment, scaling, and upgrades.

## Operator Lifecycle

![image](https://github.com/user-attachments/assets/2c275066-0846-4099-8786-564c4ca31805)

The operator’s role is to reconcile the actual state of the application with the desired state by the CRD using a control loop in which it can automatically scale, update, or restart the application. CRDs are used to extend Kubernetes by introducing new types of resources that are not part of the core Kubernetes API. By defining a CRD, users or operators can create their own custom resources and define how those resources should be managed.

## Why use Kubernetes Operators?
Kubernetes primitives by default is not built to manage state by default. Managing stateful application requirements such as upgrade, backup/recovery, failover automation hence becomes difficult. Operator pattern helps us to solve this issue using domain specific knowledge and deeclarative state. The domain specific knowledge is captured in code and exposed using a declarative API. Operators allow developers to define custom resources and the associated controllers to manage those resources. This allows for greater productivity and consistency in managing complex applications and infrastructure, as well as easier automation and better control over resources.
Prominent example is the Prometheus Operator for Kubernetes which provides easy monitoring definitions for Kubernetes services, deployment and management of Prometheus instances.

## Step 1: Installing Kubebuilder and creating a project
Firstly, we install [kubebuilder](https://book.kubebuilder.io/quick-start.html#installation). Next, we create a directory to setup our kubernetes operator project with a sample API and controller using the commands below.
```
mkdir my-operator
cd my-operator
go mod init example.com/my-operator
kubebuilder init --domain test.com
kubebuilder create api --group=apps --version=v1 --kind=DeploymentSync
```
![Screenshot (1022)](https://github.com/user-attachments/assets/805847b9-62c1-4ea3-9318-b1c08931f355)

![Screenshot (1024)](https://github.com/user-attachments/assets/34f1447f-c523-42e6-89a8-d1465465af3f)

## Step 2: Define Custom Resource
Here, we define our custom resource by editing the api/v1alpha1/deploymentsync_types.go file and modifying the DeploymentSyncSpec and DeploymentSyncStatus structs to define the desired fields and status of your custom resource. Below is the implemented custom resource

```
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentSyncSpec defines the desired state of DeploymentSync
type DeploymentSyncSpec struct {
    SourceNamespace      string `json:"sourceNamespace"`
    DestinationNamespace string `json:"destinationNamespace"`
    DeploymentName        string `json:"deploymentName"`
}
// DeploymentSyncStatus defines the observed state of DeploymentSync
type DeploymentSyncStatus struct {
    LastSyncTime metav1.Time `json:"lastSyncTime"`
}
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// DeploymentSync is the Schema for the deploymentsyncs API
type DeploymentSync struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   DeploymentSyncSpec   `json:"spec,omitempty"`
    Status DeploymentSyncStatus `json:"status,omitempty"`
}
// +kubebuilder:object:root=true
// DeploymentSyncList contains a list of DeploymentSync
type DeploymentSyncList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []DeploymentSync `json:"items"`
}
func init() {
    SchemeBuilder.Register(&DeploymentSync{}, &DeploymentSyncList{})
}
```



Then, we generate the CRD for the custom resource using the command below:
```
make manifests
```
![Screenshot (1025)](https://github.com/user-attachments/assets/f2bf9b9b-fd86-4ef7-9d03-8d87a1c43345)

![Screenshot (1026)](https://github.com/user-attachments/assets/fc63fee9-2929-4d3c-ac58-ba21d4c75ded)

## Step 3: Create Controller Logic
Next, we implement the reconciliation loop in the controller. We do this by editing the controllers/deploymentsync_controller.go file and inputting the logic (below) for the controller. This will typically involve watching for changes to our custom resource, reconciling the desired state with the actual state, and updating the status of your resource as necessary.
```
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
```

![Screenshot (1027)](https://github.com/user-attachments/assets/4519b66f-fb22-467a-9ceb-e7470c4676e5)

## Step 4: Building and Deploying the Operator
Next, we build and deploy our operator using the command below. This will build a Docker image for the operator, push it to dockerhub registry, and deploy it to your Kubernetes cluster.
```
make docker-build docker-push IMG=kenchuks44/deploymentsync-operator:1.0
make deploy IMG=example/deploymentsync-operator:1.0
```

![Screenshot (1029)](https://github.com/user-attachments/assets/6dfd8393-671c-4f89-a054-cc68c06b8c80)

![Screenshot (1030)](https://github.com/user-attachments/assets/7ddfaf78-7e69-4b6a-a568-772df204ee2c)

![Screenshot (1032)](https://github.com/user-attachments/assets/070da604-db7d-4378-aff7-86f239296201)

![Screenshot (1034)](https://github.com/user-attachments/assets/fc942506-504a-46d7-a274-7f03123da3eb)

![Screenshot (1035)](https://github.com/user-attachments/assets/d3d93427-0352-4f98-aa2f-054f4c5f0d56)

![Screenshot (1036)](https://github.com/user-attachments/assets/f6e358d6-e0c1-40b1-9b0d-ae8981d322f0)

![image](https://github.com/user-attachments/assets/c2286ea2-1063-4c75-8141-961a9971620e)

![Screenshot (1038)](https://github.com/user-attachments/assets/8f5aeca7-29c9-4a34-a077-680ed06c82d6)

![Screenshot (1039)](https://github.com/user-attachments/assets/47552d5d-6eff-4c49-9c46-b1e9c23d340a)

## Step 5: Testing and Validating Operator
We now create a sample manifest based on new opeator CRD to the cluster as below:
```
apiVersion: apps.test.com/v1
kind: DeploymentSync
metadata:
  labels:
    app.kubernetes.io/name: deploymentsync
    app.kubernetes.io/instance: deploymentsync-test
    app.kubernetes.io/part-of: deploymentsync
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: deploymentsync
  name: deploymentsync-test
spec:
  SourceNamespace: "default"
  DestinationNamespace: "deploymentsync-ns"
  DeploymentName: "nginx-deployment"
```

![Screenshot (1043)](https://github.com/user-attachments/assets/667b7e37-a272-45cc-b763-ee29c1b9f4da)

From the screenshot above, we can see the custom resource deployed but no deployment observed yet in the destination namespace, "deploymentsync-ns" where our deployments are synced to.

Now, we create a test nginx deployment in the source namespace, "default" with the manifest file below:
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
```

![image](https://github.com/user-attachments/assets/3383c93c-6c8f-4579-8a5e-5c1908c1b5bc)

![Screenshot (1049)](https://github.com/user-attachments/assets/9c0016a0-1537-4a97-aa43-8789f0838677)

![Screenshot (1046)](https://github.com/user-attachments/assets/61445d90-0b3c-440a-9729-62124eb5af80)

Additionally, ensure that the service account used by the operator has the necessary RBAC permissions to list, get, watch, create, update, patch, and delete deployments across the cluster. This will allow the operator to perform its intended syncing operations.

We have now been able to explore how the DeploymentSync operator simplifies the management of Deployments, facilitating their synchronization across namespaces within a Kubernetes cluster. 














