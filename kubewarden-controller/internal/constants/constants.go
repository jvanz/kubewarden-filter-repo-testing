package constants

import "time"

const (
	// DefaultPolicyServer is the default policy server name to be used when
	// policies does not have a policy server name defined.
	DefaultPolicyServer = "default"
	// PolicyServer CA Secret.
	PolicyServerTLSCert                  = "policy-server-cert"
	PolicyServerTLSKey                   = "policy-server-key"
	PolicyServerCARootSecretName         = "policy-server-root-ca"
	PolicyServerCARootPemName            = "policy-server-root-ca-pem"
	PolicyServerCARootCACert             = "policy-server-root-ca-cert"
	PolicyServerCARootPrivateKeyCertName = "policy-server-root-ca-privatekey-cert"

	// PolicyServer Deployment.
	PolicyServerEnableMetricsEnvVar                 = "KUBEWARDEN_ENABLE_METRICS"
	PolicyServerDeploymentConfigVersionAnnotation   = "kubewarden/config-version"
	PolicyServerDeploymentPodSpecConfigVersionLabel = "kubewarden/config-version"
	PolicyServerPort                                = 8443
	PolicyServerMetricsPortEnvVar                   = "KUBEWARDEN_POLICY_SERVER_SERVICES_METRICS_PORT"
	PolicyServerMetricsPort                         = 8080
	PolicyServerReadinessProbe                      = "/readiness"
	PolicyServerLogFmtEnvVar                        = "KUBEWARDEN_LOG_FMT"

	// PolicyServer ConfigMap.
	PolicyServerConfigPoliciesEntry         = "policies.yml"
	PolicyServerDeploymentRestartAnnotation = "kubectl.kubernetes.io/restartedAt"
	PolicyServerConfigSourcesEntry          = "sources.yml"
	PolicyServerSourcesConfigContainerPath  = "/sources"

	// PolicyServer VerificationSecret.
	PolicyServerVerificationConfigEntry         = "verification-config"
	PolicyServerVerificationConfigContainerPath = "/verification"

	// Label.
	AppLabelKey          = "app"
	PolicyServerLabelKey = "kubewarden/policy-server"

	// Index.
	PolicyServerIndexKey = ".spec.policyServer"

	// Finalizers.
	KubewardenFinalizerPre114 = "kubewarden"
	KubewardenFinalizer       = "kubewarden.io/finalizer"

	// Kubernetes.
	KubernetesRevisionAnnotation = "deployment.kubernetes.io/revision"

	// OPTEL.
	OptelInjectAnnotation = "sidecar.opentelemetry.io/inject"

	// Webhook Configurations.
	WebhookConfigurationPolicyScopeLabelKey          = "kubewardenPolicyScope"
	WebhookConfigurationPolicyNameAnnotationKey      = "kubewardenPolicyName"
	WebhookConfigurationPolicyNamespaceAnnotationKey = "kubewardenPolicyNamespace"

	// Scope.
	NamespacePolicyScope = "namespace"
	ClusterPolicyScope   = "cluster"

	// Duration to be used when a policy should be reconciliation should be
	/// requeued.
	TimeToRequeuePolicyReconciliation = 2 * time.Second
	MetricsShutdownTimeout            = 5 * time.Second

	// Certs.
	CertExpirationYears = 10
)
