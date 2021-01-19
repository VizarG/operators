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

// postgresql secret
func postgresqlSecretName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-postgresql-secret"
}

func (r *MemcachedReconciler) postgresqlSecret(m *examplecomv1alpha1.Memcached) *corev1.Secret {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgresqlSecretName(m),
			Namespace: m.Namespace,
		},
		Type: "Opaque",
		StringData: map[string]string{
			"username": "postgres",
			"password": "root",
		},
	}
	ctrl.SetControllerReference(m, s, r.Scheme)
	return s
}

// postgresql service
func postgresqlServiceName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-postgresql-service"
}

func (r *MemcachedReconciler) PostgresqlService(m *examplecomv1alpha1.Memcached) *corev1.Service {
	ls := labels("postgres")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgresqlServiceName(m),
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       5432,
				TargetPort: intstr.FromInt(5432),
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	log.Info("Service Spec", "Service.Name", s.ObjectMeta.Name)

	ctrl.SetControllerReference(m, s, r.Scheme)
	return s
}

// postgresql volume
func postgresqlVolumeName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-postgresql-volume"
}

// posgresql pod
func postgresqlDeploymentName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-postgresql"
}

func (r *MemcachedReconciler) postgresqlDeployment(m *examplecomv1alpha1.Memcached) *appsv1.Deployment {
	ls := labels("postgres")
	replicas := m.Spec.Size

	userSecret := corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: postgresqlSecretName(m),
			},
			Key: "username",
		},
	}

	passwordSecret := corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: postgresqlSecretName(m),
			},
			Key: "password",
		},
	}

	deppostgres := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      postgresqlDeploymentName(m),
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
						Image: "postgres:latest",
						Name:  "postgresql-db",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 5432,
							Name:          "postgresql",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "POSTGRESQL_DATABASE",
								Value: "test",
							},
							{
								Name:      "POSTGRESQL_USER",
								ValueFrom: &userSecret,
							},
							{
								Name:      "POSTGRES_PASSWORD",
								ValueFrom: &passwordSecret,
							},
						},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(m, deppostgres, r.Scheme)
	return deppostgres
}
