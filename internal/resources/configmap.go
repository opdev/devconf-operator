package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MySQLConfigMapForRecipe creates a ConfigMap for MySQL configuration
func MySQLConfigMapForRecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name + "-mysql-config",
			Namespace: recipe.Namespace,
		},
		Data: map[string]string{
			"DB_HOST":        "mysql",
			"DB_PORT":        "3306",
			"MYSQL_DATABASE": "recipes",
			"MYSQL_USER":     "recipeuser",
		},
	}

	if err := ctrl.SetControllerReference(recipe, configMap, scheme); err != nil {
		return nil, err
	}

	return configMap, nil
}

// MySQLInitDBConfigMapForRecipe creates a ConfigMap for MySQL initialization
func MySQLInitDBConfigMapForRecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      recipe.Name + "-mysql-initdb-config",
			Namespace: recipe.Namespace,
		},
		Data: map[string]string{
			"initdb.sql": `
				CREATE USER IF NOT EXISTS 'recipeuser'@'%' IDENTIFIED BY 'recipepassword';
				GRANT ALL PRIVILEGES ON recipes.* TO 'recipeuser'@'%';
				FLUSH PRIVILEGES;`,
		},
	}

	if err := ctrl.SetControllerReference(recipe, configMap, scheme); err != nil {
		return nil, err
	}

	return configMap, nil
}
