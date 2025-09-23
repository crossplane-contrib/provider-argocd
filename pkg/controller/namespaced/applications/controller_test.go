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

package applications

import (
	"context"
	"testing"

	argocdApplication "github.com/argoproj/argo-cd/v3/pkg/apiclient/application"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-argocd/apis/namespaced/applications/v1alpha1"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/applications"
	mockclient "github.com/crossplane-contrib/provider-argocd/pkg/clients/mock/applications"
)

var (
	errBoom                     = errors.New("boom")
	testApplicationExternalName = "testapplication"
	testProjectName             = "default"
	testDestinationNamespace    = "default-at-destination"
	testAppNamespace            = "test-app-namespace"
	emptyString                 = ""
	repoURL                     = "https://github.com/stefanprodan/podinfo/"
	chartPath                   = "charts/podinfo"
	revision                    = "HEAD"
	selfHealEnabled             = true
	testApplicationAnnotations  = map[string]string{"annotation1": "value1", "annotation2": "value2"}
	testApplicationFinalizers   = []string{"resources-finalizer.argocd.argoproj.io"}
)

type args struct {
	client applications.ServiceClient
	cr     *v1alpha1.Application
}

type mockModifier func(*mockclient.MockServiceClient)

func withMockClient(t *testing.T, mod mockModifier) *mockclient.MockServiceClient {
	ctrl := gomock.NewController(t)
	mock := mockclient.NewMockServiceClient(ctrl)
	mod(mock)
	return mock
}

