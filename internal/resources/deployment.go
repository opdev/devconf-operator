package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
)

// KarbanatekReconciler reconciles a Karbanatek object
type KarbanatekReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// deploymentForKarbanatek returns a Karbanatek Deployment object
func (r *KarbanatekReconciler) deploymentForKarbanatek(
	karbanatek *devconfczv1alpha1.Karbanatek) (*appsv1.Deployment, error) {
	replicas := karbanatek.Spec.Count

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      karbanatek.Name,
			Namespace: karbanatek.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "recipe-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "recipe-app",
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{{
						Image:           "quay.io/rocrisp/recipe:v13",
						Name:            "karbanatek",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 5000,
								Name:          "http",
							},
						},
						Env: []corev1.EnvVar{
							{
								Name: "DB_HOST",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "mysql-config",
										},
										Key: "DB_HOST",
									},
								},
							},
							{
								Name: "DB_PORT",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "mysql-config",
										},
										Key: "DB_PORT",
									},
								},
							},
							{
								Name: "DB_NAME",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "mysql-config",
										},
										Key: "MYSQL_DATABASE",
									},
								},
							},
							{
								Name:  "DB_USER",
								Value: "recipeuser",
							},
							{
								Name: "DB_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "mysql-config",
										},
										Key: "MYSQL_PASSWORD",
									},
								},
							},
						},
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:             &[]bool{true}[0],
							RunAsUser:                &[]int64{1001}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
					}},
				},
			},
		},
	}
	// Set the ownerRef for the Deployment
	if err := ctrl.SetControllerReference(karbanatek, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil
}
