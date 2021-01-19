/*
Copyright 2021.

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
	// "time"
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	// "k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	examplecomv1alpha1 "github.com/memcached-operator/api/v1alpha1"
)

var log = logf.Log.WithName("controller_wordpress")

// MemcachedReconciler reconciles a Memcached object
type MemcachedReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=memcacheds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=memcacheds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Memcached object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete

// func frontendIngressName(v *examplecomv1alpha1.Memcached) string {
// 	return v.Spec.Foo + "-frontend-ingress"
// }

func (r *MemcachedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = context.Background()
	_ = r.Log.WithValues("memcached", req.NamespacedName)
	logf.Log.WithName("start reconcile")

	// Watch for changes to primary resource WordPress
	memcached := &examplecomv1alpha1.Memcached{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, memcached)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logf.Log.WithName("there is no errors")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logf.Log.WithName("Failed to get Memcached")
		return ctrl.Result{}, nil
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment of nginx
		dep := r.deploymentForNginx(memcached)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}

		// Define a new deployment of postgresql
		depPostgres := r.postgresqlDeployment(memcached)
		log.Info("Creating a new Deployment", "Deployment.Namespace", depPostgres.Namespace, "Deployment.Name", depPostgres.Name)
		err = r.Create(ctx, depPostgres)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", depPostgres.Namespace, "Deployment.Name", depPostgres.Name)
			return ctrl.Result{}, err
		}

		// Define a new deployment of pgadmin4
		depPgamin := r.pgadminDeploymentName(memcached)
		log.Info("Creating a new Deployment", "Deployment.Namespace", depPgamin.Namespace, "Deployment.Name", depPgamin.Name)
		err = r.Create(ctx, depPgamin)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", depPgamin.Namespace, "Deployment.Name", depPgamin.Name)
			return ctrl.Result{}, err
		}

		//  Define new frontend service
		servfront := r.frontendService(memcached)
		log.Info("Creating a new Deployment", "Deployment.Namespace", servfront.Namespace, "Deployment.Name", servfront.Name)
		err = r.Create(ctx, servfront)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", servfront.Namespace, "Deployment.Name", servfront.Name)
			return ctrl.Result{}, err
		}

		//  Define new nginx config
		ConfigMapFront := r.frontenConfigName(memcached)
		log.Info("Creating a new ConfigMap", "Deployment.Namespace", ConfigMapFront.Namespace, "Deployment.Name", ConfigMapFront.Name)
		err = r.Create(ctx, ConfigMapFront)
		if err != nil {
			log.Error(err, "Failed to create new ConfigMap for nginx", "Deployment.Namespace", ConfigMapFront.Namespace, "Deployment.Name", ConfigMapFront.Name)
			return ctrl.Result{}, err
		}

		//  Define new pgadmin service
		servpgadmin := r.pgadminServiceName(memcached)
		log.Info("Creating a new Deployment", "Deployment.Namespace", servpgadmin.Namespace, "Deployment.Name", servpgadmin.Name)
		err = r.Create(ctx, servpgadmin)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", servpgadmin.Namespace, "Deployment.Name", servpgadmin.Name)
			return ctrl.Result{}, err
		}

		//  Define new pgadmin secret
		secretpgadmin := r.postgresqlSecret(memcached)
		log.Info("Creating a new secretpgadmin", "Deployment.Namespace", secretpgadmin.Namespace, "Deployment.Name", secretpgadmin.Name)
		err = r.Create(ctx, secretpgadmin)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", secretpgadmin.Namespace, "Deployment.Name", secretpgadmin.Name)
			return ctrl.Result{}, err
		}

		//  Define new postgresql secret
		secretpostgres := r.pgadminSecret(memcached)
		log.Info("Creating a new secretpgadmin", "Deployment.Namespace", secretpostgres.Namespace, "Deployment.Name", secretpostgres.Name)
		err = r.Create(ctx, secretpostgres)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", secretpostgres.Namespace, "Deployment.Name", secretpostgres.Name)
			return ctrl.Result{}, err
		}

		//  Define new postgresql service
		servpostgres := r.PostgresqlService(memcached)
		log.Info("Creating a new Deployment", "Deployment.Namespace", servpostgres.Namespace, "Deployment.Name", servpostgres.Name)
		err = r.Create(ctx, servpostgres)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", servpostgres.Namespace, "Deployment.Name", servpostgres.Name)
			return ctrl.Result{}, err
		}

		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := memcached.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(memcached.Namespace),
		client.MatchingLabels(labels(memcached.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Memcached.Namespace", memcached.Namespace, "Memcached.Name", memcached.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, memcached.Status.Nodes) {
		memcached.Status.Nodes = podNames
		err := r.Status().Update(ctx, memcached)
		if err != nil {
			log.Error(err, "Failed to update Memcached status")
			return ctrl.Result{}, err
		}
	}

	// // Check if the service already exists
	// found_service := &corev1.Service{}
	// err = r.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}, found_service)
	// if err != nil && errors.IsNotFound(err) {
	// 	// Define a new deployment
	// 	dep := r.frontendService(memcached)
	// 	log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 	err = r.Create(ctx, dep)
	// 	if err != nil {
	// 		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	// 		return ctrl.Result{}, err
	// 	}
	// 	// Deployment created successfully - return and requeue
	// 	return ctrl.Result{Requeue: true}, nil
	// } else if err != nil {
	// 	log.Error(err, "Failed to get Deployment")
	// 	return ctrl.Result{}, err
	// }

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplecomv1alpha1.Memcached{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}

// labels returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labels(name string) map[string]string {
	return map[string]string{"app": "memcached", "memcached_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
