// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ibcchannelupgradablemodule

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

// IbcchannelupgradablemoduleMetaData contains all meta data concerning the Ibcchannelupgradablemodule contract.
var IbcchannelupgradablemoduleMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"allowTransitionToFlushComplete\",\"inputs\":[{\"name\":\"portId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"channelId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"upgradeSequence\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getUpgradeProposal\",\"inputs\":[{\"name\":\"portId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"channelId\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIIBCChannelUpgradableModule.UpgradeProposal\",\"components\":[{\"name\":\"fields\",\"type\":\"tuple\",\"internalType\":\"structUpgradeFields.Data\",\"components\":[{\"name\":\"ordering\",\"type\":\"uint8\",\"internalType\":\"enumChannel.Order\"},{\"name\":\"connection_hops\",\"type\":\"string[]\",\"internalType\":\"string[]\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"name\":\"timeout\",\"type\":\"tuple\",\"internalType\":\"structTimeout.Data\",\"components\":[{\"name\":\"height\",\"type\":\"tuple\",\"internalType\":\"structHeight.Data\",\"components\":[{\"name\":\"revision_number\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"revision_height\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"timestamp\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposeUpgrade\",\"inputs\":[{\"name\":\"portId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"channelId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"upgradeFields\",\"type\":\"tuple\",\"internalType\":\"structUpgradeFields.Data\",\"components\":[{\"name\":\"ordering\",\"type\":\"uint8\",\"internalType\":\"enumChannel.Order\"},{\"name\":\"connection_hops\",\"type\":\"string[]\",\"internalType\":\"string[]\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"name\":\"timeout\",\"type\":\"tuple\",\"internalType\":\"structTimeout.Data\",\"components\":[{\"name\":\"height\",\"type\":\"tuple\",\"internalType\":\"structHeight.Data\",\"components\":[{\"name\":\"revision_number\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"revision_height\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"name\":\"timestamp\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeUpgradeProposal\",\"inputs\":[{\"name\":\"portId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"channelId\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"}]",
}

// IbcchannelupgradablemoduleABI is the input ABI used to generate the binding from.
// Deprecated: Use IbcchannelupgradablemoduleMetaData.ABI instead.
var IbcchannelupgradablemoduleABI = IbcchannelupgradablemoduleMetaData.ABI

// Ibcchannelupgradablemodule is an auto generated Go binding around an Ethereum contract.
type Ibcchannelupgradablemodule struct {
	IbcchannelupgradablemoduleCaller     // Read-only binding to the contract
	IbcchannelupgradablemoduleTransactor // Write-only binding to the contract
	IbcchannelupgradablemoduleFilterer   // Log filterer for contract events
}

// IbcchannelupgradablemoduleCaller is an auto generated read-only Go binding around an Ethereum contract.
type IbcchannelupgradablemoduleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IbcchannelupgradablemoduleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IbcchannelupgradablemoduleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IbcchannelupgradablemoduleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IbcchannelupgradablemoduleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IbcchannelupgradablemoduleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IbcchannelupgradablemoduleSession struct {
	Contract     *Ibcchannelupgradablemodule // Generic contract binding to set the session for
	CallOpts     bind.CallOpts               // Call options to use throughout this session
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// IbcchannelupgradablemoduleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IbcchannelupgradablemoduleCallerSession struct {
	Contract *IbcchannelupgradablemoduleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                     // Call options to use throughout this session
}

// IbcchannelupgradablemoduleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IbcchannelupgradablemoduleTransactorSession struct {
	Contract     *IbcchannelupgradablemoduleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                     // Transaction auth options to use throughout this session
}

// IbcchannelupgradablemoduleRaw is an auto generated low-level Go binding around an Ethereum contract.
type IbcchannelupgradablemoduleRaw struct {
	Contract *Ibcchannelupgradablemodule // Generic contract binding to access the raw methods on
}

// IbcchannelupgradablemoduleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IbcchannelupgradablemoduleCallerRaw struct {
	Contract *IbcchannelupgradablemoduleCaller // Generic read-only contract binding to access the raw methods on
}

// IbcchannelupgradablemoduleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IbcchannelupgradablemoduleTransactorRaw struct {
	Contract *IbcchannelupgradablemoduleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIbcchannelupgradablemodule creates a new instance of Ibcchannelupgradablemodule, bound to a specific deployed contract.
