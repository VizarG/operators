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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	examplecomv1alpha1 "github.com/memcached-operator/api/v1alpha1"
)

const frontendPort = 80
const frontendServicePort = 80
const frontendImage = "nginx"

var nginx_cofnf string = `
upstream backend {
	server vizar-app-pgadmin-service:80;
}
server {
    listen 80;
    server_name vizar-app-pgadmin-service;
    root /usr/share/nginx/html;
    index index.html index.html;

    location / {
        proxy_pass http://vizar-app-pgadmin-service;
				proxy_redirect http://vizar-app-pgadmin-service http://adminer.k8s.com;
    }

}
`

func frontendIngressName(v *examplecomv1alpha1.Memcached) string {
	return v.Spec.Foo + "-frontend-ingress"
}

// ingress for nginx returns an ingress  object
func (r *MemcachedReconciler) frontendIngress(m *examplecomv1alpha1.Memcached) *corev1.LoadBalancerIngress {
	// ls := labels(m.Name)
	i := &corev1.LoadBalancerIngress{
		IP:       "192.168.49.3",
		Hostname: "k8s.adminer.com",
		// PortStatus {{
		// 	Protocol: TCP,
		// 	Port:     80,
		// }},
		// ObjectMeta: metav1.ObjectMeta{
		// 	Name:      frontendIngressName(m),
		// 	Namespace: m.Namespace,
		// },
		// Ports []PortStatus{{
		// 	Port: 80,
		// 	Protocol:   corev1.ProtocolTCP,
		// 	}},

	}
	// log.Info("Service Spec", "Service.Name", i.ObjectMeta.Name)
	// ctrl.SetControllerReference(m, i, r.Scheme)
	return i
}

// service for nginx

func frontendServiceName(v *examplecomv1alpha1.Memcached) string {
	return v.Spec.Foo + "-frontend-service"
}

func (r *MemcachedReconciler) frontendService(m *examplecomv1alpha1.Memcached) *corev1.Service {
	ls := labels(m, "nginx")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      frontendServiceName(m),
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       frontendPort,
				TargetPort: intstr.FromInt(frontendPort),
				// NodePort:   frontendServicePort,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	log.Info("Service Spec", "Service.Name", s.ObjectMeta.Name)

	ctrl.SetControllerReference(m, s, r.Scheme)
	return s
}

// config map nginx
func frontenConfigName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-frontend-config"
}

func (r *MemcachedReconciler) frontenConfigName(m *examplecomv1alpha1.Memcached) *corev1.ConfigMap {
	s := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      frontenConfigName(m),
			Namespace: m.Namespace,
		},
		Data: map[string]string{
			"default.conf": nginx_cofnf,
		},
	}

	log.Info("Service Spec", "Service.Name", s.ObjectMeta.Name)

	ctrl.SetControllerReference(m, s, r.Scheme)
	return s
}

// deployment nginx
func frontenDeploymentName(m *examplecomv1alpha1.Memcached) string {
	return m.Spec.Foo + "-frontend-deployment"
}

func (r *MemcachedReconciler) deploymentForNginx(m *examplecomv1alpha1.Memcached) *appsv1.Deployment {
	ls := labels(m, "nginx")
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      frontenDeploymentName(m),
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
						Image: frontendImage,
						Name:  "nginx",
						Ports: []corev1.ContainerPort{{
							ContainerPort: frontendServicePort,
							Name:          "nginx",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "nginx-conf",
							MountPath: "/etc/nginx/conf.d/",
							// SubPath:   "etc/nginx/conf.d/default.conf/",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "nginx-conf",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: frontenConfigName(m),
								},
							},
						},
					},
					}},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *MemcachedReconciler) checkSize(m *examplecomv1alpha1.Memcached) (*ctrl.Result, error) {

	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      frontenDeploymentName(m),
		Namespace: m.Namespace,
	}, found)

	size := m.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(context.TODO(), found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", m.Namespace, "Deployment.Name", m.Name)
			return &ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return &ctrl.Result{Requeue: true}, nil
	}

	return nil, nil

}
