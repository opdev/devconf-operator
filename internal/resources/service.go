package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MySQLServiceForrecipe creates a Service for the MySQL Deployment and sets the owner reference
func MySQLServiceForrecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: recipe.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 3306,
				},
			},
			Selector: map[string]string{
				"app": "mysql",
			},
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(recipe, service, scheme); err != nil {
		return nil, err
	}

	return service, nil
}

// RecipeServiceForrecipe creates a Service for the Recipe application and sets the owner reference
func RecipeServiceForrecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "recipe-app-service",
			Namespace: recipe.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "recipe-app",
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(5000),
				},
			},
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(recipe, service, scheme); err != nil {
		return nil, err
	}

	return service, nil
}
