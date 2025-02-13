package applications

import (
	"maps"
	"slices"
	"strings"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-argocd/apis/applications/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/applications"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// IsApplicationUpToDate converts ApplicationParameters to its ArgoCD Counterpart and returns if they equal
func IsApplicationUpToDate(cr *v1alpha1.ApplicationParameters, remote *argocdv1alpha1.Application) bool { // nolint:gocyclo
	converter := applications.ConverterImpl{}
	cluster := converter.ToArgoApplicationSpec(cr)

	opts := []cmp.Option{
		// explicitly ignore the unexported in this type instead of adding a generic allow on all type.
		// the unexported fields should not bother here, since we don't copy them or write them
		cmpopts.IgnoreUnexported(argocdv1alpha1.ApplicationDestination{}),
	}

	// Sort finalizer slices for comparison
	slices.Sort(cr.Finalizers)
	slices.Sort(remote.Finalizers)

	return cmp.Equal(*cluster, remote.Spec, opts...) && maps.Equal(cr.Annotations, remote.Annotations) && slices.Equal(cr.Finalizers, remote.Finalizers)
}

// GetApplicationCondition returns the condition of the application based on its status
// Ready State Matrix:
// ┌──────────────────┬───────────────┬────────────┬──────────────┬───────────┐
// │ Health Status    │ Sync Status   │ Operation  │ Ready State  │ Notes     │
// ├──────────────────┼───────────────┼────────────┼──────────────┼───────────┤
// │ ""               │ ""            │     *      │ Unavailable  │ No Status │
// │ "Missing"        │     *         │     *      │ Available    │ Not Sync'd│
// │ "Healthy"        │     *         │ Succeeded  │ Available    │ Deployed  │
// │ "Healthy"        │     *         │ Failed     │ Unavailable  │ Op Failed │
// │ "Degraded"       │     *         │     *      │ Unavailable  │ Unhealthy │
// │ "Progressing"    │     *         │     *      │ Unavailable  │ Not Ready │
// │ Unknown          │     *         │     *      │ Unavailable  │ Bad State │
// └──────────────────┴───────────────┴────────────┴──────────────┴───────────┘
//
// Notes:
// - Empty status ("") indicates status not yet propagated
// - Sync status is tracked but does not affect Ready state
func GetApplicationCondition(status *v1alpha1.ArgoApplicationStatus) xpv1.Condition {
	if status == nil {
		return xpv1.Unavailable().WithMessage("Application status is missing")
	}

	var messages []string

	// Check if status fields are empty
	if status.Health.Status == "" && status.Sync.Status == "" {
		return xpv1.Unavailable().WithMessage("Application status not yet propagated")
	}

	if status.Health.Status == "Missing" {
		return xpv1.Available().WithMessage("Application created but not yet deployed (auto-sync disabled)")
	}

	operationSucceeded := false
	if status.OperationState != nil {
		if status.OperationState.Phase == "Succeeded" {
			operationSucceeded = true
		} else {
			if status.OperationState.Phase != "" {
				messages = append(messages, "Operation phase: "+string(status.OperationState.Phase))
			} else {
				messages = append(messages, "Operation phase: Unknown")
			}
		}
	} else {
		operationSucceeded = true // No operation in progress means success
	}

	healthOK := false // Changed default to false
	if status.Health.Status == "Healthy" {
		healthOK = true
	} else if status.Health.Status != "" {
		messages = append(messages, "Health status: "+status.Health.Status)
	} else {
		messages = append(messages, "Health status: Unknown")
	}

	if status.Sync.Status != "Synced" {
		if status.Sync.Status != "" {
			messages = append(messages, "Sync status: "+status.Sync.Status)
		} else {
			messages = append(messages, "Sync status: Unknown")
		}
	}

	message := strings.Join(messages, "; ")

	if operationSucceeded && healthOK {
		if message == "" {
			return xpv1.Available()
		}
		return xpv1.Available().WithMessage(message)
	}

	return xpv1.Unavailable().WithMessage(message)
}