package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func DeploymentForRecipeApp(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*appsv1.Deployment, error) {

	replicas := recipe.Spec.Replicas
	version := recipe.Spec.Version
	image := "quay.io/opdev/recipe_app:" + version

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name,
			Namespace: recipe.Namespace,
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
						Image:           image,
						Name:            "recipe-app",
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
							// WARNING: Ensure that the image used defines an UserID in the Dockerfile
							// otherwise the Pod will not run and will fail with `container has runAsNonRoot and image has non-numeric user`.
							// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
							// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the `RunAsNonRoot` and
							// RunAsUser fields empty.
							RunAsNonRoot:             &[]bool{true}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
						Resources: recipe.Spec.Resources,
					}},
				},
			},
		},
	}
	// Set the ownerRef for the Deployment
	if err := ctrl.SetControllerReference(recipe, dep, scheme); err != nil {
		return nil, err
	}
	return dep, nil
}
