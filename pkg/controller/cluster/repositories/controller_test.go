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

package repositories

import (
	"context"
	"testing"

	argocdRepository "github.com/argoproj/argo-cd/v3/pkg/apiclient/repository"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/repositories/v1alpha1"
	mockclient "github.com/crossplane-contrib/provider-argocd/pkg/clients/mock/repositories"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/repositories"
)

var (
	errBoom = errors.New("boom")
	// Unused until issue https://github.com/argoproj/argo-cd/issues/20005 in Argo CD project is resolved
	// errNotFound                = errors.New("code = NotFound desc = repo")
	errPermissionDenied        = errors.New("code = PermissionDenied desc = permission denied")
	testRepositoryExternalName = "testRepo"
	testRepo                   = "https://gitlab.com/example-group/example-project.git"
	testUsername               = "testUser"
	testInsecure               = false
	testEnableLFS              = false
	testInheritedCreds         = false
	testEnableOCI              = false
)

type args struct {
	client repositories.RepositoryServiceClient
	cr     *v1alpha1.Repository
}

type mockModifier func(client *mockclient.MockRepositoryServiceClient)

func withMockClient(t *testing.T, mod mockModifier) *mockclient.MockRepositoryServiceClient {
	ctrl := gomock.NewController(t)
	mock := mockclient.NewMockRepositoryServiceClient(ctrl)
	mod(mock)
	return mock
}

