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
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sv1alpha1 "github.com/piny940/kwimount/api/v1alpha1"
)

func sampleDeployment(name, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container-1",
							Image: "test-image",
						},
						{
							Name:  "test-container-2",
							Image: "test-image",
						},
					},
				},
			},
		},
	}
}

var sampleProvider = k8sv1alpha1.Provider{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-provider",
		Namespace: "default",
	},
	Spec: k8sv1alpha1.ProviderSpec{
		Project: k8sv1alpha1.Project{
			Number: "test-project-number",
			Name:   "test-project-name",
		},
		Location:   "global",
		Target:     "gcp",
		PoolID:     "test-pool-id",
		ProviderID: "test-provider-id",
	},
}

var _ = Describe("WorkloadIdentity Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		targetNamespacedName := types.NamespacedName{
			Name:      "target-deployment",
			Namespace: typeNamespacedName.Namespace,
		}

		BeforeEach(func() {
			By("creating the custom resource for the Kind WorkloadIdentity")
			workloadidentity := &k8sv1alpha1.WorkloadIdentity{}
			err := k8sClient.Get(ctx, typeNamespacedName, workloadidentity)
			if err != nil && errors.IsNotFound(err) {
				resource := &k8sv1alpha1.WorkloadIdentity{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: typeNamespacedName.Namespace,
					},
					Spec: k8sv1alpha1.WorkloadIdentitySpec{
						Provider: k8sv1alpha1.WorkloadIdentityProvider{
							Name:      sampleProvider.Name,
							Namespace: "default",
						},
						TargetServiceAccount: "test-service-account",
						Deployment:           targetNamespacedName.Name,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				Expect(k8sClient.Create(ctx, &sampleProvider)).To(Succeed())
			}
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(&sampleProvider), &sampleProvider)
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, &sampleProvider)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &k8sv1alpha1.WorkloadIdentity{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance WorkloadIdentity")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			deps := &appsv1.DeploymentList{}
			label, err := labels.Parse("app=test-app")
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.List(ctx, deps, &client.ListOptions{LabelSelector: label})
			Expect(err).NotTo(HaveOccurred())
			for _, dep := range deps.Items {
				Expect(k8sClient.Delete(ctx, &dep)).To(Succeed())
			}

			confs := &corev1.ConfigMapList{}
			err = k8sClient.List(ctx, confs)
			Expect(err).NotTo(HaveOccurred())
			for _, conf := range confs.Items {
				Expect(k8sClient.Delete(ctx, &conf)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			var err error
			workloadidentity := &k8sv1alpha1.WorkloadIdentity{}
			err = k8sClient.Get(ctx, typeNamespacedName, workloadidentity)
			Expect(err).NotTo(HaveOccurred())
			controllerReconciler := &WorkloadIdentityReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			Expect(k8sClient.Create(ctx,
				sampleDeployment(targetNamespacedName.Name, targetNamespacedName.Namespace),
			)).To(Succeed())

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			{
				By("Checking the ConfigMap")
				cm := &corev1.ConfigMap{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      configMapName(workloadidentity),
					Namespace: workloadidentity.Namespace,
				}, cm)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).NotTo(BeNil())

				actual := make(map[string]interface{})
				fmt.Println(cm.Data[GCP_CONFIGURATION_FILE_NAME])
				err = json.Unmarshal([]byte(cm.Data[GCP_CONFIGURATION_FILE_NAME]), &actual)
				Expect(err).NotTo(HaveOccurred())
				expected := make(map[string]interface{})
				err = json.Unmarshal([]byte(fmt.Sprintf(GCP_CONF_BASE,
					sampleProvider.Spec.Project.Number,
					sampleProvider.Spec.Location,
					sampleProvider.Spec.PoolID,
					sampleProvider.Spec.ProviderID,
					GCP_TOKEN_MOUNT_PATH+GCP_TOKEN_PATH,
					workloadidentity.Spec.TargetServiceAccount,
				)), &expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(Equal(expected))
			}
			{
				By("Checking the Deployment")
				dep := &appsv1.Deployment{}
				err = k8sClient.Get(ctx, targetNamespacedName, dep)
				Expect(err).NotTo(HaveOccurred())
				for _, container := range dep.Spec.Template.Spec.Containers {
					Expect(container.Env).To(ContainElement(corev1.EnvVar{
						Name:  GOOGLE_CREDENTIALS_ENV,
						Value: GCP_CONFIGURATION_MOUNT_PATH + GCP_CONFIGURATION_FILE_NAME,
					}))
					containsTokenVolume := false
					for _, volume := range dep.Spec.Template.Spec.Volumes {
						if volume.Name == GCP_TOKEN_VOLUME_NAME {
							containsTokenVolume = true
						}
						if volume.Name == GCP_CONFIGURATION_FILE_NAME {
							Expect(volume.ConfigMap.Name).To(Equal(configMapName(workloadidentity)))
						}
					}
					Expect(containsTokenVolume).To(BeTrue())
				}
			}
		})
	})
})
