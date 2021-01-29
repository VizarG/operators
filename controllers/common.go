package controllers

import (
	"context"

	examplecomv1alpha1 "github.com/memcached-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	// "sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MemcachedReconciler) ensureDeployment(request ctrl.Request,
	instance *examplecomv1alpha1.Memcached,
	dep *appsv1.Deployment,
) (*ctrl.Result, error) {

	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the deployment
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Client.Create(context.TODO(), dep)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &ctrl.Result{}, err
		} else {
			// Deployment was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get Deployment")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *MemcachedReconciler) ensureService(request ctrl.Request,
	instance *examplecomv1alpha1.Memcached,
	s *corev1.Service,
) (*ctrl.Result, error) {
	found := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		log.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		err = r.Client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *MemcachedReconciler) ensureSecret(request ctrl.Request,
	instance *examplecomv1alpha1.Memcached,
	s *corev1.Secret,
) (*ctrl.Result, error) {
	found := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the secret
		log.Info("Creating a new secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		err = r.Client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the secret not existing
		log.Error(err, "Failed to get Secret")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *MemcachedReconciler) ensureConfig(request ctrl.Request,
	instance *examplecomv1alpha1.Memcached,
	s *corev1.ConfigMap,
) (*ctrl.Result, error) {
	found := &corev1.ConfigMap{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the ConfigMap
		log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", s.Namespace, "ConfigMap.Name", s.Name)
		err = r.Client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", s.Namespace, "ConfigMap.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the secret not existing
		log.Error(err, "Failed to get ConfigMap")
		return &ctrl.Result{}, err
	}

	return nil, nil
}
