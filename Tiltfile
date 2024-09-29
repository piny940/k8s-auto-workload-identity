load('ext://cert_manager', 'deploy_cert_manager')
load('ext://kubebuilder', 'kubebuilder') 

deploy_cert_manager(version="v1.15.3")
kubebuilder("piny940.com", "k8s", "v1alpha1", "Provider")
