package resources

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
)

// PersistentVolumeClaimForKarbanatek creates a PVC for MySQL and sets the owner reference
func PersistentVolumeClaimForKarbanatek(karbanatek *devconfczv1alpha1.Karbanatek, scheme *runtime.Scheme) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql",
			Namespace: karbanatek.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(karbanatek, pvc, scheme); err != nil {
		return nil, err
	}

	return pvc, nil
}