func Repository(m ...RepositoryModifier) *v1alpha1.Repository {
	cr := &v1alpha1.Repository{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type RepositoryModifier func(repository *v1alpha1.Repository)

func withExternalName(v string) RepositoryModifier {
	return func(s *v1alpha1.Repository) {
		meta.SetExternalName(s, v)
	}
}

func withSpec(p v1alpha1.RepositoryParameters) RepositoryModifier {
	return func(r *v1alpha1.Repository) { r.Spec.ForProvider = p }
}

func withObservation(p v1alpha1.RepositoryObservation) RepositoryModifier {
	return func(r *v1alpha1.Repository) { r.Status.AtProvider = p }
}

func withConditions(c ...xpv1.Condition) RepositoryModifier {
	return func(r *v1alpha1.Repository) { r.Status.ConditionedStatus.Conditions = c }
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Repository
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo: testRepositoryExternalName,
						},
					).Return(
						&argocdv1alpha1.Repository{
							Repo: testRepo,
							Name: testRepositoryExternalName,
						}, nil)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name:           ptr.To(testRepositoryExternalName),
						Repo:           testRepo,
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name:           ptr.To(testRepositoryExternalName),
						Repo:           testRepo,
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.RepositoryObservation{
						ConnectionState: v1alpha1.ConnectionState{},
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
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo: testRepositoryExternalName,
						},
					).Return(
						&argocdv1alpha1.Repository{
							Name: testRepositoryExternalName,
							Repo: testRepo,
						}, nil)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name: ptr.To(testRepositoryExternalName),
						Repo: testRepo,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name:           ptr.To(testRepositoryExternalName),
						Repo:           testRepo,
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.RepositoryObservation{}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
				err: nil,
			},
		},
		"NeedsCreation": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo: testRepositoryExternalName,
						},
					).Return(nil, errPermissionDenied) // Switch to errNotFound when issue https://github.com/argoproj/argo-cd/issues/20005 in Argo CD is solved
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name: ptr.To(testRepositoryExternalName),
						Repo: testRepo,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name: ptr.To(testRepositoryExternalName),
						Repo: testRepo,
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
				err: nil,
			},
		},
		"NeedsCreationNoExternalName": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {}),

				cr: Repository(
					withSpec(v1alpha1.RepositoryParameters{
						Name: ptr.To(testRepositoryExternalName),
						Repo: testRepo,
					}),
				),
			},
			want: want{
				cr: Repository(
					withSpec(v1alpha1.RepositoryParameters{
						Name: ptr.To(testRepositoryExternalName),
						Repo: testRepo,
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"SuccessfulWithAppProject": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo:       testRepositoryExternalName,
							AppProject: "test-project",
						},
					).Return(
						&argocdv1alpha1.Repository{
							Repo:    testRepo,
							Name:    testRepositoryExternalName,
							Project: "test-project",
						}, nil)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name:           ptr.To(testRepositoryExternalName),
						Repo:           testRepo,
						Project:        ptr.To("test-project"),
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name:           ptr.To(testRepositoryExternalName),
						Repo:           testRepo,
						Project:        ptr.To("test-project"),
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.RepositoryObservation{
						ConnectionState: v1alpha1.ConnectionState{},
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
		"SuccessfulWithoutAppProject": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo: testRepositoryExternalName,
						},
					).Return(
						&argocdv1alpha1.Repository{
							Repo: testRepo,
							Name: testRepositoryExternalName,
						}, nil)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name:           ptr.To(testRepositoryExternalName),
						Repo:           testRepo,
						Project:        nil, // Explicitly nil
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Name:           ptr.To(testRepositoryExternalName),
						Repo:           testRepo,
						Project:        nil,
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.RepositoryObservation{
						ConnectionState: v1alpha1.ConnectionState{},
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
		cr     *v1alpha1.Repository
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().CreateRepository(
						context.Background(),
						&argocdRepository.RepoCreateRequest{
							Repo: &argocdv1alpha1.Repository{
								Repo: testRepositoryExternalName,
							},
						},
					).Return(
						&argocdv1alpha1.Repository{
							Repo: testRepositoryExternalName,
						}, nil)
				}),
				cr: Repository(
					withSpec(v1alpha1.RepositoryParameters{
						Repo: testRepositoryExternalName,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo: testRepositoryExternalName,
					}),
				),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateSystemFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().CreateRepository(
						context.Background(),
						&argocdRepository.RepoCreateRequest{
							Repo: &argocdv1alpha1.Repository{
								Repo: testRepositoryExternalName,
							},
						},
					).Return(
						nil, errBoom)
				}),
				cr: Repository(
					withSpec(v1alpha1.RepositoryParameters{
						Repo: testRepositoryExternalName,
					}),
				),
			},
			want: want{
				cr: Repository(
					withSpec(v1alpha1.RepositoryParameters{
						Repo: testRepositoryExternalName,
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
		cr     *v1alpha1.Repository
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().UpdateRepository(
						context.Background(),
						&argocdRepository.RepoUpdateRequest{
							Repo: &argocdv1alpha1.Repository{
								Repo:     testRepositoryExternalName,
								Username: testUsername,
							},
						},
					).Return(&argocdv1alpha1.Repository{
						Repo:     testRepositoryExternalName,
						Username: testUsername,
					}, nil)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo:           testRepositoryExternalName,
						Username:       &testUsername,
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo:           testRepositoryExternalName,
						Username:       &testUsername,
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"UpdateRepositoryFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().UpdateRepository(
						context.Background(),
						&argocdRepository.RepoUpdateRequest{
							Repo: &argocdv1alpha1.Repository{
								Repo:     testRepositoryExternalName,
								Username: testUsername,
							},
						},
					).Return(nil, errBoom)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo:           testRepositoryExternalName,
						Username:       &testUsername,
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo:           testRepositoryExternalName,
						Username:       &testUsername,
						Insecure:       &testInsecure,
						EnableLFS:      &testEnableLFS,
						InheritedCreds: &testInheritedCreds,
						EnableOCI:      &testEnableOCI,
					}),
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
		cr  *v1alpha1.Repository
		err error
		res managed.ExternalDelete
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().DeleteRepository(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo: testRepositoryExternalName,
						},
					).Return(
						&argocdRepository.RepoResponse{}, nil)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo: testRepositoryExternalName,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo: testRepositoryExternalName,
					}),
				),
				err: nil,
			},
		},
		"DeleteFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().DeleteRepository(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo: testRepositoryExternalName,
						},
					).Return(
						&argocdRepository.RepoResponse{}, errBoom)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo: testRepositoryExternalName,
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo: testRepositoryExternalName,
					}),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"SuccessfulWithAppProject": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().DeleteRepository(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo:       testRepositoryExternalName,
							AppProject: "test-project",
						},
					).Return(
						&argocdRepository.RepoResponse{}, nil)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo:    testRepositoryExternalName,
						Project: ptr.To("test-project"),
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo:    testRepositoryExternalName,
						Project: ptr.To("test-project"),
					}),
				),
				err: nil,
			},
		},
		"SuccessfulWithoutAppProject": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockRepositoryServiceClient) {
					mcs.EXPECT().DeleteRepository(
						context.Background(),
						&argocdRepository.RepoQuery{
							Repo: testRepositoryExternalName,
						},
					).Return(
						&argocdRepository.RepoResponse{}, nil)
				}),
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo:    testRepositoryExternalName,
						Project: nil, // Explicitly nil
					}),
				),
			},
			want: want{
				cr: Repository(
					withExternalName(testRepositoryExternalName),
					withSpec(v1alpha1.RepositoryParameters{
						Repo:    testRepositoryExternalName,
						Project: nil,
					}),
				),
				err: nil,
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
