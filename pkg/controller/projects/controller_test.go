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

package projects

import (
	"context"
	"testing"
	"time"

	"k8s.io/utils/ptr"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-argocd/apis/projects/v1alpha1"
	mockclient "github.com/crossplane-contrib/provider-argocd/pkg/clients/mock/projects"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/projects"
)

var (
	errBoom                 = errors.New("boom")
	errNotFound             = errors.New("code = NotFound desc = appprojects")
	testProjectExternalName = "testproject"
	testDescription         = "This is a Test"
	testDescription2        = "This description changed"
	testLabels              = map[string]string{"label1": "value1"}
)

type args struct {
	client projects.ProjectServiceClient
	cr     *v1alpha1.Project
}

type mockModifier func(*mockclient.MockProjectServiceClient)

func withMockClient(t *testing.T, mod mockModifier) *mockclient.MockProjectServiceClient {
	ctrl := gomock.NewController(t)
	mock := mockclient.NewMockProjectServiceClient(ctrl)
	mod(mock)
	return mock
}

func Project(m ...ProjectModifier) *v1alpha1.Project {
	cr := &v1alpha1.Project{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type ProjectModifier func(*v1alpha1.Project)

func withExternalName(v string) ProjectModifier {
	return func(s *v1alpha1.Project) {
		meta.SetExternalName(s, v)
	}
}

func withSpec(p v1alpha1.ProjectParameters) ProjectModifier {
	return func(r *v1alpha1.Project) { r.Spec.ForProvider = p }
}

func withObjectMeta(p metav1.ObjectMeta) ProjectModifier {
	return func(r *v1alpha1.Project) { r.ObjectMeta = p }
}

func withObservation(p v1alpha1.ProjectObservation) ProjectModifier {
	return func(r *v1alpha1.Project) { r.Status.AtProvider = p }
}

func withConditions(c ...xpv1.Condition) ProjectModifier {
	return func(r *v1alpha1.Project) { r.Status.ConditionedStatus.Conditions = c }
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Project
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name:   testProjectExternalName,
								Labels: testLabels,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Description: testDescription,
							},
							Status: argocdv1alpha1.AppProjectStatus{},
						}, nil)
				}),
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description:   &testDescription,
						ProjectLabels: testLabels,
					}),
				),
			},
			want: want{
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description:   &testDescription,
						ProjectLabels: testLabels,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.ProjectObservation{
						JWTTokensByRole: map[string]v1alpha1.JWTTokens{},
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
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: testProjectExternalName,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Description: testDescription,
							},
							Status: argocdv1alpha1.AppProjectStatus{},
						}, nil)
				}),
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{}),
				),
			},
			want: want{
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.ProjectObservation{
						JWTTokensByRole: map[string]v1alpha1.JWTTokens{},
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
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: testProjectExternalName,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Description: testDescription,
							},
							Status: argocdv1alpha1.AppProjectStatus{},
						}, nil)
				}),
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription2,
					}),
				),
			},
			want: want{
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription2,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.ProjectObservation{
						JWTTokensByRole: map[string]v1alpha1.JWTTokens{},
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
		"GetProjectFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						nil, errBoom)
				}),
				cr: Project(
					withExternalName(testProjectExternalName),
				),
			},
			want: want{
				cr: Project(
					withExternalName(testProjectExternalName),
				),
				err: errors.Wrap(errBoom, errGetFailed),
			},
		},
		"GetProjectNotFound": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						nil, errNotFound)
				}),
				cr: Project(
					withExternalName(testProjectExternalName),
				),
			},
			want: want{
				cr: Project(
					withExternalName(testProjectExternalName),
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
		cr     *v1alpha1.Project
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Create(
						context.Background(),
						&project.ProjectCreateRequest{
							Project: &argocdv1alpha1.AppProject{
								ObjectMeta: metav1.ObjectMeta{Name: testProjectExternalName, Labels: testLabels},
								Spec: argocdv1alpha1.AppProjectSpec{
									Description: testDescription,
								},
							},
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name:   testProjectExternalName,
								Labels: testLabels,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Description: testDescription,
							},
							Status: argocdv1alpha1.AppProjectStatus{},
						}, nil)
				}),
				cr: Project(
					withObjectMeta(metav1.ObjectMeta{
						Name: testProjectExternalName,
					}),
					withSpec(v1alpha1.ProjectParameters{
						Description:   &testDescription,
						ProjectLabels: testLabels,
					}),
				),
			},
			want: want{
				cr: Project(
					withSpec(v1alpha1.ProjectParameters{
						Description:   &testDescription,
						ProjectLabels: testLabels,
					}),
					withObjectMeta(metav1.ObjectMeta{
						Name: testProjectExternalName,
					}),
					withExternalName(testProjectExternalName),
				),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateSystemFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Create(
						context.Background(),
						&project.ProjectCreateRequest{
							Project: &argocdv1alpha1.AppProject{
								ObjectMeta: metav1.ObjectMeta{Name: testProjectExternalName},
								Spec: argocdv1alpha1.AppProjectSpec{
									Description: testDescription,
								},
							},
						},
					).Return(
						nil, errBoom)
				}),
				cr: Project(
					withObjectMeta(metav1.ObjectMeta{
						Name: testProjectExternalName,
					}),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription,
					}),
					withExternalName(testProjectExternalName),
				),
			},
			want: want{
				cr: Project(
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription,
					}),
					withObjectMeta(metav1.ObjectMeta{
						Name: testProjectExternalName,
					}),
					withExternalName(testProjectExternalName),
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
		cr     *v1alpha1.Project
		result managed.ExternalUpdate
		err    error
	}

	annotations := make(map[string]string)
	annotations["crossplane.io/external-name"] = testProjectExternalName
	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name:   testProjectExternalName,
								Labels: testLabels,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Description: testDescription,
							},
							Status: argocdv1alpha1.AppProjectStatus{},
						}, nil)
					mcs.EXPECT().Update(
						context.Background(),
						gomock.Any(), // FIXME project.ProjectUpdateRequest objects can't be matched by gomock
						// project.ProjectUpdateRequest{
						// 	Project: &argocdv1alpha1.AppProject{
						// 		ObjectMeta: metav1.ObjectMeta{
						// 			Name:            testProjectExternalName,
						// 		},
						// 		Spec: argocdv1alpha1.AppProjectSpec{
						// 			Description: testDescription2,
						// 		},
						// 	},
						// },
					).Return(&argocdv1alpha1.AppProject{
						TypeMeta: metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{
							Name:   testProjectExternalName,
							Labels: testLabels,
						},
						Spec: argocdv1alpha1.AppProjectSpec{
							Description: testDescription2,
						},
						Status: argocdv1alpha1.AppProjectStatus{},
					}, nil)
				}),
				cr: Project(
					withObjectMeta(metav1.ObjectMeta{
						Name:   testProjectExternalName,
						Labels: testLabels,
					}),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription2,
					}),
					withExternalName(testProjectExternalName),
				),
			},
			want: want{
				cr: Project(
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription2,
					}),
					withObjectMeta(metav1.ObjectMeta{
						Name:   testProjectExternalName,
						Labels: testLabels,
					}),
					withExternalName(testProjectExternalName),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"ProjectNotFound": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						nil, errBoom)
				}),
				cr: Project(
					withObjectMeta(metav1.ObjectMeta{
						Name: testProjectExternalName,
					}),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription2,
					}),
					withExternalName(testProjectExternalName),
				),
			},
			want: want{
				cr: Project(
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription2,
					}),
					withObjectMeta(metav1.ObjectMeta{
						Name: testProjectExternalName,
					}),
					withExternalName(testProjectExternalName),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errBoom, errUpdateFailed),
			},
		},
		"UpdateProjectFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: testProjectExternalName,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Description: testDescription,
							},
							Status: argocdv1alpha1.AppProjectStatus{},
						}, nil)
					mcs.EXPECT().Update(
						context.Background(),
						gomock.Any(), // FIXME project.ProjectUpdateRequest objects can't be matched by gomock
						// project.ProjectUpdateRequest{
						// 	Project: &argocdv1alpha1.AppProject{
						// 		ObjectMeta: metav1.ObjectMeta{
						// 			Name:            testProjectExternalName,
						// 		},
						// 		Spec: argocdv1alpha1.AppProjectSpec{
						// 			Description: testDescription2,
						// 		},
						// 	},
						// },
					).Return(nil, errBoom)
				}),
				cr: Project(
					withObjectMeta(metav1.ObjectMeta{
						Name: testProjectExternalName,
					}),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription2,
					}),
					withExternalName(testProjectExternalName),
				),
			},
			want: want{
				cr: Project(
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription2,
					}),
					withObjectMeta(metav1.ObjectMeta{
						Name: testProjectExternalName,
					}),
					withExternalName(testProjectExternalName),
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
		cr  *v1alpha1.Project
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Delete(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						&project.EmptyResponse{}, nil)
				}),
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription,
					}),
				),
			},
			want: want{
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription,
					}),
				),
				err: nil,
			},
		},
		"DeleteFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Delete(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectExternalName,
						},
					).Return(
						&project.EmptyResponse{}, errBoom)
				}),
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription,
					}),
				),
			},
			want: want{
				cr: Project(
					withExternalName(testProjectExternalName),
					withSpec(v1alpha1.ProjectParameters{
						Description: &testDescription,
					}),
				),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsEqualJWTTokens(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name string
		p    []v1alpha1.ProjectToken
		r    []argocdv1alpha1.JWTToken
		want bool
	}{
		{
			name: "EqualTokens",
			p: []v1alpha1.ProjectToken{{
				ID:        "token1",
				ExpiresIn: ptr.To("1h"),
			}},
			r: []argocdv1alpha1.JWTToken{{
				ID:        "token1",
				IssuedAt:  now,
				ExpiresAt: now + 3600,
			}},
			want: true,
		},
		{
			name: "DifferentIDs",
			p: []v1alpha1.ProjectToken{{
				ID:        "token1",
				ExpiresIn: ptr.To("1h"),
			}},
			r: []argocdv1alpha1.JWTToken{{
				ID:        "token2",
				IssuedAt:  now,
				ExpiresAt: now + 3600,
			}},
			want: false,
		},
		{
			name: "DifferentExpiration",
			p: []v1alpha1.ProjectToken{{
				ID:        "token1",
				ExpiresIn: ptr.To("2h"),
			}},
			r: []argocdv1alpha1.JWTToken{{
				ID:        "token1",
				IssuedAt:  now,
				ExpiresAt: now + 3600,
			}},
			want: false,
		},
		{
			name: "TokenExpired",
			p: []v1alpha1.ProjectToken{{
				ID:        "token1",
				ExpiresIn: ptr.To("1h"),
			}},
			r: []argocdv1alpha1.JWTToken{{
				ID:        "token1",
				IssuedAt:  now - 7200,
				ExpiresAt: now - 3600,
			}},
			want: false,
		},
		{
			name: "DifferentLengths",
			p: []v1alpha1.ProjectToken{{
				ID:        "token1",
				ExpiresIn: ptr.To("1h"),
			}},
			r: []argocdv1alpha1.JWTToken{{
				ID:        "token1",
				IssuedAt:  now,
				ExpiresAt: now + 3600,
			}, {
				ID:        "token2",
				IssuedAt:  now,
				ExpiresAt: now + 7200,
			}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEqualJWTTokens(tt.p, tt.r); got != tt.want {
				t.Errorf("isEqualJWTTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateTokenUpdateRequests(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name           string
		projectName    string
		roles          []v1alpha1.ProjectRole
		existingTokens map[string]v1alpha1.JWTTokens
		want           map[string]*project.ProjectTokenCreateRequest
	}{
		{
			name:        "NoExistingTokens",
			projectName: "test-project",
			roles: []v1alpha1.ProjectRole{{
				Name: "role1",
				Tokens: []v1alpha1.ProjectToken{{
					ID:          "token1",
					ExpiresIn:   ptr.To("1h"),
					Description: ptr.To("Test token"),
				}},
			}},
			existingTokens: map[string]v1alpha1.JWTTokens{},
			want: map[string]*project.ProjectTokenCreateRequest{
				"role1.token1": {
					Project:     "test-project",
					Role:        "role1",
					Description: "Test token",
					ExpiresIn:   3600,
					Id:          "token1",
				},
			},
		},
		{
			name:        "ExistingTokenOK",
			projectName: "test-project",
			roles: []v1alpha1.ProjectRole{{
				Name: "role1",
				Tokens: []v1alpha1.ProjectToken{{
					ID:        "token1",
					ExpiresIn: ptr.To("1h"),
				}},
			}},
			existingTokens: map[string]v1alpha1.JWTTokens{
				"role1": {
					Items: []v1alpha1.JWTToken{{
						ID:        ptr.To("token1"),
						IssuedAt:  now - 4000,
						ExpiresAt: ptr.To(now + 3600),
					}},
				},
			},
			want: map[string]*project.ProjectTokenCreateRequest{},
		},
		{
			name:        "TokenNeedsRenewalBefore",
			projectName: "test-project",
			roles: []v1alpha1.ProjectRole{{
				Name: "role1",
				Tokens: []v1alpha1.ProjectToken{{
					ID:          "token1",
					ExpiresIn:   ptr.To("1h"),
					RenewBefore: ptr.To("30m"),
				}},
			}},
			existingTokens: map[string]v1alpha1.JWTTokens{
				"role1": {
					Items: []v1alpha1.JWTToken{{
						ID:        ptr.To("token1"),
						IssuedAt:  now - 3000,
						ExpiresAt: ptr.To(now + 600),
					}},
				},
			},
			want: map[string]*project.ProjectTokenCreateRequest{
				"role1.token1": {
					Project:   "test-project",
					Role:      "role1",
					ExpiresIn: 3600,
					Id:        "token1",
				},
			},
		},
		{
			name:        "TokenNeedsRenewalAfter",
			projectName: "test-project",
			roles: []v1alpha1.ProjectRole{{
				Name: "role1",
				Tokens: []v1alpha1.ProjectToken{{
					ID:         "token1",
					ExpiresIn:  ptr.To("1h"),
					RenewAfter: ptr.To("30m"),
				}},
			}},
			existingTokens: map[string]v1alpha1.JWTTokens{
				"role1": {
					Items: []v1alpha1.JWTToken{{
						ID:        ptr.To("token1"),
						IssuedAt:  now - 3000,
						ExpiresAt: ptr.To(now + 600),
					}},
				},
			},
			want: map[string]*project.ProjectTokenCreateRequest{
				"role1.token1": {
					Project:   "test-project",
					Role:      "role1",
					ExpiresIn: 3600,
					Id:        "token1",
				},
			},
		},
		{
			name:        "MultipleTokensOneNeedsRenewal",
			projectName: "test-project",
			roles: []v1alpha1.ProjectRole{{
				Name: "role1",
				Tokens: []v1alpha1.ProjectToken{
					{
						ID:          "token1",
						ExpiresIn:   ptr.To("1h"),
						RenewBefore: ptr.To("30m"),
					},
					{
						ID:          "token2",
						ExpiresIn:   ptr.To("1h"),
						RenewBefore: ptr.To("30m"),
					},
				},
			}},
			existingTokens: map[string]v1alpha1.JWTTokens{
				"role1": {
					Items: []v1alpha1.JWTToken{
						{
							ID:        ptr.To("token1"),
							IssuedAt:  now - 3000,
							ExpiresAt: ptr.To(now + 600),
						},
						{
							ID:        ptr.To("token2"),
							IssuedAt:  now - 100,
							ExpiresAt: ptr.To(now + 3500),
						},
					},
				},
			},
			want: map[string]*project.ProjectTokenCreateRequest{
				"role1.token1": {
					Project:   "test-project",
					Role:      "role1",
					ExpiresIn: 3600,
					Id:        "token1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateTokenUpdateRequests(tt.projectName, tt.roles, tt.existingTokens)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("generateTokenUpdateRequests() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
