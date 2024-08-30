// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package iibcchannelupgradablemodule

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// HeightData is an auto generated low-level Go binding around an user-defined struct.
type HeightData struct {
	RevisionNumber uint64
	RevisionHeight uint64
}

// IIBCChannelUpgradableModuleUpgradeProposal is an auto generated low-level Go binding around an user-defined struct.
type IIBCChannelUpgradableModuleUpgradeProposal struct {
	Fields  UpgradeFieldsData
	Timeout TimeoutData
}

// TimeoutData is an auto generated low-level Go binding around an user-defined struct.
type TimeoutData struct {
	Height    HeightData
	Timestamp uint64
}

// UpgradeFieldsData is an auto generated low-level Go binding around an user-defined struct.
type UpgradeFieldsData struct {
	Ordering       uint8
	ConnectionHops []string
	Version        string
}

// IibcchannelupgradablemoduleMetaData contains all meta data concerning the Iibcchannelupgradablemodule contract.
var IibcchannelupgradablemoduleMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"allowTransitionToFlushComplete\",\"inputs\":[{\"name\":\"portId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"channelId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"upgradeSequence\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getUpgradeProposal\",\"inputs\":[{\"name\":\"portId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"channelId\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIIBCChannelUpgradableModule.UpgradeProposal\",\"components\":[{\"name\":\"fields\",\"type\":\"tuple\",\"internalType\":\"structUpgradeFields.Data\",\"components\":[{\"name\":\"ordering\",\"type\":\"uint8\",\"internalType\":\"enumChannel.Order\"},{\"name\":\"connection_hops\",\"type\":\"string[]\",\"internalType\":\"string[]\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"name\":\"timeout\",\"type\":\"tuple\",\"internalType\":\"structTimeout.Data\",\"components\":[{\"name\":\"height\",\"type\":\"tuple\",\"internalType\":\"structHeight.Data\",\"components\":[{\"name\":\"revision_number\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"revision_height\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"timestamp\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposeUpgrade\",\"inputs\":[{\"name\":\"portId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"channelId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"upgradeFields\",\"type\":\"tuple\",\"internalType\":\"structUpgradeFields.Data\",\"components\":[{\"name\":\"ordering\",\"type\":\"uint8\",\"internalType\":\"enumChannel.Order\"},{\"name\":\"connection_hops\",\"type\":\"string[]\",\"internalType\":\"string[]\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"name\":\"timeout\",\"type\":\"tuple\",\"internalType\":\"structTimeout.Data\",\"components\":[{\"name\":\"height\",\"type\":\"tuple\",\"internalType\":\"structHeight.Data\",\"components\":[{\"name\":\"revision_number\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"revision_height\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"timestamp\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeUpgradeProposal\",\"inputs\":[{\"name\":\"portId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"channelId\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"}]",
}

// IibcchannelupgradablemoduleABI is the input ABI used to generate the binding from.
// Deprecated: Use IibcchannelupgradablemoduleMetaData.ABI instead.
var IibcchannelupgradablemoduleABI = IibcchannelupgradablemoduleMetaData.ABI

// Iibcchannelupgradablemodule is an auto generated Go binding around an Ethereum contract.
type Iibcchannelupgradablemodule struct {
	IibcchannelupgradablemoduleCaller     // Read-only binding to the contract
	IibcchannelupgradablemoduleTransactor // Write-only binding to the contract
	IibcchannelupgradablemoduleFilterer   // Log filterer for contract events
}

// IibcchannelupgradablemoduleCaller is an auto generated read-only Go binding around an Ethereum contract.
type IibcchannelupgradablemoduleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IibcchannelupgradablemoduleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IibcchannelupgradablemoduleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IibcchannelupgradablemoduleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IibcchannelupgradablemoduleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IibcchannelupgradablemoduleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IibcchannelupgradablemoduleSession struct {
	Contract     *Iibcchannelupgradablemodule // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                // Call options to use throughout this session
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// IibcchannelupgradablemoduleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IibcchannelupgradablemoduleCallerSession struct {
	Contract *IibcchannelupgradablemoduleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                      // Call options to use throughout this session
}

// IibcchannelupgradablemoduleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IibcchannelupgradablemoduleTransactorSession struct {
	Contract     *IibcchannelupgradablemoduleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                      // Transaction auth options to use throughout this session
}

// IibcchannelupgradablemoduleRaw is an auto generated low-level Go binding around an Ethereum contract.
type IibcchannelupgradablemoduleRaw struct {
	Contract *Iibcchannelupgradablemodule // Generic contract binding to access the raw methods on
}

// IibcchannelupgradablemoduleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IibcchannelupgradablemoduleCallerRaw struct {
	Contract *IibcchannelupgradablemoduleCaller // Generic read-only contract binding to access the raw methods on
}

// IibcchannelupgradablemoduleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IibcchannelupgradablemoduleTransactorRaw struct {
	Contract *IibcchannelupgradablemoduleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIibcchannelupgradablemodule creates a new instance of Iibcchannelupgradablemodule, bound to a specific deployed contract.
func NewIibcchannelupgradablemodule(address common.Address, backend bind.ContractBackend) (*Iibcchannelupgradablemodule, error) {
	contract, err := bindIibcchannelupgradablemodule(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Iibcchannelupgradablemodule{IibcchannelupgradablemoduleCaller: IibcchannelupgradablemoduleCaller{contract: contract}, IibcchannelupgradablemoduleTransactor: IibcchannelupgradablemoduleTransactor{contract: contract}, IibcchannelupgradablemoduleFilterer: IibcchannelupgradablemoduleFilterer{contract: contract}}, nil
}

// NewIibcchannelupgradablemoduleCaller creates a new read-only instance of Iibcchannelupgradablemodule, bound to a specific deployed contract.
func NewIibcchannelupgradablemoduleCaller(address common.Address, caller bind.ContractCaller) (*IibcchannelupgradablemoduleCaller, error) {
	contract, err := bindIibcchannelupgradablemodule(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IibcchannelupgradablemoduleCaller{contract: contract}, nil
}

// NewIibcchannelupgradablemoduleTransactor creates a new write-only instance of Iibcchannelupgradablemodule, bound to a specific deployed contract.
func NewIibcchannelupgradablemoduleTransactor(address common.Address, transactor bind.ContractTransactor) (*IibcchannelupgradablemoduleTransactor, error) {
	contract, err := bindIibcchannelupgradablemodule(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IibcchannelupgradablemoduleTransactor{contract: contract}, nil
}

// NewIibcchannelupgradablemoduleFilterer creates a new log filterer instance of Iibcchannelupgradablemodule, bound to a specific deployed contract.
func NewIibcchannelupgradablemoduleFilterer(address common.Address, filterer bind.ContractFilterer) (*IibcchannelupgradablemoduleFilterer, error) {
	contract, err := bindIibcchannelupgradablemodule(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IibcchannelupgradablemoduleFilterer{contract: contract}, nil
}

// bindIibcchannelupgradablemodule binds a generic wrapper to an already deployed contract.
func bindIibcchannelupgradablemodule(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IibcchannelupgradablemoduleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Iibcchannelupgradablemodule.Contract.IibcchannelupgradablemoduleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.IibcchannelupgradablemoduleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.IibcchannelupgradablemoduleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Iibcchannelupgradablemodule.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.contract.Transact(opts, method, params...)
}

// GetUpgradeProposal is a free data retrieval call binding the contract method 0x6df5433b.
//
// Solidity: function getUpgradeProposal(string portId, string channelId) view returns(((uint8,string[],string),((uint64,uint64),uint64)))
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleCaller) GetUpgradeProposal(opts *bind.CallOpts, portId string, channelId string) (IIBCChannelUpgradableModuleUpgradeProposal, error) {
	var out []interface{}
	err := _Iibcchannelupgradablemodule.contract.Call(opts, &out, "getUpgradeProposal", portId, channelId)

	if err != nil {
		return *new(IIBCChannelUpgradableModuleUpgradeProposal), err
	}

	out0 := *abi.ConvertType(out[0], new(IIBCChannelUpgradableModuleUpgradeProposal)).(*IIBCChannelUpgradableModuleUpgradeProposal)

	return out0, err

}

// GetUpgradeProposal is a free data retrieval call binding the contract method 0x6df5433b.
//
// Solidity: function getUpgradeProposal(string portId, string channelId) view returns(((uint8,string[],string),((uint64,uint64),uint64)))
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleSession) GetUpgradeProposal(portId string, channelId string) (IIBCChannelUpgradableModuleUpgradeProposal, error) {
	return _Iibcchannelupgradablemodule.Contract.GetUpgradeProposal(&_Iibcchannelupgradablemodule.CallOpts, portId, channelId)
}

// GetUpgradeProposal is a free data retrieval call binding the contract method 0x6df5433b.
//
// Solidity: function getUpgradeProposal(string portId, string channelId) view returns(((uint8,string[],string),((uint64,uint64),uint64)))
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleCallerSession) GetUpgradeProposal(portId string, channelId string) (IIBCChannelUpgradableModuleUpgradeProposal, error) {
	return _Iibcchannelupgradablemodule.Contract.GetUpgradeProposal(&_Iibcchannelupgradablemodule.CallOpts, portId, channelId)
}

// AllowTransitionToFlushComplete is a paid mutator transaction binding the contract method 0x8680841e.
//
// Solidity: function allowTransitionToFlushComplete(string portId, string channelId, uint64 upgradeSequence) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleTransactor) AllowTransitionToFlushComplete(opts *bind.TransactOpts, portId string, channelId string, upgradeSequence uint64) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.contract.Transact(opts, "allowTransitionToFlushComplete", portId, channelId, upgradeSequence)
}