func NewIbcchannelupgradablemodule(address common.Address, backend bind.ContractBackend) (*Ibcchannelupgradablemodule, error) {
	contract, err := bindIbcchannelupgradablemodule(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ibcchannelupgradablemodule{IbcchannelupgradablemoduleCaller: IbcchannelupgradablemoduleCaller{contract: contract}, IbcchannelupgradablemoduleTransactor: IbcchannelupgradablemoduleTransactor{contract: contract}, IbcchannelupgradablemoduleFilterer: IbcchannelupgradablemoduleFilterer{contract: contract}}, nil
}

// NewIbcchannelupgradablemoduleCaller creates a new read-only instance of Ibcchannelupgradablemodule, bound to a specific deployed contract.
func NewIbcchannelupgradablemoduleCaller(address common.Address, caller bind.ContractCaller) (*IbcchannelupgradablemoduleCaller, error) {
	contract, err := bindIbcchannelupgradablemodule(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IbcchannelupgradablemoduleCaller{contract: contract}, nil
}

// NewIbcchannelupgradablemoduleTransactor creates a new write-only instance of Ibcchannelupgradablemodule, bound to a specific deployed contract.
func NewIbcchannelupgradablemoduleTransactor(address common.Address, transactor bind.ContractTransactor) (*IbcchannelupgradablemoduleTransactor, error) {
	contract, err := bindIbcchannelupgradablemodule(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IbcchannelupgradablemoduleTransactor{contract: contract}, nil
}

// NewIbcchannelupgradablemoduleFilterer creates a new log filterer instance of Ibcchannelupgradablemodule, bound to a specific deployed contract.
func NewIbcchannelupgradablemoduleFilterer(address common.Address, filterer bind.ContractFilterer) (*IbcchannelupgradablemoduleFilterer, error) {
	contract, err := bindIbcchannelupgradablemodule(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IbcchannelupgradablemoduleFilterer{contract: contract}, nil
}

// bindIbcchannelupgradablemodule binds a generic wrapper to an already deployed contract.
func bindIbcchannelupgradablemodule(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IbcchannelupgradablemoduleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ibcchannelupgradablemodule.Contract.IbcchannelupgradablemoduleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.IbcchannelupgradablemoduleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.IbcchannelupgradablemoduleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ibcchannelupgradablemodule.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.contract.Transact(opts, method, params...)
}

// GetUpgradeProposal is a free data retrieval call binding the contract method 0x6df5433b.
//
// Solidity: function getUpgradeProposal(string portId, string channelId) view returns(((uint8,string[],string),((uint64,uint64),uint64)))
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleCaller) GetUpgradeProposal(opts *bind.CallOpts, portId string, channelId string) (IIBCChannelUpgradableModuleUpgradeProposal, error) {
	var out []interface{}
	err := _Ibcchannelupgradablemodule.contract.Call(opts, &out, "getUpgradeProposal", portId, channelId)

	if err != nil {
		return *new(IIBCChannelUpgradableModuleUpgradeProposal), err
	}

	out0 := *abi.ConvertType(out[0], new(IIBCChannelUpgradableModuleUpgradeProposal)).(*IIBCChannelUpgradableModuleUpgradeProposal)

	return out0, err

}

// GetUpgradeProposal is a free data retrieval call binding the contract method 0x6df5433b.
//
// Solidity: function getUpgradeProposal(string portId, string channelId) view returns(((uint8,string[],string),((uint64,uint64),uint64)))
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleSession) GetUpgradeProposal(portId string, channelId string) (IIBCChannelUpgradableModuleUpgradeProposal, error) {
	return _Ibcchannelupgradablemodule.Contract.GetUpgradeProposal(&_Ibcchannelupgradablemodule.CallOpts, portId, channelId)
}

// GetUpgradeProposal is a free data retrieval call binding the contract method 0x6df5433b.
//
// Solidity: function getUpgradeProposal(string portId, string channelId) view returns(((uint8,string[],string),((uint64,uint64),uint64)))
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleCallerSession) GetUpgradeProposal(portId string, channelId string) (IIBCChannelUpgradableModuleUpgradeProposal, error) {
	return _Ibcchannelupgradablemodule.Contract.GetUpgradeProposal(&_Ibcchannelupgradablemodule.CallOpts, portId, channelId)
}

// AllowTransitionToFlushComplete is a paid mutator transaction binding the contract method 0x8680841e.
//
// Solidity: function allowTransitionToFlushComplete(string portId, string channelId, uint64 upgradeSequence) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleTransactor) AllowTransitionToFlushComplete(opts *bind.TransactOpts, portId string, channelId string, upgradeSequence uint64) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.contract.Transact(opts, "allowTransitionToFlushComplete", portId, channelId, upgradeSequence)
}

