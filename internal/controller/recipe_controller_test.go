/*
Copyright 2024.

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

package controller

import (
	"context"
	"os"
	"time"

	//nolint:golint
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	devconfczv1alpha1 "github.com/opdev/devconf-operator/api/v1alpha1"
)

var _ = Describe("Recipe controller", func() {
	Context("Recipe controller test", func() {

		const RecipeName = "test-recipe"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      RecipeName,
				Namespace: RecipeName,
			},
		}

		typeNamespaceName := types.NamespacedName{
			Name:      RecipeName,
			Namespace: RecipeName,
		}
		recipe := &devconfczv1alpha1.Recipe{}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("Setting the Image ENV VAR which stores the Operand image")
			err = os.Setenv("RECIPE_IMAGE", "example.com/image:test")
			Expect(err).To(Not(HaveOccurred()))

			By("creating the custom resource for the Kind Recipe")
			err = k8sClient.Get(ctx, typeNamespaceName, recipe)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				recipe := &devconfczv1alpha1.Recipe{
					ObjectMeta: metav1.ObjectMeta{
						Name:      RecipeName,
						Namespace: namespace.Name,
					},
					Spec: devconfczv1alpha1.RecipeSpec{
						Replicas: 1,
						Version:  "v13",
					},
				}

				err = k8sClient.Create(ctx, recipe)
				Expect(err).To(Not(HaveOccurred()))
			}
		})

		AfterEach(func() {
			By("removing the custom resource for the Kind Recipe")
			found := &devconfczv1alpha1.Recipe{}
			err := k8sClient.Get(ctx, typeNamespaceName, found)
			Expect(err).To(Not(HaveOccurred()))

			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), found)
			}, 2*time.Minute, time.Second).Should(Succeed())

			// TODO(user): Attention if you improve this code by adding other context test you MUST
			// be aware of the current delete namespace limitations.
			// More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)

			By("Removing the Image ENV VAR which stores the Operand image")
			_ = os.Unsetenv("RECIPE_IMAGE")
		})

		It("should successfully reconcile a custom resource for Recipe", func() {
			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &devconfczv1alpha1.Recipe{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			recipeReconciler := &RecipeReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := recipeReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if Deployment was successfully created in the reconciliation")
			Eventually(func() error {
				found := &appsv1.Deployment{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			// By("Checking the latest Status Condition added to the Recipe instance")
			// Eventually(func() error {
			// 	if recipe.Status.Conditions != nil &&
			// 		len(recipe.Status.Conditions) != 0 {
			// 		latestStatusCondition := recipe.Status.Conditions[len(recipe.Status.Conditions)-1]
			// 		expectedLatestStatusCondition := metav1.Condition{
			// 			Type:   typeAvailableRecipe,
			// 			Status: metav1.ConditionTrue,
			// 			Reason: "Reconciling",
			// 			Message: fmt.Sprintf(
			// 				"Deployment for custom resource (%s) with %d replicas created successfully",
			// 				recipe.Name,
			// 				recipe.Spec.Size),
			// 		}
			// 		if latestStatusCondition != expectedLatestStatusCondition {
			// 			return fmt.Errorf("The latest status condition added to the Recipe instance is not as expected")
			// 		}
			// 	}
			// 	return nil
			// }, time.Minute, time.Second).Should(Succeed())
		})
	})
})
