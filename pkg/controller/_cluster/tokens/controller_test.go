package tokens

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/argoproj/argo-cd/v3/pkg/apiclient/project"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-argocd/apis/cluster/projects/v1alpha1"
	mockclient "github.com/crossplane-contrib/provider-argocd/pkg/clients/mock/projects"
	"github.com/crossplane-contrib/provider-argocd/pkg/clients/projects"
)

var (
	testProjectName              = "test-project"
	testRoleName                 = "test-role"
	testTokenExternalName        = "test-token"
	testExpiresInZero      int64 = 0
	testExpiresInOneMinute int64 = 60
	testIssuedAt           int64 = 1
	errBoom                      = errors.New("boom")
	errProjectNotFound           = errors.New("code = NotFound desc = appprojects")
	testJWTHeaderJSON            = `{"alg":"HS256","typ":"JWT"}`
	testJWTPayloadJSON           = `{"jti":"test-token","iss":"test-issuer"}`
)

type args struct {
	client projects.ProjectServiceClient
	cr     *v1alpha1.Token
}

type mockModifier func(*mockclient.MockProjectServiceClient)

func withMockClient(t *testing.T, mod mockModifier) *mockclient.MockProjectServiceClient {
	ctrl := gomock.NewController(t)
	mock := mockclient.NewMockProjectServiceClient(ctrl)
	mod(mock)
	return mock
}