func Application(m ...ApplicationModifier) *v1alpha1.Application {
	cr := &v1alpha1.Application{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type ApplicationModifier func(*v1alpha1.Application)

func withExternalName(v string) ApplicationModifier {
	return func(s *v1alpha1.Application) {
		meta.SetExternalName(s, v)
	}
}

func withName(v string) ApplicationModifier {
	return func(s *v1alpha1.Application) {
		s.Name = v
	}
}

func withSpec(p v1alpha1.ApplicationParameters) ApplicationModifier {
	return func(r *v1alpha1.Application) { r.Spec.ForProvider = p }
}

func withAppNamespace(n *string) ApplicationModifier {
	return func(r *v1alpha1.Application) { r.Spec.ForProvider.AppNamespace = n }
}

func withObservation(p v1alpha1.ArgoApplicationStatus) ApplicationModifier {
	return func(r *v1alpha1.Application) { r.Status.AtProvider = p }
}

func withConditions(c ...xpv1.Condition) ApplicationModifier {
	return func(r *v1alpha1.Application) { r.Status.ConditionedStatus.Conditions = c }
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Application
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
					mcs.EXPECT().List(
						context.Background(),
						&argocdApplication.ApplicationQuery{
							Name:     &testApplicationExternalName,
							Projects: []string{testProjectName},
						},
					).Return(
						&argocdv1alpha1.ApplicationList{
							TypeMeta: metav1.TypeMeta{},
							ListMeta: metav1.ListMeta{},
							Items: []argocdv1alpha1.Application{{
								TypeMeta: metav1.TypeMeta{},
								ObjectMeta: metav1.ObjectMeta{
									Name:        testApplicationExternalName,
									Annotations: testApplicationAnnotations,
									Finalizers:  testApplicationFinalizers,
								},
								Spec: argocdv1alpha1.ApplicationSpec{
									Project: testProjectName,
									Source: &argocdv1alpha1.ApplicationSource{
										RepoURL:        repoURL,
										Path:           chartPath,
										TargetRevision: revision,
									},
									Destination: argocdv1alpha1.ApplicationDestination{
										Namespace: testDestinationNamespace,
									},
									SyncPolicy: &argocdv1alpha1.SyncPolicy{
										Automated: &argocdv1alpha1.SyncPolicyAutomated{
											SelfHeal: true,
										},
									},
								},
								Status: argocdv1alpha1.ApplicationStatus{
									Health: argocdv1alpha1.HealthStatus{
										Status: "Healthy",
									},
									Sync: argocdv1alpha1.SyncStatus{
										Status: "Synced",
									},
								},
							}},
						}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project: testProjectName,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Source: &v1alpha1.ApplicationSource{
							RepoURL:        repoURL,
							Path:           &chartPath,
							TargetRevision: &revision,
						},
						SyncPolicy: &v1alpha1.SyncPolicy{
							Automated: &v1alpha1.SyncPolicyAutomated{
								SelfHeal: &selfHealEnabled,
							},
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project: testProjectName,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Source: &v1alpha1.ApplicationSource{
							RepoURL:        repoURL,
							Path:           &chartPath,
							TargetRevision: &revision,
						},
						SyncPolicy: &v1alpha1.SyncPolicy{
							Automated: &v1alpha1.SyncPolicyAutomated{
								SelfHeal: &selfHealEnabled,
							},
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.ArgoApplicationStatus{
						Resources: nil,
						Sync: v1alpha1.SyncStatus{
							Status:   "Synced",
							Revision: &emptyString,
							ComparedTo: v1alpha1.ComparedTo{
								Source: v1alpha1.ApplicationSource{
									Path:           &emptyString,
									TargetRevision: &emptyString,
									Chart:          &emptyString,
									Ref:            &emptyString,
								},
								Destination: v1alpha1.ApplicationDestination{
									Server:    &emptyString,
									Namespace: &emptyString,
									Name:      &emptyString,
								},
							},
						},
						Health: v1alpha1.HealthStatus{
							Status:  "Healthy",
							Message: &emptyString,
						},
						SourceType:           "",
						Summary:              v1alpha1.ApplicationSummary{},
						ResourceHealthSource: "",
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
		"SuccessfulAvailableWithAppNamespace": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().List(
						context.Background(),
						&argocdApplication.ApplicationQuery{
							Name:         &testApplicationExternalName,
							Projects:     []string{testProjectName},
							AppNamespace: &testAppNamespace,
						},
					).Return(
						&argocdv1alpha1.ApplicationList{
							TypeMeta: metav1.TypeMeta{},
							ListMeta: metav1.ListMeta{},
							Items: []argocdv1alpha1.Application{{
								TypeMeta: metav1.TypeMeta{},
								ObjectMeta: metav1.ObjectMeta{
									Name:        testApplicationExternalName,
									Namespace:   testAppNamespace,
									Annotations: testApplicationAnnotations,
									Finalizers:  testApplicationFinalizers,
								},
								Spec: argocdv1alpha1.ApplicationSpec{
									Project: testProjectName,
									Source: &argocdv1alpha1.ApplicationSource{
										RepoURL:        repoURL,
										Path:           chartPath,
										TargetRevision: revision,
									},
									Destination: argocdv1alpha1.ApplicationDestination{
										Namespace: testDestinationNamespace,
									},
									SyncPolicy: &argocdv1alpha1.SyncPolicy{
										Automated: &argocdv1alpha1.SyncPolicyAutomated{
											SelfHeal: true,
										},
									},
								},
								Status: argocdv1alpha1.ApplicationStatus{
									Health: argocdv1alpha1.HealthStatus{
										Status: "Healthy",
									},
									Sync: argocdv1alpha1.SyncStatus{
										Status: "Synced",
									},
								},
							}},
						}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project:      testProjectName,
						AppNamespace: &testAppNamespace,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Source: &v1alpha1.ApplicationSource{
							RepoURL:        repoURL,
							Path:           &chartPath,
							TargetRevision: &revision,
						},
						SyncPolicy: &v1alpha1.SyncPolicy{
							Automated: &v1alpha1.SyncPolicyAutomated{
								SelfHeal: &selfHealEnabled,
							},
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project:      testProjectName,
						AppNamespace: &testAppNamespace,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Source: &v1alpha1.ApplicationSource{
							RepoURL:        repoURL,
							Path:           &chartPath,
							TargetRevision: &revision,
						},
						SyncPolicy: &v1alpha1.SyncPolicy{
							Automated: &v1alpha1.SyncPolicyAutomated{
								SelfHeal: &selfHealEnabled,
							},
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.ArgoApplicationStatus{
						Resources: nil,
						Sync: v1alpha1.SyncStatus{
							Status:   "Synced",
							Revision: &emptyString,
							ComparedTo: v1alpha1.ComparedTo{
								Source: v1alpha1.ApplicationSource{
									Path:           &emptyString,
									TargetRevision: &emptyString,
									Chart:          &emptyString,
									Ref:            &emptyString,
								},
								Destination: v1alpha1.ApplicationDestination{
									Server:    &emptyString,
									Namespace: &emptyString,
									Name:      &emptyString,
								},
							},
						},
						Health: v1alpha1.HealthStatus{
							Status:  "Healthy",
							Message: &emptyString,
						},
						SourceType:           "",
						Summary:              v1alpha1.ApplicationSummary{},
						ResourceHealthSource: "",
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
		"SyncPolicyNotUpToDate": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().List(
						context.Background(),
						&argocdApplication.ApplicationQuery{
							Name:     &testApplicationExternalName,
							Projects: []string{testProjectName},
						},
					).Return(
						&argocdv1alpha1.ApplicationList{
							TypeMeta: metav1.TypeMeta{},
							ListMeta: metav1.ListMeta{},
							Items: []argocdv1alpha1.Application{{
								TypeMeta: metav1.TypeMeta{},
								ObjectMeta: metav1.ObjectMeta{
									Name:        testApplicationExternalName,
									Annotations: testApplicationAnnotations,
									Finalizers:  testApplicationFinalizers,
								},
								Spec: argocdv1alpha1.ApplicationSpec{
									Project: testProjectName,
									Source: &argocdv1alpha1.ApplicationSource{
										RepoURL:        repoURL,
										Path:           chartPath,
										TargetRevision: revision,
									},
									Destination: argocdv1alpha1.ApplicationDestination{
										Namespace: testDestinationNamespace,
									},
								},
								Status: argocdv1alpha1.ApplicationStatus{
									Health: argocdv1alpha1.HealthStatus{
										Status: "Missing",
									},
									Sync: argocdv1alpha1.SyncStatus{
										Status: "OutOfSync",
									},
								},
							}},
						}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project: testProjectName,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Source: &v1alpha1.ApplicationSource{
							RepoURL:        repoURL,
							Path:           &chartPath,
							TargetRevision: &revision,
						},
						SyncPolicy: &v1alpha1.SyncPolicy{
							Automated: &v1alpha1.SyncPolicyAutomated{
								SelfHeal: &selfHealEnabled,
							},
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project: testProjectName,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Source: &v1alpha1.ApplicationSource{
							RepoURL:        repoURL,
							Path:           &chartPath,
							TargetRevision: &revision,
						},
						SyncPolicy: &v1alpha1.SyncPolicy{
							Automated: &v1alpha1.SyncPolicyAutomated{
								SelfHeal: &selfHealEnabled,
							},
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
					withConditions(xpv1.Unavailable()),
					withObservation(initializedArgoAppStatus()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
				err: nil,
			},
		},
		"ListApplicationFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().List(
						context.Background(),
						&argocdApplication.ApplicationQuery{
							Name: &testApplicationExternalName,
						},
					).Return(
						nil, errBoom)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
				),
				err: errors.Wrap(errBoom, errListFailed),
			},
		},
		"NeedsCreation": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().List(
						context.Background(),
						&argocdApplication.ApplicationQuery{
							Name: &testApplicationExternalName,
						},
					).Return(
						&argocdv1alpha1.ApplicationList{
							TypeMeta: metav1.TypeMeta{},
							ListMeta: metav1.ListMeta{},
							Items:    []argocdv1alpha1.Application{},
						}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
				),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"NoExternalName -> NeedsCreation": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {}),
				cr: Application(
					withName(testApplicationExternalName),
				),
			},
			want: want{
				cr: Application(
					withName(testApplicationExternalName),
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

func initializedArgoAppStatus() v1alpha1.ArgoApplicationStatus {
	return v1alpha1.ArgoApplicationStatus{
		Resources: nil,
		Sync: v1alpha1.SyncStatus{
			Status:   "OutOfSync",
			Revision: &emptyString,
			ComparedTo: v1alpha1.ComparedTo{
				Source: v1alpha1.ApplicationSource{
					Path:           &emptyString,
					TargetRevision: &emptyString,
					Chart:          &emptyString,
					Ref:            &emptyString,
				},
				Destination: v1alpha1.ApplicationDestination{
					Server:    &emptyString,
					Namespace: &emptyString,
					Name:      &emptyString,
				},
			},
		},
		Health: v1alpha1.HealthStatus{
			Status:  "Missing",
			Message: &emptyString,
		},
		SourceType:           "",
		Summary:              v1alpha1.ApplicationSummary{},
		ResourceHealthSource: "",
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Application
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
						&argocdApplication.ApplicationCreateRequest{
							Application: &argocdv1alpha1.Application{
								ObjectMeta: metav1.ObjectMeta{
									Name: testApplicationExternalName,
								},
							},
						},
					).Return(
						&argocdv1alpha1.Application{
							ObjectMeta: metav1.ObjectMeta{
								Name: testApplicationExternalName,
							},
						}, nil)
				}),
				cr: Application(withExternalName(testApplicationExternalName), withName(testApplicationExternalName)),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
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
						&argocdApplication.ApplicationCreateRequest{
							Application: &argocdv1alpha1.Application{
								ObjectMeta: metav1.ObjectMeta{
									Name: testApplicationExternalName,
								},
							},
						},
					).Return(
						nil, errBoom)
				}),
				cr: Application(withExternalName(testApplicationExternalName), withName(testApplicationExternalName)),
			},
			want: want{
				cr:     Application(withExternalName(testApplicationExternalName), withName(testApplicationExternalName)),
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
		cr     *v1alpha1.Application
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
						&argocdApplication.ApplicationUpdateRequest{
							Application: &argocdv1alpha1.Application{
								ObjectMeta: metav1.ObjectMeta{
									Name:        testApplicationExternalName,
									Annotations: testApplicationAnnotations,
									Finalizers:  testApplicationFinalizers,
								},
								Spec: argocdv1alpha1.ApplicationSpec{
									Project: testProjectName,
									Destination: argocdv1alpha1.ApplicationDestination{
										Namespace: testDestinationNamespace,
									},
								},
							},
						},
					).Return(&argocdv1alpha1.Application{
						ObjectMeta: metav1.ObjectMeta{
							Name:        testApplicationExternalName,
							Annotations: testApplicationAnnotations,
							Finalizers:  testApplicationFinalizers,
						},
					}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project: testProjectName,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
					withExternalName(testApplicationExternalName),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project: testProjectName,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
					withExternalName(testApplicationExternalName),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"SuccessfulWithAppNamespace": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Update(
						context.Background(),
						&argocdApplication.ApplicationUpdateRequest{
							Application: &argocdv1alpha1.Application{
								ObjectMeta: metav1.ObjectMeta{
									Name:        testApplicationExternalName,
									Annotations: testApplicationAnnotations,
									Finalizers:  testApplicationFinalizers,
									Namespace:   testAppNamespace,
								},
								Spec: argocdv1alpha1.ApplicationSpec{
									Project: testProjectName,
									Destination: argocdv1alpha1.ApplicationDestination{
										Namespace: testDestinationNamespace,
									},
								},
							},
						},
					).Return(&argocdv1alpha1.Application{
						ObjectMeta: metav1.ObjectMeta{
							Name:        testApplicationExternalName,
							Annotations: testApplicationAnnotations,
							Finalizers:  testApplicationFinalizers,
							Namespace:   testAppNamespace,
						},
					}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project:      testProjectName,
						AppNamespace: &testAppNamespace,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
					withExternalName(testApplicationExternalName),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project:      testProjectName,
						AppNamespace: &testAppNamespace,
						Destination: v1alpha1.ApplicationDestination{
							Namespace: &testDestinationNamespace,
						},
						Annotations: testApplicationAnnotations,
						Finalizers:  testApplicationFinalizers,
					}),
					withExternalName(testApplicationExternalName),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"UpdateFailed": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Update(
						context.Background(),
						&argocdApplication.ApplicationUpdateRequest{
							Application: &argocdv1alpha1.Application{
								ObjectMeta: metav1.ObjectMeta{
									Name:        testApplicationExternalName,
									Annotations: testApplicationAnnotations,
								},
								Spec: argocdv1alpha1.ApplicationSpec{
									Project: testProjectName,
								},
							},
						},
					).Return(nil, errBoom)
				}),
				cr: Application(
					withSpec(v1alpha1.ApplicationParameters{
						Project:     testProjectName,
						Annotations: testApplicationAnnotations,
					}),
					withExternalName(testApplicationExternalName),
				),
			},
			want: want{
				cr: Application(
					withSpec(v1alpha1.ApplicationParameters{
						Project:     testProjectName,
						Annotations: testApplicationAnnotations,
					}),
					withExternalName(testApplicationExternalName),
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
		cr  *v1alpha1.Application
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
						&argocdApplication.ApplicationDeleteRequest{
							Name: &testApplicationExternalName,
						},
					).Return(&argocdApplication.ApplicationResponse{}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
				),
				err: nil,
			},
		},
		"SuccessfulWithProject": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Delete(
						context.Background(),
						&argocdApplication.ApplicationDeleteRequest{
							Name:    &testApplicationExternalName,
							Project: &testProjectName,
						},
					).Return(&argocdApplication.ApplicationResponse{}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project: testProjectName,
					}),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project: testProjectName,
					}),
				),
				err: nil,
			},
		},
		"SuccessfulWithAppNamespace": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Delete(
						context.Background(),
						&argocdApplication.ApplicationDeleteRequest{
							Name:         &testApplicationExternalName,
							AppNamespace: &testAppNamespace,
						},
					).Return(&argocdApplication.ApplicationResponse{}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
					withAppNamespace(&testAppNamespace),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
					withAppNamespace(&testAppNamespace),
				),
				err: nil,
			},
		},
		"SuccessfulWithProjectAndAppNamespace": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockServiceClient) {
					mcs.EXPECT().Delete(
						context.Background(),
						&argocdApplication.ApplicationDeleteRequest{
							Name:         &testApplicationExternalName,
							Project:      &testProjectName,
							AppNamespace: &testAppNamespace,
						},
					).Return(&argocdApplication.ApplicationResponse{}, nil)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project:      testProjectName,
						AppNamespace: &testAppNamespace,
					}),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
					withSpec(v1alpha1.ApplicationParameters{
						Project:      testProjectName,
						AppNamespace: &testAppNamespace,
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
						&argocdApplication.ApplicationDeleteRequest{
							Name: &testApplicationExternalName,
						},
					).Return(&argocdApplication.ApplicationResponse{}, errBoom)
				}),
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
				),
			},
			want: want{
				cr: Application(
					withExternalName(testApplicationExternalName),
					withName(testApplicationExternalName),
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
