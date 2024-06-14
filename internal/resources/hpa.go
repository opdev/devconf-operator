package resources

import (
	
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	defaultMemoryTarget = int32(60)
)

// Autoscalers returns an HPAs based on specs.
func AutoScaler(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	targetMemoryUtilization := defaultMemoryTarget
	memoryTarget := autoscalingv2.ResourceMetricSource{
		Name: "memory",
		Target: autoscalingv2.MetricTarget{
			Type:               autoscalingv2.UtilizationMetricType,
			AverageUtilization: &targetMemoryUtilization,
		},
	}
	targetMetrics := []autoscalingv2.MetricSpec{
		{
			Type:     "Resource",
			Resource: &memoryTarget,
		},
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
			MaxReplicas: *recipe.Spec.Hpa.MaxReplicas,
			Metrics:     targetMetrics,
		},
	}
	// Set the ownerRef for the HorizontalPodAutoScaler
	if err := ctrl.SetControllerReference(recipe, hpa, scheme); err != nil {
		return nil, err
	}
	return hpa, nil
}