// AllowTransitionToFlushComplete is a paid mutator transaction binding the contract method 0x8680841e.
//
// Solidity: function allowTransitionToFlushComplete(string portId, string channelId, uint64 upgradeSequence) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleSession) AllowTransitionToFlushComplete(portId string, channelId string, upgradeSequence uint64) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.AllowTransitionToFlushComplete(&_Iibcchannelupgradablemodule.TransactOpts, portId, channelId, upgradeSequence)
}

// AllowTransitionToFlushComplete is a paid mutator transaction binding the contract method 0x8680841e.
//
// Solidity: function allowTransitionToFlushComplete(string portId, string channelId, uint64 upgradeSequence) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleTransactorSession) AllowTransitionToFlushComplete(portId string, channelId string, upgradeSequence uint64) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.AllowTransitionToFlushComplete(&_Iibcchannelupgradablemodule.TransactOpts, portId, channelId, upgradeSequence)
}

// ProposeUpgrade is a paid mutator transaction binding the contract method 0xd6438586.
//
// Solidity: function proposeUpgrade(string portId, string channelId, (uint8,string[],string) upgradeFields, ((uint64,uint64),uint64) timeout) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleTransactor) ProposeUpgrade(opts *bind.TransactOpts, portId string, channelId string, upgradeFields UpgradeFieldsData, timeout TimeoutData) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.contract.Transact(opts, "proposeUpgrade", portId, channelId, upgradeFields, timeout)
}

// ProposeUpgrade is a paid mutator transaction binding the contract method 0xd6438586.
//
// Solidity: function proposeUpgrade(string portId, string channelId, (uint8,string[],string) upgradeFields, ((uint64,uint64),uint64) timeout) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleSession) ProposeUpgrade(portId string, channelId string, upgradeFields UpgradeFieldsData, timeout TimeoutData) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.ProposeUpgrade(&_Iibcchannelupgradablemodule.TransactOpts, portId, channelId, upgradeFields, timeout)
}

// ProposeUpgrade is a paid mutator transaction binding the contract method 0xd6438586.
//
// Solidity: function proposeUpgrade(string portId, string channelId, (uint8,string[],string) upgradeFields, ((uint64,uint64),uint64) timeout) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleTransactorSession) ProposeUpgrade(portId string, channelId string, upgradeFields UpgradeFieldsData, timeout TimeoutData) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.ProposeUpgrade(&_Iibcchannelupgradablemodule.TransactOpts, portId, channelId, upgradeFields, timeout)
}

// RemoveUpgradeProposal is a paid mutator transaction binding the contract method 0x0cc5a7af.
//
// Solidity: function removeUpgradeProposal(string portId, string channelId) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleTransactor) RemoveUpgradeProposal(opts *bind.TransactOpts, portId string, channelId string) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.contract.Transact(opts, "removeUpgradeProposal", portId, channelId)
}

// RemoveUpgradeProposal is a paid mutator transaction binding the contract method 0x0cc5a7af.
//
// Solidity: function removeUpgradeProposal(string portId, string channelId) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleSession) RemoveUpgradeProposal(portId string, channelId string) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.RemoveUpgradeProposal(&_Iibcchannelupgradablemodule.TransactOpts, portId, channelId)
}

// RemoveUpgradeProposal is a paid mutator transaction binding the contract method 0x0cc5a7af.
//
// Solidity: function removeUpgradeProposal(string portId, string channelId) returns()
func (_Iibcchannelupgradablemodule *IibcchannelupgradablemoduleTransactorSession) RemoveUpgradeProposal(portId string, channelId string) (*types.Transaction, error) {
	return _Iibcchannelupgradablemodule.Contract.RemoveUpgradeProposal(&_Iibcchannelupgradablemodule.TransactOpts, portId, channelId)
}
