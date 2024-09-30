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

package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Provider Webhook", func() {

	Context("When creating Provider under Defaulting Webhook", func() {
		It("Should fill in the default value if a required field is empty", func() {
			provider := &Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name: "provider-1",
				},
				Spec: ProviderSpec{
					Target:     "gcp",
					PoolID:     "pool-1",
					ProviderID: "gcp-provider-1",
					Project: Project{
						Name:   "my-project",
						Number: "12345",
					},
				},
			}
			provider.Default()
			Expect(provider.Spec.Location).To(Equal("global"))
		})
	})

	Context("When creating Provider under Validating Webhook", func() {
		DescribeTable("Should deny if a required field is empty",
			func(provider *Provider) {
				_, err := provider.ValidateCreate()
				Expect(err).To(HaveOccurred())
			},
			Entry("Empty Target", &Provider{}),
			Entry("Empty PoolID", &Provider{
				Spec: ProviderSpec{
					Target:     "gcp",
					ProviderID: "gcp-provider-1",
					Project: Project{
						Name:   "my-project",
						Number: "12345",
					},
				},
			}),
			Entry("Empty ProviderID", &Provider{
				Spec: ProviderSpec{
					Target: "gcp",
					PoolID: "pool-1",
					Project: Project{
						Name:   "my-project",
						Number: "12345",
					},
				},
			}),
			Entry("Empty Project", &Provider{
				Spec: ProviderSpec{
					Target:     "gcp",
					PoolID:     "pool-1",
					ProviderID: "gcp-provider-1",
				},
			}),
			Entry("Empty Project Name", &Provider{
				Spec: ProviderSpec{
					Target:     "gcp",
					PoolID:     "pool-1",
					ProviderID: "gcp-provider-1",
					Project: Project{
						Number: "12345",
					},
				},
			}),
			Entry("Empty Project Number", &Provider{
				Spec: ProviderSpec{
					Target:     "gcp",
					PoolID:     "pool-1",
					ProviderID: "gcp-provider-1",
					Project: Project{
						Name: "my-project",
					},
				},
			}),
		)

		It("Should admit if all required fields are provided", func() {
			provider := &Provider{
				Spec: ProviderSpec{
					Target:     "gcp",
					PoolID:     "pool-1",
					ProviderID: "gcp-provider-1",
					Project: Project{
						Name:   "my-project",
						Number: "12345",
					},
				},
			}
			warns, err := provider.ValidateCreate()
			Expect(err).NotTo(HaveOccurred())
			Expect(warns).To(BeNil())
		})
	})

})
