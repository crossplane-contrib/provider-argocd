/*
Copyright 2022 The Crossplane Authors.

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

package applicationsets

import (
	"context"
	"testing"

	argoapplicationset "github.com/argoproj/argo-cd/v3/pkg/apiclient/applicationset"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/applicationsets/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/applicationsets"
	mockclient "github.com/crossplane-contrib/provider-argocd/pkg/clients/mock/applicationsets"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

var (
	errBoom                        = errors.New("boom")
	testApplicationSetExternalName = "test"
	testApplicationSetNamespace    = "test-namespace"
	testTemplateName               = "myTemplate"
	testProjectName                = "myProject"
	otherProjectName               = "otherProject"
)

type mockModifier func(*mockclient.MockServiceClient)

func withMockClient(t *testing.T, mod mockModifier) *mockclient.MockServiceClient {
	ctrl := gomock.NewController(t)
	mock := mockclient.NewMockServiceClient(ctrl)
	mod(mock)
	return mock
}

func ApplicationSet(m ...ApplicationSetModifier) *v1alpha1.ApplicationSet {
	cr := &v1alpha1.ApplicationSet{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type ApplicationSetModifier func(*v1alpha1.ApplicationSet)

func withExternalName(v string) ApplicationSetModifier {
	return func(s *v1alpha1.ApplicationSet) {
		meta.SetExternalName(s, v)
	}
}

func withName(v string) ApplicationSetModifier {
	return func(s *v1alpha1.ApplicationSet) {
		s.Name = v
	}
}

func withSpec(p v1alpha1.ApplicationSetParameters) ApplicationSetModifier {
	return func(r *v1alpha1.ApplicationSet) { r.Spec.ForProvider = p }
}

func withAppSetNamespace(a *string) ApplicationSetModifier {
	return func(r *v1alpha1.ApplicationSet) { r.Spec.ForProvider.AppsetNamespace = a }
}

func withObservation(p v1alpha1.ArgoApplicationSetStatus) ApplicationSetModifier {
	return func(r *v1alpha1.ApplicationSet) { r.Status.AtProvider = p }
}

func withConditions(c ...xpv1.Condition) ApplicationSetModifier {
	return func(r *v1alpha1.ApplicationSet) { r.Status.ConditionedStatus.Conditions = c }
}

type args struct {
	cr     *v1alpha1.ApplicationSet
	client applicationsets.ServiceClient
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.ApplicationSet
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulAvailable": {
			args: args{
				client: withMockClient(t, func(m *mockclient.MockServiceClient) {
					m.EXPECT().Get(gomock.Any(), &argoapplicationset.ApplicationSetGetQuery{
						Name: testApplicationSetExternalName,
					}).Return(
						&argocdv1alpha1.ApplicationSet{
							Spec: argocdv1alpha1.ApplicationSetSpec{
								Template: argocdv1alpha1.ApplicationSetTemplate{
									ApplicationSetTemplateMeta: argocdv1alpha1.ApplicationSetTemplateMeta{
										Name: testTemplateName,
									},
									Spec: argocdv1alpha1.ApplicationSpec{
										Project: testProjectName,
									},
								},
							},
							Status: argocdv1alpha1.ApplicationSetStatus{
								Conditions: []argocdv1alpha1.ApplicationSetCondition{
									{Type: argocdv1alpha1.ApplicationSetConditionErrorOccurred},
								},
							},
						},
						nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.ArgoApplicationSetStatus{
						Conditions: []v1alpha1.ApplicationSetCondition{
							{Type: "ErrorOccurred"},
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
		"SuccessfulAvailableWithAppSetNamespace": {
			args: args{
				client: withMockClient(t, func(m *mockclient.MockServiceClient) {
					m.EXPECT().Get(gomock.Any(), &argoapplicationset.ApplicationSetGetQuery{
						Name:            testApplicationSetExternalName,
						AppsetNamespace: testApplicationSetNamespace,
					}).Return(
						&argocdv1alpha1.ApplicationSet{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: testApplicationSetNamespace,
							},
							Spec: argocdv1alpha1.ApplicationSetSpec{
								Template: argocdv1alpha1.ApplicationSetTemplate{
									ApplicationSetTemplateMeta: argocdv1alpha1.ApplicationSetTemplateMeta{
										Name: testTemplateName,
									},
									Spec: argocdv1alpha1.ApplicationSpec{
										Project: testProjectName,
									},
								},
							},
							Status: argocdv1alpha1.ApplicationSetStatus{
								Conditions: []argocdv1alpha1.ApplicationSetCondition{
									{Type: argocdv1alpha1.ApplicationSetConditionErrorOccurred},
								},
							},
						},
						nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
					withAppSetNamespace(&testApplicationSetNamespace),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
					withAppSetNamespace(&testApplicationSetNamespace),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.ArgoApplicationSetStatus{
						Conditions: []v1alpha1.ApplicationSetCondition{
							{Type: "ErrorOccurred"},
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
		"ProjectName different, needsUpdate": {
			args: args{
				client: withMockClient(t, func(m *mockclient.MockServiceClient) {
					m.EXPECT().Get(gomock.Any(), &argoapplicationset.ApplicationSetGetQuery{
						Name: testApplicationSetExternalName,
					}).Return(
						&argocdv1alpha1.ApplicationSet{
							Spec: argocdv1alpha1.ApplicationSetSpec{
								Template: argocdv1alpha1.ApplicationSetTemplate{
									ApplicationSetTemplateMeta: argocdv1alpha1.ApplicationSetTemplateMeta{
										Name: testTemplateName,
									},
									Spec: argocdv1alpha1.ApplicationSpec{
										Project: otherProjectName,
									},
								},
							},
						},
						nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
				err: nil,
			},
		},
		"Get of ApplicationSet fails, returns error": {
			args: args{
				client: withMockClient(t, func(m *mockclient.MockServiceClient) {
					m.EXPECT().Get(gomock.Any(), &argoapplicationset.ApplicationSetGetQuery{
						Name: testApplicationSetExternalName,
					}).Return(
						nil, errBoom)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errBoom, errGetApplicationSet),
			},
		},
		"ApplicationSet does not exists, needsCreate": {
			args: args{
				client: withMockClient(t, func(m *mockclient.MockServiceClient) {
					m.EXPECT().Get(gomock.Any(), &argoapplicationset.ApplicationSetGetQuery{
						Name: testApplicationSetExternalName,
					}).Return(
						nil, notFoundErr(),
					)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          false,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
				err: nil,
			},
		},
		"No external name, needsCreate": {
			args: args{
				client: withMockClient(t, func(m *mockclient.MockServiceClient) {}),
				cr: ApplicationSet(
					withName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
				),
			},
			want: want{
				cr: ApplicationSet(
					withName(testApplicationSetExternalName),
					withSpec(simpleApplicationSetParameters()),
				),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.args.client}
			o, err := e.Observe(context.TODO(), tc.args.cr)
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

func notFoundErr() error {
	return status.Errorf(codes.NotFound, "not found")
}

func simpleApplicationSetParameters() v1alpha1.ApplicationSetParameters {
	return v1alpha1.ApplicationSetParameters{
		Template: v1alpha1.ApplicationSetTemplate{
			ApplicationSetTemplateMeta: v1alpha1.ApplicationSetTemplateMeta{
				Name: testTemplateName,
			},
			Spec: v1alpha1.ApplicationSpec{
				Project: testProjectName,
			},
		},
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.ApplicationSet
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: withMockClient(t, func(m *mockclient.MockServiceClient) {
					// Might panic on diff.
					// See https://github.com/argoproj/argo-cd/issues/22081
					m.EXPECT().Create(
						gomock.Any(),
						&argoapplicationset.ApplicationSetCreateRequest{
							Applicationset: &argocdv1alpha1.ApplicationSet{
								ObjectMeta: metav1.ObjectMeta{
									Name: testApplicationSetExternalName,
								},
								Spec: *ArgoAppSpec(),
							},
						},
					).Return(
						&argocdv1alpha1.ApplicationSet{
							ObjectMeta: metav1.ObjectMeta{
								Name: testApplicationSetExternalName,
							},
						}, nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withName(testApplicationSetExternalName),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withName(testApplicationSetExternalName),
				),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"SuccessfulWithAppSetNamespace": {
			args: args{
				client: withMockClient(t, func(m *mockclient.MockServiceClient) {
					// Might panic on diff.
					// See https://github.com/argoproj/argo-cd/issues/22081
					m.EXPECT().Create(
						gomock.Any(),
						&argoapplicationset.ApplicationSetCreateRequest{
							Applicationset: &argocdv1alpha1.ApplicationSet{
								ObjectMeta: metav1.ObjectMeta{
									Name:      testApplicationSetExternalName,
									Namespace: testApplicationSetNamespace,
								},
								Spec: *ArgoAppSpec(),
							},
						},
					).Return(
						&argocdv1alpha1.ApplicationSet{
							ObjectMeta: metav1.ObjectMeta{
								Name:      testApplicationSetExternalName,
								Namespace: testApplicationSetNamespace,
							},
						}, nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withName(testApplicationSetExternalName),
					withAppSetNamespace(&testApplicationSetNamespace),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withName(testApplicationSetExternalName),
					withAppSetNamespace(&testApplicationSetNamespace),
				),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateSystemFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Create(
						gomock.Any(),
						gomock.Any(),
					).Return(
						nil, errBoom)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withName(testApplicationSetExternalName),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withName(testApplicationSetExternalName),
				),
				result: managed.ExternalCreation{},
				err:    errBoom,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.args.client}
			o, err := e.Create(context.TODO(), tc.args.cr)

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

type ArgoApplicationSetSpecModifier func(*argocdv1alpha1.ApplicationSetSpec)

func ArgoAppSpec(m ...ArgoApplicationSetSpecModifier) *argocdv1alpha1.ApplicationSetSpec {
	cr := argocdv1alpha1.ApplicationSetSpec{
		Template: argocdv1alpha1.ApplicationSetTemplate{
			ApplicationSetTemplateMeta: argocdv1alpha1.ApplicationSetTemplateMeta{},
		},
	}
	for _, f := range m {
		f(&cr)
	}
	return &cr
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.ApplicationSet
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
					mcs.EXPECT().Create(
						gomock.Any(),
						&argoapplicationset.ApplicationSetCreateRequest{
							Applicationset: &argocdv1alpha1.ApplicationSet{
								ObjectMeta: metav1.ObjectMeta{
									Name: testApplicationSetExternalName,
								},
								Spec: *ArgoAppSpec(),
							},
							Upsert: true,
						},
					).Return(&argocdv1alpha1.ApplicationSet{
						ObjectMeta: metav1.ObjectMeta{
							Name: testApplicationSetExternalName,
						},
					}, nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(v1alpha1.ApplicationSetParameters{}),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(v1alpha1.ApplicationSetParameters{}),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"SuccessfulWithAppSetNamespace": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Create(
						gomock.Any(),
						&argoapplicationset.ApplicationSetCreateRequest{
							Applicationset: &argocdv1alpha1.ApplicationSet{
								ObjectMeta: metav1.ObjectMeta{
									Name:      testApplicationSetExternalName,
									Namespace: testApplicationSetNamespace,
								},
								Spec: *ArgoAppSpec(),
							},
							Upsert: true,
						},
					).Return(&argocdv1alpha1.ApplicationSet{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testApplicationSetExternalName,
							Namespace: testApplicationSetNamespace,
						},
					}, nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(v1alpha1.ApplicationSetParameters{}),
					withAppSetNamespace(&testApplicationSetNamespace),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withSpec(v1alpha1.ApplicationSetParameters{}),
					withAppSetNamespace(&testApplicationSetNamespace),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"UpdateFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Create(
						gomock.Any(),
						&argoapplicationset.ApplicationSetCreateRequest{
							Applicationset: &argocdv1alpha1.ApplicationSet{
								ObjectMeta: metav1.ObjectMeta{
									Name: testApplicationSetExternalName,
								},
								Spec: *ArgoAppSpec(),
							},
							Upsert: true,
						},
					).Return(nil, errBoom)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
				),
				result: managed.ExternalUpdate{},
				err:    errBoom,
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
		cr  *v1alpha1.ApplicationSet
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
						gomock.Any(),
						&argoapplicationset.ApplicationSetDeleteRequest{
							Name: testApplicationSetExternalName,
						},
					).Return(&argoapplicationset.ApplicationSetResponse{}, nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
				),
				err: nil,
			},
		},
		"SuccessfulWithAppSetNamespace": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Delete(
						gomock.Any(),
						&argoapplicationset.ApplicationSetDeleteRequest{
							Name:            testApplicationSetExternalName,
							AppsetNamespace: testApplicationSetNamespace,
						},
					).Return(&argoapplicationset.ApplicationSetResponse{}, nil)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withAppSetNamespace(&testApplicationSetNamespace),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
					withAppSetNamespace(&testApplicationSetNamespace),
				),
				err: nil,
			},
		},
		"DeleteFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Delete(
						gomock.Any(),
						&argoapplicationset.ApplicationSetDeleteRequest{
							Name: testApplicationSetExternalName,
						},
					).Return(&argoapplicationset.ApplicationSetResponse{}, errBoom)
				}),
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
				),
			},
			want: want{
				cr: ApplicationSet(
					withExternalName(testApplicationSetExternalName),
				),
				err: errBoom,
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
