// Code generated by mockery v2.42.0. DO NOT EDIT.

package mocks

import (
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	mock "github.com/stretchr/testify/mock"

	types "github.com/cosmos/cosmos-sdk/types"
)

// DistrKeeper is an autogenerated mock type for the DistrKeeper type
type DistrKeeper struct {
	mock.Mock
}

type DistrKeeper_Expecter struct {
	mock *mock.Mock
}

func (_m *DistrKeeper) EXPECT() *DistrKeeper_Expecter {
	return &DistrKeeper_Expecter{mock: &_m.Mock}
}

// GetFeePool provides a mock function with given fields: ctx
func (_m *DistrKeeper) GetFeePool(ctx types.Context) distributiontypes.FeePool {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetFeePool")
	}

	var r0 distributiontypes.FeePool
	if rf, ok := ret.Get(0).(func(types.Context) distributiontypes.FeePool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(distributiontypes.FeePool)
	}

	return r0
}

// DistrKeeper_GetFeePool_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFeePool'
type DistrKeeper_GetFeePool_Call struct {
	*mock.Call
}

// GetFeePool is a helper method to define mock.On call
//   - ctx types.Context
func (_e *DistrKeeper_Expecter) GetFeePool(ctx interface{}) *DistrKeeper_GetFeePool_Call {
	return &DistrKeeper_GetFeePool_Call{Call: _e.mock.On("GetFeePool", ctx)}
}

func (_c *DistrKeeper_GetFeePool_Call) Run(run func(ctx types.Context)) *DistrKeeper_GetFeePool_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(types.Context))
	})
	return _c
}

func (_c *DistrKeeper_GetFeePool_Call) Return(_a0 distributiontypes.FeePool) *DistrKeeper_GetFeePool_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DistrKeeper_GetFeePool_Call) RunAndReturn(run func(types.Context) distributiontypes.FeePool) *DistrKeeper_GetFeePool_Call {
	_c.Call.Return(run)
	return _c
}

// SetFeePool provides a mock function with given fields: ctx, pool
func (_m *DistrKeeper) SetFeePool(ctx types.Context, pool distributiontypes.FeePool) {
	_m.Called(ctx, pool)
}

// DistrKeeper_SetFeePool_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetFeePool'
type DistrKeeper_SetFeePool_Call struct {
	*mock.Call
}

// SetFeePool is a helper method to define mock.On call
//   - ctx types.Context
//   - pool distributiontypes.FeePool
func (_e *DistrKeeper_Expecter) SetFeePool(ctx interface{}, pool interface{}) *DistrKeeper_SetFeePool_Call {
	return &DistrKeeper_SetFeePool_Call{Call: _e.mock.On("SetFeePool", ctx, pool)}
}

func (_c *DistrKeeper_SetFeePool_Call) Run(run func(ctx types.Context, pool distributiontypes.FeePool)) *DistrKeeper_SetFeePool_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(types.Context), args[1].(distributiontypes.FeePool))
	})
	return _c
}

func (_c *DistrKeeper_SetFeePool_Call) Return() *DistrKeeper_SetFeePool_Call {
	_c.Call.Return()
	return _c
}

func (_c *DistrKeeper_SetFeePool_Call) RunAndReturn(run func(types.Context, distributiontypes.FeePool)) *DistrKeeper_SetFeePool_Call {
	_c.Call.Return(run)
	return _c
}

// NewDistrKeeper creates a new instance of DistrKeeper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDistrKeeper(t interface {
	mock.TestingT
	Cleanup(func())
}) *DistrKeeper {
	mock := &DistrKeeper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
