package features

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/statemetrics"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Opts(o xpcontroller.Options) []managed.ReconcilerOption {
	opts := []managed.ReconcilerOption{}

	if o.Features.Enabled(feature.EnableAlphaChangeLogs) {
		opts = append(opts, managed.WithChangeLogger(o.ChangeLogOptions.ChangeLogger))
	}

	if o.Features.Enabled(feature.EnableBetaManagementPolicies) {
		opts = append(opts, managed.WithManagementPolicies())
	}

	return opts
}

func AddMRMetrics(mgr ctrl.Manager, o xpcontroller.Options, object resource.ManagedList) error {
	return mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, object, o.MetricOptions.PollStateMetricInterval))
}
