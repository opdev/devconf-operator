package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MySQLSecretForRecipe creates a ConfigMap for MySQL configuration
func MySQLSecretForRecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name + "-mysql",
			Namespace: recipe.Namespace,
		},
		StringData: map[string]string{
			"MYSQL_PASSWORD":      "recipepassword",
			"MYSQL_ROOT_PASSWORD": "rootpassword",
		},
	}

	if err := ctrl.SetControllerReference(recipe, secret, scheme); err != nil {
		return nil, err
	}

	return secret, nil
}
