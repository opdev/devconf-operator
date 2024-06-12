package resources

import (
	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MySQLConfigMapForrecipe creates a ConfigMap for MySQL configuration
func MySQLConfigMapForrecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-config",
			Namespace: recipe.Namespace,
		},
		Data: map[string]string{
			"DB_HOST":             "mysql",
			"DB_PORT":             "3306",
			"MYSQL_DATABASE":      "recipes",
			"MYSQL_USER":          "recipeuser",
			"MYSQL_PASSWORD":      "recipepassword",
			"MYSQL_ROOT_PASSWORD": "rootpassword",
		},
	}

	if err := ctrl.SetControllerReference(recipe, configMap, scheme); err != nil {
		return nil, err
	}

	return configMap, nil
}

// MySQLInitDBConfigMapForrecipe creates a ConfigMap for MySQL initialization
func MySQLInitDBConfigMapForrecipe(recipe *devconfczv1alpha1.Recipe, scheme *runtime.Scheme) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysql-initdb-config",
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
