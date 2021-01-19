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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	examplecomv1alpha1 "github.com/memcached-operator/api/v1alpha1"
)

// pgadmin service
func pgadminServiceName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-pgadmin-service"
}

func (r *MemcachedReconciler) pgadminServiceName(m *examplecomv1alpha1.Memcached) *corev1.Service {
	// var ls = map[string]string{"app": "pgadmin"}
	ls := labels("pgadmin")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgadminServiceName(m),
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(80),
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	log.Info("Service Spec", "Service.Name", s.ObjectMeta.Name)

	ctrl.SetControllerReference(m, s, r.Scheme)
	return s
}

// pgadmin secret
func pgadminSecretName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-pgadmin-secret"
}

func (r *MemcachedReconciler) pgadminSecret(m *examplecomv1alpha1.Memcached) *corev1.Secret {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgadminSecretName(m),
			Namespace: m.Namespace,
		},
		Type: "Opaque",
		StringData: map[string]string{
			"username": "user@domain.local",
			"password": "postgres",
		},
	}
	ctrl.SetControllerReference(m, s, r.Scheme)
	return s
}

// pgadmin Deployment
func pgadminDeploymentName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-pgadmin"
}

func (r *MemcachedReconciler) pgadminDeploymentName(m *examplecomv1alpha1.Memcached) *appsv1.Deployment {
	ls := labels("pgadmin")
	replicas := m.Spec.Size

	userSecret := corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: pgadminSecretName(m),
			},
			Key: "username",
		},
	}

	passwordSecret := corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: pgadminSecretName(m),
			},
			Key: "password",
		},
	}

	deppgadmin := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgadminDeploymentName(m),
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "dpage/pgadmin4",
						Name:  "pgadmin",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "pgadmin",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "POSTGRESQL_DATABASE",
								Value: "test",
							},
							{
								Name:      "PGADMIN_DEFAULT_EMAIL",
								ValueFrom: &userSecret,
							},
							{
								Name:      "PGADMIN_DEFAULT_PASSWORD",
								ValueFrom: &passwordSecret,
							},
						},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(m, deppgadmin, r.Scheme)
	return deppgadmin
}
