/*
Copyright 2021 The Crossplane Authors.

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

package cluster

import (
	"context"
	"testing"

	argocdCluster "github.com/argoproj/argo-cd/v3/pkg/apiclient/cluster"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/cluster/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/interface/cluster"
	mockclient "github.com/crossplane-contrib/provider-argocd/pkg/clients/interface/mock/cluster"
)

var (
	errBoom                 = errors.New("boom")
	errNotFound             = errors.New("code = NotFound desc = cluster")
	testClusterExternalName = "testcluster"
	testClusterServer       = "https://example.com/"
	testNamespaces          = [1]string{"default"}
	testUsername            = "testuser"
)

type args struct {
	kube   client.Client
	client cluster.ServiceClient
	cr     *v1alpha1.Cluster
}

type mockModifier func(*mockclient.MockServiceClient)

func withMockClient(t *testing.T, mod mockModifier) *mockclient.MockServiceClient {
	ctrl := gomock.NewController(t)
	mock := mockclient.NewMockServiceClient(ctrl)
	mod(mock)
	return mock
}

func Cluster(m ...ClusterModifier) *v1alpha1.Cluster {
	cr := &v1alpha1.Cluster{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type ClusterModifier func(*v1alpha1.Cluster)

func withExternalName(v string) ClusterModifier {
	return func(s *v1alpha1.Cluster) {
		meta.SetExternalName(s, v)
	}
}

func withSpec(p v1alpha1.ClusterParameters) ClusterModifier {
	return func(r *v1alpha1.Cluster) { r.Spec.ForProvider = p }
}

func withObservation(p v1alpha1.ClusterObservation) ClusterModifier {
	return func(r *v1alpha1.Cluster) { r.Status.AtProvider = p }
}

func withConditions(c ...xpv2.Condition) ClusterModifier {
	return func(r *v1alpha1.Cluster) { r.Status.ConditionedStatus.Conditions = c }
}

func withGeneration(g int64) ClusterModifier {
	return func(r *v1alpha1.Cluster) { r.Generation = g }
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Cluster
		result managed.ExternalObservation
		err    error
	}

	// ArgoCD redacts the bearer token, so only its ConfigHash tracks it.
	secretRefSpec := v1alpha1.ClusterParameters{
		Server: ptr.To(testClusterServer),
		Name:   ptr.To(testClusterExternalName),
		Config: v1alpha1.ClusterConfig{
			BearerTokenSecretRef: &v1alpha1.SecretReference{Name: "token", Namespace: "default", Key: "token"},
		},
	}
	withSecretResourceVersion := func(rv string) client.Client {
		return &test.MockClient{MockGet: func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
			obj.SetResourceVersion(rv)
			return nil
		}}
	}
	withObservedCluster := func(c *argocdv1alpha1.Cluster) cluster.ServiceClient {
		return withMockClient(t, func(mcs *mockclient.MockServiceClient) {
			mcs.EXPECT().Get(
				context.Background(),
				&argocdCluster.ClusterQuery{
					Name:   testClusterExternalName,
					Server: testClusterServer,
				},
			).Return(c, nil)
		})
	}
	emptyObservation := func(l *v1alpha1.AppliedClusterObservation) v1alpha1.ClusterObservation {
		return v1alpha1.ClusterObservation{
			ClusterInfo: v1alpha1.ClusterInfo{
				ConnectionState: &v1alpha1.ConnectionState{},
				ServerVersion:   new(string),
				CacheInfo: &v1alpha1.ClusterCacheInfo{
					ResourcesCount: new(int64),
					APIsCount:      new(int64),
				},
			},
			LastApplied: l,
		}
	}
	lastApplied := &v1alpha1.AppliedClusterObservation{
		ConfigHash:             "42",
		Generation:             1,
		SecretResourceVersions: map[string]string{"default/token": "1"},
	}
	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdCluster.ClusterQuery{
							Name:   testClusterExternalName,
							Server: testClusterServer,
						},
					).Return(
						&argocdv1alpha1.Cluster{
							Server: testClusterServer,
							Name:   testClusterExternalName,
							Config: argocdv1alpha1.ClusterConfig{
								TLSClientConfig: argocdv1alpha1.TLSClientConfig{
									Insecure: true,
								},
							},
						}, nil)
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
					withConditions(xpv2.Available()),
					withObservation(v1alpha1.ClusterObservation{
						ClusterInfo: v1alpha1.ClusterInfo{
							ConnectionState: &v1alpha1.ConnectionState{},
							ServerVersion:   new(string),
							CacheInfo: &v1alpha1.ClusterCacheInfo{
								ResourcesCount: new(int64),
								APIsCount:      new(int64),
							},
							ApplicationsCount: 0,
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
				err: nil,
			},
		},
		"SuccessfulLateInitialize": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdCluster.ClusterQuery{
							Name:   testClusterExternalName,
							Server: testClusterServer,
						},
					).Return(
						&argocdv1alpha1.Cluster{
							Server:     testClusterServer,
							Name:       testClusterExternalName,
							Namespaces: testNamespaces[:],
							Config: argocdv1alpha1.ClusterConfig{
								TLSClientConfig: argocdv1alpha1.TLSClientConfig{
									Insecure: true,
								},
							},
						}, nil)
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server:     ptr.To(testClusterServer),
						Name:       ptr.To(testClusterExternalName),
						Namespaces: testNamespaces[:],
						Config: v1alpha1.ClusterConfig{
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
					withConditions(xpv2.Available()),
					withObservation(v1alpha1.ClusterObservation{
						ClusterInfo: v1alpha1.ClusterInfo{
							ConnectionState: &v1alpha1.ConnectionState{},
							ServerVersion:   new(string),
							CacheInfo: &v1alpha1.ClusterCacheInfo{
								ResourcesCount: new(int64),
								APIsCount:      new(int64),
							},
							ApplicationsCount: 0,
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
				err: nil,
			},
		},
		"LabelsNotUpToDate": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdCluster.ClusterQuery{
							Name:   testClusterExternalName,
							Server: testClusterServer,
						},
					).Return(
						&argocdv1alpha1.Cluster{
							Server: testClusterServer,
							Name:   testClusterExternalName,
							Config: argocdv1alpha1.ClusterConfig{
								TLSClientConfig: argocdv1alpha1.TLSClientConfig{
									Insecure: true,
								},
							},
						}, nil)
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							Username: &testUsername,
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							Username: &testUsername,
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
					withConditions(xpv2.Available()),
					withObservation(v1alpha1.ClusterObservation{
						ClusterInfo: v1alpha1.ClusterInfo{
							ConnectionState: &v1alpha1.ConnectionState{},
							ServerVersion:   new(string),
							CacheInfo: &v1alpha1.ClusterCacheInfo{
								ResourcesCount: new(int64),
								APIsCount:      new(int64),
							},
							ApplicationsCount: 0,
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
				err: nil,
			},
		},
		"LastAppliedUpToDate": {
			args: args{
				kube: withSecretResourceVersion("1"),
				client: withObservedCluster(&argocdv1alpha1.Cluster{
					Server:     testClusterServer,
					Name:       testClusterExternalName,
					ConfigHash: ptr.To(uint64(42)),
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withObservation(emptyObservation(lastApplied)),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withConditions(xpv2.Available()),
					withObservation(emptyObservation(lastApplied)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"NoLastApplied": {
			// Nothing was applied yet, e.g. after creation, so the cluster is updated once to record it.
			args: args{
				kube: withSecretResourceVersion("1"),
				client: withObservedCluster(&argocdv1alpha1.Cluster{
					Server:     testClusterServer,
					Name:       testClusterExternalName,
					ConfigHash: ptr.To(uint64(42)),
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withObservation(emptyObservation(nil)),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withConditions(xpv2.Available()),
					withObservation(emptyObservation(nil)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"ConfigHashChanged": {
			// The cluster was changed on the ArgoCD side.
			args: args{
				kube: withSecretResourceVersion("1"),
				client: withObservedCluster(&argocdv1alpha1.Cluster{
					Server:     testClusterServer,
					Name:       testClusterExternalName,
					ConfigHash: ptr.To(uint64(43)),
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withObservation(emptyObservation(lastApplied)),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withConditions(xpv2.Available()),
					withObservation(emptyObservation(lastApplied)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"GenerationChanged": {
			// The spec was edited.
			args: args{
				kube: withSecretResourceVersion("1"),
				client: withObservedCluster(&argocdv1alpha1.Cluster{
					Server:     testClusterServer,
					Name:       testClusterExternalName,
					ConfigHash: ptr.To(uint64(42)),
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(2),
					withSpec(*secretRefSpec.DeepCopy()),
					withObservation(emptyObservation(lastApplied)),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(2),
					withSpec(*secretRefSpec.DeepCopy()),
					withConditions(xpv2.Available()),
					withObservation(emptyObservation(lastApplied)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"SecretRotated": {
			args: args{
				kube: withSecretResourceVersion("2"),
				client: withObservedCluster(&argocdv1alpha1.Cluster{
					Server:     testClusterServer,
					Name:       testClusterExternalName,
					ConfigHash: ptr.To(uint64(42)),
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withObservation(emptyObservation(lastApplied)),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withConditions(xpv2.Available()),
					withObservation(emptyObservation(lastApplied)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"LastAppliedUpToDateLabelsNotUpToDate": {
			// The ConfigHash doesn't cover labels, so they must still be compared.
			args: args{
				kube: withSecretResourceVersion("1"),
				client: withObservedCluster(&argocdv1alpha1.Cluster{
					Server:     testClusterServer,
					Name:       testClusterExternalName,
					Labels:     map[string]string{"stale": "label"},
					ConfigHash: ptr.To(uint64(42)),
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withObservation(emptyObservation(lastApplied)),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withSpec(*secretRefSpec.DeepCopy()),
					withConditions(xpv2.Available()),
					withObservation(emptyObservation(lastApplied)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"GetClusterFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdCluster.ClusterQuery{
							Name: testClusterExternalName,
						},
					).Return(
						nil, errBoom)
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
				),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"GetClusterNotFound": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdCluster.ClusterQuery{
							Name: testClusterExternalName,
						},
					).Return(
						nil, errNotFound)
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
				),
				result: managed.ExternalObservation{},
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Cluster
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Create(
						context.Background(),
						&argocdCluster.ClusterCreateRequest{
							Cluster: &argocdv1alpha1.Cluster{
								Server: testClusterServer,
								Name:   testClusterExternalName,
								Config: argocdv1alpha1.ClusterConfig{
									TLSClientConfig: argocdv1alpha1.TLSClientConfig{
										Insecure: true,
									},
								},
							},
						},
					).Return(
						&argocdv1alpha1.Cluster{
							Server: testClusterServer,
							Name:   testClusterExternalName,
							Config: argocdv1alpha1.ClusterConfig{
								TLSClientConfig: argocdv1alpha1.TLSClientConfig{
									Insecure: true,
								},
							},
						}, nil)
				}),
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
				),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateSystemFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Create(
						context.Background(),
						&argocdCluster.ClusterCreateRequest{
							Cluster: &argocdv1alpha1.Cluster{
								Server: testClusterServer,
								Name:   testClusterExternalName,
								Config: argocdv1alpha1.ClusterConfig{
									TLSClientConfig: argocdv1alpha1.TLSClientConfig{
										Insecure: true,
									},
								},
							},
						},
					).Return(
						nil, errBoom)
				}),
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
				),
			},
			want: want{
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							TLSClientConfig: &v1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Cluster
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Update(
						context.Background(),
						gomock.Any(), // FIXME cluster.ClusterUpdateRequest objects can't be matched by gomock
					).Return(&argocdv1alpha1.Cluster{
						Server:     testClusterServer,
						Name:       testClusterExternalName,
						Namespaces: testNamespaces[:],
						Config: argocdv1alpha1.ClusterConfig{
							TLSClientConfig: argocdv1alpha1.TLSClientConfig{
								Insecure: true,
							},
						},
					}, nil)
				}),
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server:     ptr.To(testClusterServer),
						Name:       ptr.To(testClusterExternalName),
						Namespaces: testNamespaces[:],
					}),
					withExternalName(testClusterExternalName),
					withGeneration(1),
				),
			},
			want: want{
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server:     ptr.To(testClusterServer),
						Name:       ptr.To(testClusterExternalName),
						Namespaces: testNamespaces[:],
					}),
					withExternalName(testClusterExternalName),
					withGeneration(1),
					withObservation(v1alpha1.ClusterObservation{
						LastApplied: &v1alpha1.AppliedClusterObservation{Generation: 1},
					}),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"SuccessfulRecordsLastApplied": {
			args: args{
				kube: &test.MockClient{MockGet: func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
					obj.SetResourceVersion("1")
					obj.(*corev1.Secret).Data = map[string][]byte{"token": []byte("secret")}
					return nil
				}},
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Update(
						context.Background(),
						gomock.Any(), // FIXME cluster.ClusterUpdateRequest objects can't be matched by gomock
					).Return(&argocdv1alpha1.Cluster{
						Server:     testClusterServer,
						Name:       testClusterExternalName,
						ConfigHash: ptr.To(uint64(42)),
					}, nil)
				}),
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							BearerTokenSecretRef: &v1alpha1.SecretReference{Name: "token", Namespace: "default", Key: "token"},
						},
					}),
					withExternalName(testClusterExternalName),
					withGeneration(2),
				),
			},
			want: want{
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
						Config: v1alpha1.ClusterConfig{
							BearerTokenSecretRef: &v1alpha1.SecretReference{Name: "token", Namespace: "default", Key: "token"},
						},
					}),
					withExternalName(testClusterExternalName),
					withGeneration(2),
					withObservation(v1alpha1.ClusterObservation{
						LastApplied: &v1alpha1.AppliedClusterObservation{
							ConfigHash:             "42",
							Generation:             2,
							SecretResourceVersions: map[string]string{"default/token": "1"},
						},
					}),
				),
				result: managed.ExternalUpdate{},
			},
		},
		"UpdateClusterFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Update(
						context.Background(),
						gomock.Any(), // FIXME cluster.ClusterUpdateRequest objects can't be matched by gomock
					).Return(nil, errBoom)
				}),
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server:     ptr.To(testClusterServer),
						Name:       ptr.To(testClusterExternalName),
						Namespaces: testNamespaces[:],
					}),
					withExternalName(testClusterExternalName),
				),
			},
			want: want{
				cr: Cluster(
					withSpec(v1alpha1.ClusterParameters{
						Server:     ptr.To(testClusterServer),
						Name:       ptr.To(testClusterExternalName),
						Namespaces: testNamespaces[:],
					}),
					withExternalName(testClusterExternalName),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
			u, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.Cluster
		err error
		res managed.ExternalDelete
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Delete(
						context.Background(),
						&argocdCluster.ClusterQuery{
							Name:   testClusterExternalName,
							Server: testClusterServer,
						},
					).Return(
						&argocdCluster.ClusterResponse{}, nil)
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
					}),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
					}),
				),
				err: nil,
			},
		},
		"DeleteFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Delete(
						context.Background(),
						&argocdCluster.ClusterQuery{
							Name:   testClusterExternalName,
							Server: testClusterServer,
						},
					).Return(
						&argocdCluster.ClusterResponse{}, errBoom)
				}),
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
					}),
				),
			},
			want: want{
				cr: Cluster(
					withExternalName(testClusterExternalName),
					withSpec(v1alpha1.ClusterParameters{
						Server: ptr.To(testClusterServer),
						Name:   ptr.To(testClusterExternalName),
					}),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			got, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.res, got); diff != "" {
				t.Errorf("res: -want +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
