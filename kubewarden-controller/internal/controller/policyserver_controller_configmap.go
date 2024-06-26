package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	policiesv1 "github.com/kubewarden/kubewarden-controller/api/policies/v1"
	"github.com/kubewarden/kubewarden-controller/internal/constants"
)

const dataType string = "Data" // only data type is supported

type policyServerConfigEntry struct {
	NamespacedName        types.NamespacedName              `json:"namespacedName"`
	URL                   string                            `json:"url"`
	PolicyMode            string                            `json:"policyMode"`
	AllowedToMutate       bool                              `json:"allowedToMutate"`
	ContextAwareResources []policiesv1.ContextAwareResource `json:"contextAwareResources,omitempty"`
	Settings              runtime.RawExtension              `json:"settings,omitempty"`
}

type policyServerSourceAuthority struct {
	Type string `json:"type"`
	Data string `json:"data"` // contains a PEM encoded certificate
}

//nolint:tagliatelle
type policyServerSourcesEntry struct {
	InsecureSources   []string                                 `json:"insecure_sources,omitempty"`
	SourceAuthorities map[string][]policyServerSourceAuthority `json:"source_authorities,omitempty"`
}

// Reconciles the ConfigMap that holds the configuration of the Policy Server
func (r *PolicyServerReconciler) reconcilePolicyServerConfigMap(
	ctx context.Context,
	policyServer *policiesv1.PolicyServer,
	policies []policiesv1.Policy,
) error {
	cfg := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyServer.NameWithPrefix(),
			Namespace: r.DeploymentsNamespace,
		},
	}
	_, err := controllerutil.CreateOrPatch(ctx, r.Client, cfg, func() error {
		return r.updateConfigMapData(cfg, policyServer, policies)
	})
	if err != nil {
		return fmt.Errorf("cannot create or update PolicyServer ConfigMap: %w", err)
	}
	return nil
}

// Function used to update the ConfigMap data when creating or updating it
func (r *PolicyServerReconciler) updateConfigMapData(cfg *corev1.ConfigMap, policyServer *policiesv1.PolicyServer, policies []policiesv1.Policy) error {
	policiesMap := buildPoliciesMap(policies)
	policiesYML, err := json.Marshal(policiesMap)
	if err != nil {
		return fmt.Errorf("cannot marshal policies: %w", err)
	}

	sources := buildSourcesMap(policyServer)
	sourcesYML, err := json.Marshal(sources)
	if err != nil {
		return fmt.Errorf("cannot marshal insecureSources: %w", err)
	}

	data := map[string]string{
		constants.PolicyServerConfigPoliciesEntry: string(policiesYML),
		constants.PolicyServerConfigSourcesEntry:  string(sourcesYML),
	}

	cfg.Data = data
	cfg.ObjectMeta.Labels = map[string]string{
		constants.PolicyServerLabelKey: policyServer.ObjectMeta.Name,
	}
	if err := controllerutil.SetOwnerReference(policyServer, cfg, r.Client.Scheme()); err != nil {
		return errors.Join(errors.New("failed to set policy server configmap owner reference"), err)
	}
	return nil
}

func (r *PolicyServerReconciler) policyServerConfigMapVersion(ctx context.Context, policyServer *policiesv1.PolicyServer) (string, error) {
	// By using Unstructured data we force the client to fetch fresh, uncached
	// data from the API server
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    "ConfigMap",
		Version: "v1",
	})
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: r.DeploymentsNamespace,
		Name:      policyServer.NameWithPrefix(),
	}, unstructuredObj)
	if err != nil {
		return "", fmt.Errorf("cannot retrieve existing policies ConfigMap: %w", err)
	}

	return unstructuredObj.GetResourceVersion(), nil
}

func buildPoliciesMap(admissionPolicies []policiesv1.Policy) policyConfigEntryMap {
	policies := policyConfigEntryMap{}
	for _, admissionPolicy := range admissionPolicies {
		policies[admissionPolicy.GetUniqueName()] = policyServerConfigEntry{
			NamespacedName: types.NamespacedName{
				Namespace: admissionPolicy.GetNamespace(),
				Name:      admissionPolicy.GetName(),
			},
			URL:                   admissionPolicy.GetModule(),
			PolicyMode:            string(admissionPolicy.GetPolicyMode()),
			AllowedToMutate:       admissionPolicy.IsMutating(),
			Settings:              admissionPolicy.GetSettings(),
			ContextAwareResources: admissionPolicy.GetContextAwareResources(),
		}
	}
	return policies
}

func buildSourcesMap(policyServer *policiesv1.PolicyServer) policyServerSourcesEntry {
	sourcesEntry := policyServerSourcesEntry{}
	sourcesEntry.InsecureSources = policyServer.Spec.InsecureSources
	if sourcesEntry.InsecureSources == nil {
		sourcesEntry.InsecureSources = make([]string, 0)
	}

	sourcesEntry.SourceAuthorities = make(map[string][]policyServerSourceAuthority)
	// build sources.yml with data keys for Policy-server
	for uri, certs := range policyServer.Spec.SourceAuthorities {
		sourcesEntry.SourceAuthorities[uri] = make([]policyServerSourceAuthority, 0)
		for _, cert := range certs {
			sourcesEntry.SourceAuthorities[uri] = append(sourcesEntry.SourceAuthorities[uri],
				policyServerSourceAuthority{
					Type: dataType,
					Data: cert,
				})
		}
	}
	return sourcesEntry
}

type policyConfigEntryMap map[string]policyServerConfigEntry

func (e policyConfigEntryMap) toAdmissionPolicyReconcileRequests() []reconcile.Request {
	res := []reconcile.Request{}
	for _, policy := range e {
		if policy.NamespacedName.Namespace == "" {
			continue
		}
		res = append(res, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: policy.NamespacedName.Namespace,
				Name:      policy.NamespacedName.Name,
			},
		})
	}
	return res
}

func (e policyConfigEntryMap) toClusterAdmissionPolicyReconcileRequests() []reconcile.Request {
	res := []reconcile.Request{}
	for _, policy := range e {
		if policy.NamespacedName.Namespace != "" {
			continue
		}
		res = append(res, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: policy.NamespacedName.Name,
			},
		})
	}
	return res
}
