package applications

import (
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/applications/v1alpha1"
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
				condition: xpv1.Unavailable(),
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
				condition: xpv1.Available(),
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
				condition: xpv1.Unavailable(),
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
				condition: xpv1.Unavailable(),
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
				condition: xpv1.Unavailable(),
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
				condition: xpv1.Unavailable(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := getApplicationCondition(tc.args.status)
			if diff := cmp.Diff(tc.want.condition, got, cmpopts.IgnoreFields(xpv1.Condition{}, "LastTransitionTime")); diff != "" {
				t.Errorf("getApplicationCondition(...): -want, +got:\n%s", diff)
			}
		})
	}
}
