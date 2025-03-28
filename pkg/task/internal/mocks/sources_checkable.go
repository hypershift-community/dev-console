// Code generated by mockery v2.24.0. DO NOT EDIT.

package mocks

import (
	ast "github.com/hypershift-community/hyper-console/pkg/task/taskfile/ast"

	mock "github.com/stretchr/testify/mock"
)

// SourcesCheckable is an autogenerated mock type for the SourcesCheckable type
type SourcesCheckable struct {
	mock.Mock
}

type SourcesCheckable_Expecter struct {
	mock *mock.Mock
}

func (_m *SourcesCheckable) EXPECT() *SourcesCheckable_Expecter {
	return &SourcesCheckable_Expecter{mock: &_m.Mock}
}

// IsUpToDate provides a mock function with given fields: t
func (_m *SourcesCheckable) IsUpToDate(t *ast.Task) (bool, error) {
	ret := _m.Called(t)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(*ast.Task) (bool, error)); ok {
		return rf(t)
	}
	if rf, ok := ret.Get(0).(func(*ast.Task) bool); ok {
		r0 = rf(t)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(*ast.Task) error); ok {
		r1 = rf(t)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SourcesCheckable_IsUpToDate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsUpToDate'
type SourcesCheckable_IsUpToDate_Call struct {
	*mock.Call
}

// IsUpToDate is a helper method to define mock.On call
//   - t *ast.Task
func (_e *SourcesCheckable_Expecter) IsUpToDate(t interface{}) *SourcesCheckable_IsUpToDate_Call {
	return &SourcesCheckable_IsUpToDate_Call{Call: _e.mock.On("IsUpToDate", t)}
}

func (_c *SourcesCheckable_IsUpToDate_Call) Run(run func(t *ast.Task)) *SourcesCheckable_IsUpToDate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*ast.Task))
	})
	return _c
}

func (_c *SourcesCheckable_IsUpToDate_Call) Return(_a0 bool, _a1 error) *SourcesCheckable_IsUpToDate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SourcesCheckable_IsUpToDate_Call) RunAndReturn(run func(*ast.Task) (bool, error)) *SourcesCheckable_IsUpToDate_Call {
	_c.Call.Return(run)
	return _c
}

// Kind provides a mock function with given fields:
func (_m *SourcesCheckable) Kind() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SourcesCheckable_Kind_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Kind'
type SourcesCheckable_Kind_Call struct {
	*mock.Call
}

// Kind is a helper method to define mock.On call
func (_e *SourcesCheckable_Expecter) Kind() *SourcesCheckable_Kind_Call {
	return &SourcesCheckable_Kind_Call{Call: _e.mock.On("Kind")}
}

func (_c *SourcesCheckable_Kind_Call) Run(run func()) *SourcesCheckable_Kind_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SourcesCheckable_Kind_Call) Return(_a0 string) *SourcesCheckable_Kind_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SourcesCheckable_Kind_Call) RunAndReturn(run func() string) *SourcesCheckable_Kind_Call {
	_c.Call.Return(run)
	return _c
}

// OnError provides a mock function with given fields: t
func (_m *SourcesCheckable) OnError(t *ast.Task) error {
	ret := _m.Called(t)

	var r0 error
	if rf, ok := ret.Get(0).(func(*ast.Task) error); ok {
		r0 = rf(t)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SourcesCheckable_OnError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnError'
type SourcesCheckable_OnError_Call struct {
	*mock.Call
}

// OnError is a helper method to define mock.On call
//   - t *ast.Task
func (_e *SourcesCheckable_Expecter) OnError(t interface{}) *SourcesCheckable_OnError_Call {
	return &SourcesCheckable_OnError_Call{Call: _e.mock.On("OnError", t)}
}

func (_c *SourcesCheckable_OnError_Call) Run(run func(t *ast.Task)) *SourcesCheckable_OnError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*ast.Task))
	})
	return _c
}

func (_c *SourcesCheckable_OnError_Call) Return(_a0 error) *SourcesCheckable_OnError_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SourcesCheckable_OnError_Call) RunAndReturn(run func(*ast.Task) error) *SourcesCheckable_OnError_Call {
	_c.Call.Return(run)
	return _c
}

// Value provides a mock function with given fields: t
func (_m *SourcesCheckable) Value(t *ast.Task) (interface{}, error) {
	ret := _m.Called(t)

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(*ast.Task) (interface{}, error)); ok {
		return rf(t)
	}
	if rf, ok := ret.Get(0).(func(*ast.Task) interface{}); ok {
		r0 = rf(t)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(*ast.Task) error); ok {
		r1 = rf(t)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SourcesCheckable_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type SourcesCheckable_Value_Call struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
//   - t *ast.Task
func (_e *SourcesCheckable_Expecter) Value(t interface{}) *SourcesCheckable_Value_Call {
	return &SourcesCheckable_Value_Call{Call: _e.mock.On("Value", t)}
}

func (_c *SourcesCheckable_Value_Call) Run(run func(t *ast.Task)) *SourcesCheckable_Value_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*ast.Task))
	})
	return _c
}

func (_c *SourcesCheckable_Value_Call) Return(_a0 interface{}, _a1 error) *SourcesCheckable_Value_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SourcesCheckable_Value_Call) RunAndReturn(run func(*ast.Task) (interface{}, error)) *SourcesCheckable_Value_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewSourcesCheckable interface {
	mock.TestingT
	Cleanup(func())
}

// NewSourcesCheckable creates a new instance of SourcesCheckable. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSourcesCheckable(t mockConstructorTestingTNewSourcesCheckable) *SourcesCheckable {
	mock := &SourcesCheckable{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
