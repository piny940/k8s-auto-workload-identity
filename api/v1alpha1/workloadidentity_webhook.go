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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var workloadidentitylog = logf.Log.WithName("workloadidentity-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *WorkloadIdentity) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-k8s-piny940-com-v1alpha1-workloadidentity,mutating=true,failurePolicy=fail,sideEffects=None,groups=k8s.piny940.com,resources=workloadidentities,verbs=create;update,versions=v1alpha1,name=mworkloadidentity.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &WorkloadIdentity{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *WorkloadIdentity) Default() {
	workloadidentitylog.Info("default", "name", r.Name)

	if r.Spec.Provider.Namespace == "" {
		r.Spec.Provider.Namespace = r.Namespace
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-k8s-piny940-com-v1alpha1-workloadidentity,mutating=false,failurePolicy=fail,sideEffects=None,groups=k8s.piny940.com,resources=workloadidentities,verbs=create;update,versions=v1alpha1,name=vworkloadidentity.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &WorkloadIdentity{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *WorkloadIdentity) ValidateCreate() (admission.Warnings, error) {
	workloadidentitylog.Info("validate create", "name", r.Name)

	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *WorkloadIdentity) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	workloadidentitylog.Info("validate update", "name", r.Name)

	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *WorkloadIdentity) ValidateDelete() (admission.Warnings, error) {
	workloadidentitylog.Info("validate delete", "name", r.Name)

	return nil, nil
}

func (r *WorkloadIdentity) validate() (admission.Warnings, error) {
	if r.Spec.Deployment == "" {
		return nil, field.Invalid(field.NewPath("spec", "deployment"), r.Spec.Deployment, "deployment cannot be empty")
	}
	if r.Spec.TargetServiceAccount == "" {
		return nil, field.Invalid(field.NewPath("spec", "targetServiceAccount"), r.Spec.TargetServiceAccount, "targetServiceAccount cannot be empty")
	}
	if r.Spec.Provider.Name == "" {
		return nil, field.Invalid(field.NewPath("spec", "provider", "name"), r.Spec.Provider.Name, "provider name cannot be empty")
	}
	return nil, nil
}
