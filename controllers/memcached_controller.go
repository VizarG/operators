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
	"fmt"
	_ "reflect"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	_ "k8s.io/apimachinery/pkg/types"
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
	// _ = r.Log.WithValues("memcached", req.NamespacedName)
	log := r.Log.WithValues("memcached", req.NamespacedName)
	logf.Log.WithName("start reconcile")

	// Watch for changes to primary resource WordPress
	memcached := &examplecomv1alpha1.Memcached{}
	err := r.Get(ctx, req.NamespacedName, memcached)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Memcached resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Memcached")
		return ctrl.Result{}, err
	}

	var result *ctrl.Result

	//						POSTGRES
	// Define a new deployment of postgresql
	result, err = r.ensureDeployment(req, memcached, r.postgresqlDeployment(memcached))
	if result != nil {
		return *result, err
	}
	//  Define new postgresql service
	result, err = r.ensureService(req, memcached, r.PostgresqlService(memcached))
	if result != nil {
		return *result, err
	}
	//  Define new postgresql secret
	result, err = r.ensureSecret(req, memcached, r.postgresqlSecret(memcached))
	if result != nil {
		return *result, err
	}

	// check is postgresql is runnning
	postgresqlRunning := r.isPostgresqlUp(memcached)

	for postgresqlRunning != true {
		if postgresqlRunning {
			log.Info("Postgresql is running now")
		} else {
			time.Sleep(3 * time.Second)
			postgresqlRunning = r.isPostgresqlUp(memcached)
			log.Info(fmt.Sprintf("Postgresql isn't running, state is %v", postgresqlRunning))
		}
	}

	//					NGINX
	// Define a new deployment of nginx
	result, err = r.ensureDeployment(req, memcached, r.deploymentForNginx(memcached))
	if result != nil {
		return *result, err
	}
	//  Define new frontend service
	result, err = r.ensureService(req, memcached, r.frontendService(memcached))
	if result != nil {
		return *result, err
	}
	//  Define new nginx config
	result, err = r.ensureConfig(req, memcached, r.frontenConfigName(memcached))
	if result != nil {
		return *result, err
	}

	//  check number of replicas of nginx
	result, err = r.checkSize(memcached)
	if result != nil {
		return *result, err
	}

	//				PGADMIN
	// Define a new deployment of pgadmin4
	result, err = r.ensureDeployment(req, memcached, r.pgadminDeploymentName(memcached))
	if result != nil {
		return *result, err
	}
	//  Define new pgadmin secret
	result, err = r.ensureSecret(req, memcached, r.pgadminSecret(memcached))
	if result != nil {
		return *result, err
	}
	//  Define new pgadmin service
	result, err = r.ensureService(req, memcached, r.pgadminServiceName(memcached))
	if result != nil {
		return *result, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplecomv1alpha1.Memcached{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}

func labels(m *examplecomv1alpha1.Memcached, name string) map[string]string {
	return map[string]string{
		"app":         m.Spec.Foo,
		"part-of-app": name}
}

//
// func test(t interface{}) (r *MemcachedReconciler) {
// 	switch reflect.TypeOf(t).Kind() {
// 	case reflect.Slice:
// 		s := reflect.ValueOf(t)
//
// 		for i := 0; i < s.Len(); i++ {
//
// 			// create all resources
// 			log.Info("Creating a new Deployment", "Deployment.Namespace", "Deployment.Name")
// 			// err = r.Create(ctx, &&v)
// 			if err != nil {
// 				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", "Deployment.Name")
// 				return ctrl.Result{}, err
// 			}
//
// 			fmt.Println(s.Index(i))
// 		}
// 	}
// 	return i
// }
