// Code generated by MockGen. DO NOT EDIT.
// Source: ./actpool/actpool.go

// Package mock_actpool is a generated GoMock package.
package mock_actpool

import (
	gomock "github.com/golang/mock/gomock"
	action "github.com/iotexproject/iotex-core/blockchain/action"
	hash "github.com/iotexproject/iotex-core/pkg/hash"
	proto "github.com/iotexproject/iotex-core/proto"
	reflect "reflect"
)

// MockActPool is a mock of ActPool interface
type MockActPool struct {
	ctrl     *gomock.Controller
	recorder *MockActPoolMockRecorder
}

// MockActPoolMockRecorder is the mock recorder for MockActPool
type MockActPoolMockRecorder struct {
	mock *MockActPool
}

// NewMockActPool creates a new mock instance
func NewMockActPool(ctrl *gomock.Controller) *MockActPool {
	mock := &MockActPool{ctrl: ctrl}
	mock.recorder = &MockActPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockActPool) EXPECT() *MockActPoolMockRecorder {
	return m.recorder
}

// Reset mocks base method
func (m *MockActPool) Reset() {
	m.ctrl.Call(m, "Reset")
}

// Reset indicates an expected call of Reset
func (mr *MockActPoolMockRecorder) Reset() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reset", reflect.TypeOf((*MockActPool)(nil).Reset))
}

// PickActs mocks base method
func (m *MockActPool) PickActs() ([]*action.Transfer, []*action.Vote, []*action.Execution) {
	ret := m.ctrl.Call(m, "PickActs")
	ret0, _ := ret[0].([]*action.Transfer)
	ret1, _ := ret[1].([]*action.Vote)
	ret2, _ := ret[2].([]*action.Execution)
	return ret0, ret1, ret2
}

// PickActs indicates an expected call of PickActs
func (mr *MockActPoolMockRecorder) PickActs() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PickActs", reflect.TypeOf((*MockActPool)(nil).PickActs))
}

// AddTsf mocks base method
func (m *MockActPool) AddTsf(tsf *action.Transfer) error {
	ret := m.ctrl.Call(m, "AddTsf", tsf)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddTsf indicates an expected call of AddTsf
func (mr *MockActPoolMockRecorder) AddTsf(tsf interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTsf", reflect.TypeOf((*MockActPool)(nil).AddTsf), tsf)
}

// AddVote mocks base method
func (m *MockActPool) AddVote(vote *action.Vote) error {
	ret := m.ctrl.Call(m, "AddVote", vote)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddVote indicates an expected call of AddVote
func (mr *MockActPoolMockRecorder) AddVote(vote interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddVote", reflect.TypeOf((*MockActPool)(nil).AddVote), vote)
}

// AddExecution mocks base method
func (m *MockActPool) AddExecution(execution *action.Execution) error {
	ret := m.ctrl.Call(m, "AddExecution", execution)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddExecution indicates an expected call of AddExecution
func (mr *MockActPoolMockRecorder) AddExecution(execution interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddExecution", reflect.TypeOf((*MockActPool)(nil).AddExecution), execution)
}

// GetPendingNonce mocks base method
func (m *MockActPool) GetPendingNonce(addr string) (uint64, error) {
	ret := m.ctrl.Call(m, "GetPendingNonce", addr)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPendingNonce indicates an expected call of GetPendingNonce
func (mr *MockActPoolMockRecorder) GetPendingNonce(addr interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPendingNonce", reflect.TypeOf((*MockActPool)(nil).GetPendingNonce), addr)
}

// GetUnconfirmedActs mocks base method
func (m *MockActPool) GetUnconfirmedActs(addr string) []*proto.ActionPb {
	ret := m.ctrl.Call(m, "GetUnconfirmedActs", addr)
	ret0, _ := ret[0].([]*proto.ActionPb)
	return ret0
}

// GetUnconfirmedActs indicates an expected call of GetUnconfirmedActs
func (mr *MockActPoolMockRecorder) GetUnconfirmedActs(addr interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUnconfirmedActs", reflect.TypeOf((*MockActPool)(nil).GetUnconfirmedActs), addr)
}

// GetActionByHash mocks base method
func (m *MockActPool) GetActionByHash(hash hash.Hash32B) (*proto.ActionPb, error) {
	ret := m.ctrl.Call(m, "GetActionByHash", hash)
	ret0, _ := ret[0].(*proto.ActionPb)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetActionByHash indicates an expected call of GetActionByHash
func (mr *MockActPoolMockRecorder) GetActionByHash(hash interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActionByHash", reflect.TypeOf((*MockActPool)(nil).GetActionByHash), hash)
}

// GetSize mocks base method
func (m *MockActPool) GetSize() uint64 {
	ret := m.ctrl.Call(m, "GetSize")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetSize indicates an expected call of GetSize
func (mr *MockActPoolMockRecorder) GetSize() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSize", reflect.TypeOf((*MockActPool)(nil).GetSize))
}

// GetCapacity mocks base method
func (m *MockActPool) GetCapacity() uint64 {
	ret := m.ctrl.Call(m, "GetCapacity")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetCapacity indicates an expected call of GetCapacity
func (mr *MockActPoolMockRecorder) GetCapacity() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCapacity", reflect.TypeOf((*MockActPool)(nil).GetCapacity))
}

// GetUnconfirmedActSize mocks base method
func (m *MockActPool) GetUnconfirmedActSize() uint64 {
	ret := m.ctrl.Call(m, "GetUnconfirmedActSize")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// GetUnconfirmedActSize indicates an expected call of GetUnconfirmedActSize
func (mr *MockActPoolMockRecorder) GetUnconfirmedActSize() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUnconfirmedActSize", reflect.TypeOf((*MockActPool)(nil).GetUnconfirmedActSize))
}