func Token(m ...TokenModifier) *v1alpha1.Token {
	cr := &v1alpha1.Token{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type TokenModifier func(*v1alpha1.Token)

func withExternalName(v string) TokenModifier {
	return func(s *v1alpha1.Token) {
		meta.SetExternalName(s, v)
	}
}

func withSpec(p v1alpha1.TokenParameters) TokenModifier {
	return func(r *v1alpha1.Token) { r.Spec.ForProvider = p }
}

func withObservation(p v1alpha1.TokenObservation) TokenModifier {
	return func(r *v1alpha1.Token) { r.Status.AtProvider = p }
}

func withConditions(c ...xpv1.Condition) TokenModifier {
	return func(r *v1alpha1.Token) { r.Status.ConditionedStatus.Conditions = c }
}

func createTestJWTToken() string {
	header := base64.RawURLEncoding.EncodeToString([]byte(testJWTHeaderJSON))
	payload := base64.RawURLEncoding.EncodeToString([]byte(testJWTPayloadJSON))
	signature := "test-signature"
	return fmt.Sprintf("%s.%s.%s", header, payload, signature)
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Token
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
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: testProjectName,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name: testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{
											{
												IssuedAt:  testIssuedAt,
												ExpiresAt: testExpiresInZero,
												ID:        testTokenExternalName,
											},
										},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:      testTokenExternalName,
						Project: &testProjectName,
						Role:    testRoleName,
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:      testTokenExternalName,
						Project: &testProjectName,
						Role:    testRoleName,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.TokenObservation{
						IssuedAt:  testIssuedAt,
						ExpiresAt: &testExpiresInZero,
						ID:        &testTokenExternalName,
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
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: testProjectName,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name: testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{
											{
												IssuedAt:  testIssuedAt,
												ExpiresAt: testExpiresInZero,
												ID:        testTokenExternalName,
											},
										},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:      testTokenExternalName,
						Project: &testProjectName,
						Role:    testRoleName,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.TokenObservation{
						IssuedAt:  testIssuedAt,
						ExpiresAt: &testExpiresInZero,
						ID:        &testTokenExternalName,
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
		"ExpireNotUpdated": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: testProjectName,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name: testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{
											{
												IssuedAt:  testIssuedAt,
												ExpiresAt: testExpiresInZero,
												ID:        testTokenExternalName,
											},
										},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:      testTokenExternalName,
						Project: &testProjectName,
						Role:    testRoleName,
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.TokenObservation{
						IssuedAt:  testIssuedAt,
						ExpiresAt: &testExpiresInZero,
						ID:        &testTokenExternalName,
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
		"NeedsUpdateDueToExpiration": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name: testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{
											{
												IssuedAt:  time.Now().Add(-50 * time.Minute).Unix(),
												ExpiresAt: time.Now().Add(-1 * time.Minute).Unix(),
												ID:        testTokenExternalName,
											},
										},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:        testTokenExternalName,
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1h"),
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:        testTokenExternalName,
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1h"),
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.TokenObservation{
						IssuedAt:  time.Now().Add(-50 * time.Minute).Unix(),
						ExpiresAt: ptr.To(time.Now().Add(-1 * time.Minute).Unix()),
						ID:        &testTokenExternalName,
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
		"NeedsUpdateDueToRenewBefore": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name: testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{
											{
												IssuedAt:  time.Now().Add(-50 * time.Minute).Unix(),
												ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
												ID:        testTokenExternalName,
											},
										},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:          testTokenExternalName,
						Project:     &testProjectName,
						Role:        testRoleName,
						ExpiresIn:   ptr.To("1h"),
						RenewBefore: ptr.To("10m"),
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:          testTokenExternalName,
						Project:     &testProjectName,
						Role:        testRoleName,
						ExpiresIn:   ptr.To("1h"),
						RenewBefore: ptr.To("10m"),
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.TokenObservation{
						IssuedAt:  time.Now().Add(-50 * time.Minute).Unix(),
						ExpiresAt: ptr.To(time.Now().Add(5 * time.Minute).Unix()),
						ID:        &testTokenExternalName,
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
		"NeedsUpdateDueToRenewAfter": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name: testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{
											{
												IssuedAt:  time.Now().Add(-30 * time.Minute).Unix(),
												ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
												ID:        testTokenExternalName,
											},
										},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:         testTokenExternalName,
						Project:    &testProjectName,
						Role:       testRoleName,
						ExpiresIn:  ptr.To("1h"),
						RenewAfter: ptr.To("20m"),
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:         testTokenExternalName,
						Project:    &testProjectName,
						Role:       testRoleName,
						ExpiresIn:  ptr.To("1h"),
						RenewAfter: ptr.To("20m"),
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.TokenObservation{
						IssuedAt:  time.Now().Add(-30 * time.Minute).Unix(),
						ExpiresAt: ptr.To(time.Now().Add(30 * time.Minute).Unix()),
						ID:        &testTokenExternalName,
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
		"NeedsUpdateDueToExpirationChange": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name: testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{
											{
												IssuedAt:  time.Now().Unix(),
												ExpiresAt: testExpiresInZero,
												ID:        testTokenExternalName,
											},
										},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:        testTokenExternalName,
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1h"),
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:        testTokenExternalName,
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1h"),
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.TokenObservation{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: &testExpiresInZero,
						ID:        &testTokenExternalName,
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
		"NeedsUpdateDueToExpirationRemoval": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name: testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{
											{
												IssuedAt:  time.Now().Unix(),
												ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
												ID:        testTokenExternalName,
											},
										},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:        testTokenExternalName,
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("0"),
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						ID:        testTokenExternalName,
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("0"),
					}),
					withConditions(xpv1.Available()),
					withObservation(v1alpha1.TokenObservation{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: ptr.To(time.Now().Add(1 * time.Hour).Unix()),
						ID:        &testTokenExternalName,
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
							Name: testProjectName,
						},
					).Return(
						nil, errBoom)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
					}),
				),
				err: errors.Wrap(errBoom, errGetProjectFailed),
			},
		},
		"GetProjectNotFound": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						nil, errProjectNotFound)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errProjectNotFound, errGetProjectFailed),
			},
		},
		"GetRoleNotFound": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							TypeMeta: metav1.TypeMeta{},
							ObjectMeta: metav1.ObjectMeta{
								Name: testProjectName,
							},
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: nil,
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
				),
				err: errors.Wrap(
					fmt.Errorf("role '%s' does not exist in project '%s'", testRoleName, testProjectName),
					errGetRoleFailed,
				),
			},
		},
		"TokenNotFound": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().Get(
						context.Background(),
						&project.ProjectQuery{
							Name: testProjectName,
						},
					).Return(
						&argocdv1alpha1.AppProject{
							Spec: argocdv1alpha1.AppProjectSpec{
								Roles: []argocdv1alpha1.ProjectRole{
									{
										Name:      testRoleName,
										JWTTokens: []argocdv1alpha1.JWTToken{},
									},
								},
							},
						}, nil)
				}),
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
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
		cr     *v1alpha1.Token
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulNoExpire": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().CreateToken(
						context.Background(),
						&project.ProjectTokenCreateRequest{
							Project:   testProjectName,
							Role:      testRoleName,
							ExpiresIn: testExpiresInZero,
						},
					).Return(
						&project.ProjectTokenResponse{
							Token: createTestJWTToken(),
						}, nil)
				}),
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("0"),
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("0"),
					}),
				),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"SuccessfulExpire": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().CreateToken(
						context.Background(),
						&project.ProjectTokenCreateRequest{
							Project:   testProjectName,
							Role:      testRoleName,
							ExpiresIn: testExpiresInOneMinute,
						},
					).Return(
						&project.ProjectTokenResponse{
							Token: createTestJWTToken(),
						}, nil)
				}),
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1m"),
					}),
				),
			},
			want: want{
				cr: Token(
					withExternalName(testTokenExternalName),
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1m"),
					}),
				),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"CreateError": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().CreateToken(
						context.Background(),
						gomock.Any(),
					).Return(nil, errBoom)
				}),
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("0"),
					}),
				),
			},
			want: want{
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("0"),
					}),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errBoom, errCreateTokenFailed),
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
		cr     *v1alpha1.Token
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdate": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().DeleteToken(
						context.Background(),
						&project.ProjectTokenDeleteRequest{
							Project: testProjectName,
							Role:    testRoleName,
							Id:      testTokenExternalName,
						},
					).Return(&project.EmptyResponse{}, nil)
					mcs.EXPECT().CreateToken(
						context.Background(),
						&project.ProjectTokenCreateRequest{
							Project:   testProjectName,
							Role:      testRoleName,
							ExpiresIn: testExpiresInOneMinute,
						},
					).Return(
						&project.ProjectTokenResponse{
							Token: createTestJWTToken(),
						}, nil)
				}),
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1m"),
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
			},
			want: want{
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1m"),
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"DeleteError": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().DeleteToken(
						context.Background(),
						&project.ProjectTokenDeleteRequest{
							Project: testProjectName,
							Role:    testRoleName,
							Id:      testTokenExternalName,
						},
					).Return(nil, errBoom)
				}),
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1m"),
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
			},
			want: want{
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1m"),
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errBoom, errDeleteFailed),
			},
		},
		"CreateError": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().DeleteToken(
						context.Background(),
						&project.ProjectTokenDeleteRequest{
							Project: testProjectName,
							Role:    testRoleName,
							Id:      testTokenExternalName,
						},
					).Return(&project.EmptyResponse{}, nil)
					mcs.EXPECT().CreateToken(
						context.Background(),
						&project.ProjectTokenCreateRequest{
							Project:   testProjectName,
							Role:      testRoleName,
							ExpiresIn: testExpiresInOneMinute,
						},
					).Return(nil, errBoom)
				}),
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1m"),
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
			},
			want: want{
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project:   &testProjectName,
						Role:      testRoleName,
						ExpiresIn: ptr.To("1m"),
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errBoom, errCreateTokenFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			o, err := e.Update(context.Background(), tc.args.cr)

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

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.Token
		err error
		res managed.ExternalDelete
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDelete": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().DeleteToken(
						context.Background(),
						&project.ProjectTokenDeleteRequest{
							Project: testProjectName,
							Role:    testRoleName,
							Id:      testTokenExternalName,
						},
					).Return(&project.EmptyResponse{}, nil)
				}),
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
			},
			want: want{
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
				err: nil,
			},
		},
		"DeleteError": {
			args: args{
				client: withMockClient(t, func(mcs *mockclient.MockProjectServiceClient) {
					mcs.EXPECT().DeleteToken(
						context.Background(),
						&project.ProjectTokenDeleteRequest{
							Project: testProjectName,
							Role:    testRoleName,
							Id:      testTokenExternalName,
						},
					).Return(nil, errBoom)
				}),
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
					}),
				),
			},
			want: want{
				cr: Token(
					withSpec(v1alpha1.TokenParameters{
						Project: &testProjectName,
						Role:    testRoleName,
					}),
					withObservation(v1alpha1.TokenObservation{
						ID: &testTokenExternalName,
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

			if diff := cmp.Diff(tc.want.res, got, test.EquateErrors()); diff != "" {
				t.Errorf("res: -want, +got:\n%s", diff)
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
