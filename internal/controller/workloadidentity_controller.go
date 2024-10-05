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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sv1alpha1 "github.com/piny940/kwimount/api/v1alpha1"
)

// WorkloadIdentityReconciler reconciles a WorkloadIdentity object
type WorkloadIdentityReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	FIELD_MANAGER                = "kwimount"
	GCP_TOKEN_MOUNT_PATH         = "/var/run/kwimount-gcp-service-account/"
	GCP_TOKEN_PATH               = "token"
	GCP_CONFIGURATION_MOUNT_PATH = "/etc/kwimount-gcp-workload-identity/"
	GCP_TOKEN_VOLUME_NAME        = "kwimount-gcp-token"
	GCP_TOKEN_AUDIENCE           = "https://iam.googleapis.com/projects/%s/locations/%s/workloadIdentityPools/%s/providers/%s"
	TOKEN_EXPIRATION_SEC         = 3600
)

// +kubebuilder:rbac:groups=k8s.piny940.com,resources=workloadidentities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.piny940.com,resources=workloadidentities/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=k8s.piny940.com,resources=workloadidentities/finalizers,verbs=update
// +kubebuilder:rbac:groups=k8s.piny940.com,resources=providers,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WorkloadIdentity object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *WorkloadIdentityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var wi k8sv1alpha1.WorkloadIdentity
	err := r.Client.Get(ctx, req.NamespacedName, &wi)
	if err != nil {
		logger.Error(err, "unable to fetch WorkloadIdentity")
		return ctrl.Result{}, err
	}

	var provider k8sv1alpha1.Provider
	err = r.Client.Get(ctx, client.ObjectKey{
		Namespace: wi.Spec.Provider.Namespace,
		Name:      wi.Spec.Provider.Name,
	}, &provider)
	if err != nil {
		logger.Error(err, "unable to fetch Provider")
		return ctrl.Result{}, err
	}
	err = r.reconcileConfigMap(ctx, &wi, &provider)
	if err != nil {
		logger.Error(err, "unable to reconcile ConfigMap")
		return ctrl.Result{}, err
	}
	err = r.reconcileDeployment(ctx, &wi, &provider)
	if err != nil {
		logger.Error(err, "unable to reconcile Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WorkloadIdentityReconciler) reconcileConfigMap(ctx context.Context, wi *k8sv1alpha1.WorkloadIdentity, pr *k8sv1alpha1.Provider) error {
	logger := log.FromContext(ctx)

	cm := &corev1.ConfigMap{}
	cm.SetNamespace(wi.Namespace)
	cm.SetName(configMapName(wi))
	op, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			switch pr.Spec.Target {
			case k8sv1alpha1.ProviderTargetTypeGCP:
				cm.Data = gcpConfigMapData(wi, pr)
			default:
				err := fmt.Errorf("unsupported provider target type %s", pr.Spec.Target)
				logger.Error(err, "unable to createOrUpdate ConfigMap")
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "unable to createOrUpdate ConfigMap")
		return err
	}
	if op != controllerutil.OperationResultNone {
		logger.Info("successfully reconciled ConfigMap", "operation", op)
	}
	return nil
}

func gcpConfigMapData(wi *k8sv1alpha1.WorkloadIdentity, pr *k8sv1alpha1.Provider) map[string]string {
	return map[string]string{
		"data": fmt.Sprintf(`{
  "universe_domain": "googleapis.com",
  "type": "external_account",
  "audience": "//iam.googleapis.com/projects/%s/locations/%s/workloadIdentityPools/%s/providers/%s",
  "subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
  "token_url": "https://sts.googleapis.com/v1/token",
  "credential_source": {
    "file": %s,
    "format": {
      "type": "text"
    }
  },
  "service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken"
}`, pr.Spec.Project.Number,
			pr.Spec.Location,
			pr.Spec.PoolID,
			pr.Spec.ProviderID,
			GCP_TOKEN_MOUNT_PATH+GCP_TOKEN_PATH,
			wi.Spec.TargetServiceAccount,
		)}
}

func (r *WorkloadIdentityReconciler) reconcileDeployment(ctx context.Context, wi *k8sv1alpha1.WorkloadIdentity, pr *k8sv1alpha1.Provider) error {
	logger := log.FromContext(ctx)

	var current appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{
		Namespace: wi.Namespace,
		Name:      wi.Spec.Deployment,
	}, &current)
	if err != nil {
		logger.Error(err, "unable to fetch Deployment")
		return err
	}

	containers := make([]*corev1apply.ContainerApplyConfiguration, 0, len(current.Spec.Template.Spec.Containers))
	for _, container := range current.Spec.Template.Spec.Containers {
		containers = append(containers, corev1apply.Container().
			WithName(container.Name).
			WithVolumeMounts(
				corev1apply.VolumeMount().
					WithName(GCP_TOKEN_VOLUME_NAME).
					WithMountPath(GCP_TOKEN_MOUNT_PATH).
					WithReadOnly(true),
				corev1apply.VolumeMount().
					WithName(configMapName(wi)).
					WithMountPath(GCP_CONFIGURATION_MOUNT_PATH).
					WithReadOnly(true),
			))
	}
	audience := fmt.Sprintf(GCP_TOKEN_AUDIENCE, pr.Spec.Project.Number, pr.Spec.Location, pr.Spec.PoolID, pr.Spec.ProviderID)
	expected := appsv1apply.Deployment(wi.Spec.Deployment, wi.Namespace).
		WithSpec(appsv1apply.DeploymentSpec().
			WithTemplate(corev1apply.PodTemplateSpec().
				WithSpec(corev1apply.PodSpec().
					WithContainers(containers...).
					WithVolumes(
						corev1apply.Volume().
							WithName(configMapName(wi)).
							WithConfigMap(corev1apply.ConfigMapVolumeSource().
								WithName(configMapName(wi)),
							),
						corev1apply.Volume().
							WithName("kwimount-gcp-token").
							WithProjected(
								corev1apply.ProjectedVolumeSource().
									WithSources(corev1apply.VolumeProjection().
										WithServiceAccountToken(
											corev1apply.ServiceAccountTokenProjection().
												WithAudience(audience).
												WithExpirationSeconds(TOKEN_EXPIRATION_SEC).
												WithPath(GCP_TOKEN_PATH),
										),
									),
							),
					),
				),
			),
		)
	currentApply, err := appsv1apply.ExtractDeployment(&current, FIELD_MANAGER)
	if err != nil {
		logger.Error(err, "unable to extract current Deployment")
		return err
	}
	if equality.Semantic.DeepEqual(expected, currentApply) {
		return nil
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(expected)
	if err != nil {
		logger.Error(err, "unable to convert Deployment to unstructured")
		return err
	}
	patch := &unstructured.Unstructured{Object: obj}
	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: FIELD_MANAGER,
		Force:        pointer.Bool(true),
	})
	logger.Info("successfully patched Deployment with name: %s, namespace: %s", wi.Spec.Deployment, wi.Namespace)
	return nil
}

func configMapName(wi *k8sv1alpha1.WorkloadIdentity) string {
	return fmt.Sprintf("kwimount-%s-%s-conf", wi.Name, wi.Spec.Deployment)
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkloadIdentityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1alpha1.WorkloadIdentity{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
