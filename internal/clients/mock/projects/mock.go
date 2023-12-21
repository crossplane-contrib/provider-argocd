// Code generated by MockGen. DO NOT EDIT.
// Source: ../projects/client.go

// Package projects is a generated GoMock package.
package projects

import (
	context "context"
	reflect "reflect"

	project "github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockProjectServiceClient is a mock of ProjectServiceClient interface.
type MockProjectServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockProjectServiceClientMockRecorder
}

// MockProjectServiceClientMockRecorder is the mock recorder for MockProjectServiceClient.
type MockProjectServiceClientMockRecorder struct {
	mock *MockProjectServiceClient
}

// NewMockProjectServiceClient creates a new mock instance.
func NewMockProjectServiceClient(ctrl *gomock.Controller) *MockProjectServiceClient {
	mock := &MockProjectServiceClient{ctrl: ctrl}
	mock.recorder = &MockProjectServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProjectServiceClient) EXPECT() *MockProjectServiceClientMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockProjectServiceClient) Create(ctx context.Context, in *project.ProjectCreateRequest, opts ...grpc.CallOption) (*v1alpha1.AppProject, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Create", varargs...)
	ret0, _ := ret[0].(*v1alpha1.AppProject)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockProjectServiceClientMockRecorder) Create(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockProjectServiceClient)(nil).Create), varargs...)
}

// Delete mocks base method.
func (m *MockProjectServiceClient) Delete(ctx context.Context, in *project.ProjectQuery, opts ...grpc.CallOption) (*project.EmptyResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Delete", varargs...)
	ret0, _ := ret[0].(*project.EmptyResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockProjectServiceClientMockRecorder) Delete(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockProjectServiceClient)(nil).Delete), varargs...)
}

// Get mocks base method.
func (m *MockProjectServiceClient) Get(ctx context.Context, in *project.ProjectQuery, opts ...grpc.CallOption) (*v1alpha1.AppProject, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Get", varargs...)
	ret0, _ := ret[0].(*v1alpha1.AppProject)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockProjectServiceClientMockRecorder) Get(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockProjectServiceClient)(nil).Get), varargs...)
}

// Update mocks base method.
func (m *MockProjectServiceClient) Update(ctx context.Context, in *project.ProjectUpdateRequest, opts ...grpc.CallOption) (*v1alpha1.AppProject, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(*v1alpha1.AppProject)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockProjectServiceClientMockRecorder) Update(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockProjectServiceClient)(nil).Update), varargs...)
}