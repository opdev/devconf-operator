package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Autoscalers returns an HPAs based on specs.
func AutoScaler(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	metrics := []autoscalingv2.MetricSpec{}

	if recipe.Spec.Hpa.TargetMemoryUtilization != nil {
		memoryTarget := autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: recipe.Spec.Hpa.TargetMemoryUtilization,
				},
			},
		}
		metrics = append(metrics, memoryTarget)
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name + "-hpa",
			Namespace: recipe.Namespace,
			Labels:    recipe.Labels,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "devconfcz.opdev.com/v1alpha1",
				Kind:       "Recipe",
				Name:       recipe.Name,
			},
			MinReplicas: recipe.Spec.Hpa.MinReplicas,
			MaxReplicas: func(max *int32) int32 {
				if max == nil {
					return 0
				}
				return *max
			}(recipe.Spec.Hpa.MaxReplicas),
			Metrics: metrics,
		},
	}
	// Set the ownerRef for the HorizontalPodAutoScaler
	if err := ctrl.SetControllerReference(recipe, hpa, scheme); err != nil {
		return nil, err
	}
	return hpa, nil
}
