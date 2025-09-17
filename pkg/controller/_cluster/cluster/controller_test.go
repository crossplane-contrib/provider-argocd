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
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/cluster/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/cluster"
	mockclient "github.com/crossplane-contrib/provider-argocd/pkg/clients/mock/cluster"
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

func withConditions(c ...xpv1.Condition) ClusterModifier {
	return func(r *v1alpha1.Cluster) { r.Status.ConditionedStatus.Conditions = c }
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Cluster
		result managed.ExternalObservation
		err    error
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
					withConditions(xpv1.Available()),
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
					withConditions(xpv1.Available()),
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
					withConditions(xpv1.Available()),
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
			e := &external{client: tc.client}
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
				err:    nil,
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
			e := &external{client: tc.client}
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
