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
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var providerlog = logf.Log.WithName("provider-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *Provider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-k8s-piny940-com-v1alpha1-provider,mutating=true,failurePolicy=fail,sideEffects=None,groups=k8s.piny940.com,resources=providers,verbs=create;update,versions=v1alpha1,name=mprovider.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Provider{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Provider) Default() {
	providerlog.Info("default", "name", r.Name)

	if r.Spec.Location == "" {
		r.Spec.Location = "global"
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-k8s-piny940-com-v1alpha1-provider,mutating=false,failurePolicy=fail,sideEffects=None,groups=k8s.piny940.com,resources=providers,verbs=create;update,versions=v1alpha1,name=vprovider.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Provider{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Provider) ValidateCreate() (admission.Warnings, error) {
	providerlog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Provider) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	providerlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Provider) ValidateDelete() (admission.Warnings, error) {
	providerlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
