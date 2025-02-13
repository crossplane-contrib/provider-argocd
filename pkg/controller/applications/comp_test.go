package applications

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-argocd/apis/applications/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

func TestGetApplicationCondition(t *testing.T) {
	type args struct {
		status *v1alpha1.ArgoApplicationStatus
	}

	type want struct {
		condition xpv1.Condition
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NilStatus": {
			args: args{
				status: nil,
			},
			want: want{
				condition: xpv1.Unavailable().WithMessage("Application status is missing"),
			},
		},
		"EmptyStatus": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					Health: v1alpha1.HealthStatus{
						Status: "",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "",
					},
				},
			},
			want: want{
				condition: xpv1.Unavailable().WithMessage("Application status not yet propagated"),
			},
		},
		"EmptyHealthStatus": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					Health: v1alpha1.HealthStatus{
						Status: "",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "Synced",
					},
				},
			},
			want: want{
				condition: xpv1.Unavailable().WithMessage("Health status: Unknown"),
			},
		},
		"EmptySyncStatus": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					Health: v1alpha1.HealthStatus{
						Status: "Healthy",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "",
					},
				},
			},
			want: want{
				condition: xpv1.Available().WithMessage("Sync status: Unknown"),
			},
		},
		"NoOperationState": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					Health: v1alpha1.HealthStatus{
						Status: "Healthy",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "Synced",
					},
				},
			},
			want: want{
				condition: xpv1.Available(),
			},
		},
		"ReadyWithSyncedStatus": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					OperationState: &v1alpha1.OperationState{
						Phase: "Succeeded",
					},
					Health: v1alpha1.HealthStatus{
						Status: "Healthy",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "Synced",
					},
				},
			},
			want: want{
				condition: xpv1.Available(),
			},
		},
		"ReadyWithOutOfSyncStatus": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					OperationState: &v1alpha1.OperationState{
						Phase: "Succeeded",
					},
					Health: v1alpha1.HealthStatus{
						Status: "Healthy",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "OutOfSync",
					},
				},
			},
			want: want{
				condition: xpv1.Available().WithMessage("Sync status: OutOfSync"),
			},
		},
		"NotReadyDueToHealth": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					OperationState: &v1alpha1.OperationState{
						Phase: "Succeeded",
					},
					Health: v1alpha1.HealthStatus{
						Status: "Degraded",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "Synced",
					},
				},
			},
			want: want{
				condition: xpv1.Unavailable().WithMessage("Health status: Degraded"),
			},
		},
		"NotReadyDueToOperation": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					OperationState: &v1alpha1.OperationState{
						Phase: "Failed",
					},
					Health: v1alpha1.HealthStatus{
						Status: "Healthy",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "Synced",
					},
				},
			},
			want: want{
				condition: xpv1.Unavailable().WithMessage("Operation phase: Failed"),
			},
		},
		"NotReadyDueToEmptyOperation": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					OperationState: &v1alpha1.OperationState{
						Phase: "",
					},
					Health: v1alpha1.HealthStatus{
						Status: "Healthy",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "Synced",
					},
				},
			},
			want: want{
				condition: xpv1.Unavailable().WithMessage("Operation phase: Unknown"),
			},
		},
		"NotReadyWithMultipleIssues": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					OperationState: &v1alpha1.OperationState{
						Phase: "Failed",
					},
					Health: v1alpha1.HealthStatus{
						Status: "Degraded",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "OutOfSync",
					},
				},
			},
			want: want{
				condition: xpv1.Unavailable().WithMessage("Operation phase: Failed; Health status: Degraded; Sync status: OutOfSync"),
			},
		},
		"ReadyWithMissingHealthStatus": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					Health: v1alpha1.HealthStatus{
						Status: "Missing",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "OutOfSync",
					},
				},
			},
			want: want{
				condition: xpv1.Available().WithMessage("Application created but not yet deployed (auto-sync disabled)"),
			},
		},
		"ReadyWithMissingHealthAndNoSync": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					Health: v1alpha1.HealthStatus{
						Status: "Missing",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "",
					},
				},
			},
			want: want{
				condition: xpv1.Available().WithMessage("Application created but not yet deployed (auto-sync disabled)"),
			},
		},
		"ReadyWithMissingHealthAndFailedOperation": {
			args: args{
				status: &v1alpha1.ArgoApplicationStatus{
					OperationState: &v1alpha1.OperationState{
						Phase: "Failed",
					},
					Health: v1alpha1.HealthStatus{
						Status: "Missing",
					},
					Sync: v1alpha1.SyncStatus{
						Status: "OutOfSync",
					},
				},
			},
			want: want{
				condition: xpv1.Available().WithMessage("Application created but not yet deployed (auto-sync disabled)"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GetApplicationCondition(tc.args.status)
			if diff := cmp.Diff(tc.want.condition, got, cmpopts.IgnoreFields(xpv1.Condition{}, "LastTransitionTime")); diff != "" {
				t.Errorf("GetApplicationCondition(...): -want, +got:\n%s", diff)
			}
		})
	}
}