// AllowTransitionToFlushComplete is a paid mutator transaction binding the contract method 0x8680841e.
//
// Solidity: function allowTransitionToFlushComplete(string portId, string channelId, uint64 upgradeSequence) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleSession) AllowTransitionToFlushComplete(portId string, channelId string, upgradeSequence uint64) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.AllowTransitionToFlushComplete(&_Ibcchannelupgradablemodule.TransactOpts, portId, channelId, upgradeSequence)
}

// AllowTransitionToFlushComplete is a paid mutator transaction binding the contract method 0x8680841e.
//
// Solidity: function allowTransitionToFlushComplete(string portId, string channelId, uint64 upgradeSequence) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleTransactorSession) AllowTransitionToFlushComplete(portId string, channelId string, upgradeSequence uint64) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.AllowTransitionToFlushComplete(&_Ibcchannelupgradablemodule.TransactOpts, portId, channelId, upgradeSequence)
}

// ProposeUpgrade is a paid mutator transaction binding the contract method 0xd6438586.
//
// Solidity: function proposeUpgrade(string portId, string channelId, (uint8,string[],string) upgradeFields, ((uint64,uint64),uint64) timeout) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleTransactor) ProposeUpgrade(opts *bind.TransactOpts, portId string, channelId string, upgradeFields UpgradeFieldsData, timeout TimeoutData) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.contract.Transact(opts, "proposeUpgrade", portId, channelId, upgradeFields, timeout)
}

// ProposeUpgrade is a paid mutator transaction binding the contract method 0xd6438586.
//
// Solidity: function proposeUpgrade(string portId, string channelId, (uint8,string[],string) upgradeFields, ((uint64,uint64),uint64) timeout) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleSession) ProposeUpgrade(portId string, channelId string, upgradeFields UpgradeFieldsData, timeout TimeoutData) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.ProposeUpgrade(&_Ibcchannelupgradablemodule.TransactOpts, portId, channelId, upgradeFields, timeout)
}

// ProposeUpgrade is a paid mutator transaction binding the contract method 0xd6438586.
//
// Solidity: function proposeUpgrade(string portId, string channelId, (uint8,string[],string) upgradeFields, ((uint64,uint64),uint64) timeout) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleTransactorSession) ProposeUpgrade(portId string, channelId string, upgradeFields UpgradeFieldsData, timeout TimeoutData) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.ProposeUpgrade(&_Ibcchannelupgradablemodule.TransactOpts, portId, channelId, upgradeFields, timeout)
}

// RemoveUpgradeProposal is a paid mutator transaction binding the contract method 0x0cc5a7af.
//
// Solidity: function removeUpgradeProposal(string portId, string channelId) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleTransactor) RemoveUpgradeProposal(opts *bind.TransactOpts, portId string, channelId string) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.contract.Transact(opts, "removeUpgradeProposal", portId, channelId)
}

// RemoveUpgradeProposal is a paid mutator transaction binding the contract method 0x0cc5a7af.
//
// Solidity: function removeUpgradeProposal(string portId, string channelId) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleSession) RemoveUpgradeProposal(portId string, channelId string) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.RemoveUpgradeProposal(&_Ibcchannelupgradablemodule.TransactOpts, portId, channelId)
}

// RemoveUpgradeProposal is a paid mutator transaction binding the contract method 0x0cc5a7af.
//
// Solidity: function removeUpgradeProposal(string portId, string channelId) returns()
func (_Ibcchannelupgradablemodule *IbcchannelupgradablemoduleTransactorSession) RemoveUpgradeProposal(portId string, channelId string) (*types.Transaction, error) {
	return _Ibcchannelupgradablemodule.Contract.RemoveUpgradeProposal(&_Ibcchannelupgradablemodule.TransactOpts, portId, channelId)
}
