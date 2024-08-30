// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package iibccontractupgradablemodule

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

// IIBCContractUpgradableModuleAppInfo is an auto generated low-level Go binding around an user-defined struct.
type IIBCContractUpgradableModuleAppInfo struct {
	Implementation  common.Address
	InitialCalldata []byte
	Consumed        bool
}

// IibccontractupgradablemoduleMetaData contains all meta data concerning the Iibccontractupgradablemodule contract.
var IibccontractupgradablemoduleMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"getAppInfoProposal\",\"inputs\":[{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIIBCContractUpgradableModule.AppInfo\",\"components\":[{\"name\":\"implementation\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"initialCalldata\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"consumed\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposeAppVersion\",\"inputs\":[{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"appInfo\",\"type\":\"tuple\",\"internalType\":\"structIIBCContractUpgradableModule.AppInfo\",\"components\":[{\"name\":\"implementation\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"initialCalldata\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"consumed\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"}]",
}

// IibccontractupgradablemoduleABI is the input ABI used to generate the binding from.
// Deprecated: Use IibccontractupgradablemoduleMetaData.ABI instead.
var IibccontractupgradablemoduleABI = IibccontractupgradablemoduleMetaData.ABI

// Iibccontractupgradablemodule is an auto generated Go binding around an Ethereum contract.
type Iibccontractupgradablemodule struct {
	IibccontractupgradablemoduleCaller     // Read-only binding to the contract
	IibccontractupgradablemoduleTransactor // Write-only binding to the contract
	IibccontractupgradablemoduleFilterer   // Log filterer for contract events
}

// IibccontractupgradablemoduleCaller is an auto generated read-only Go binding around an Ethereum contract.
type IibccontractupgradablemoduleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IibccontractupgradablemoduleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IibccontractupgradablemoduleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IibccontractupgradablemoduleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IibccontractupgradablemoduleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IibccontractupgradablemoduleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IibccontractupgradablemoduleSession struct {
	Contract     *Iibccontractupgradablemodule // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                 // Call options to use throughout this session
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// IibccontractupgradablemoduleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IibccontractupgradablemoduleCallerSession struct {
	Contract *IibccontractupgradablemoduleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                       // Call options to use throughout this session
}

// IibccontractupgradablemoduleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IibccontractupgradablemoduleTransactorSession struct {
	Contract     *IibccontractupgradablemoduleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                       // Transaction auth options to use throughout this session
}

// IibccontractupgradablemoduleRaw is an auto generated low-level Go binding around an Ethereum contract.
type IibccontractupgradablemoduleRaw struct {
	Contract *Iibccontractupgradablemodule // Generic contract binding to access the raw methods on
}

// IibccontractupgradablemoduleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IibccontractupgradablemoduleCallerRaw struct {
	Contract *IibccontractupgradablemoduleCaller // Generic read-only contract binding to access the raw methods on
}

// IibccontractupgradablemoduleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IibccontractupgradablemoduleTransactorRaw struct {
	Contract *IibccontractupgradablemoduleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIibccontractupgradablemodule creates a new instance of Iibccontractupgradablemodule, bound to a specific deployed contract.
func NewIibccontractupgradablemodule(address common.Address, backend bind.ContractBackend) (*Iibccontractupgradablemodule, error) {
	contract, err := bindIibccontractupgradablemodule(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Iibccontractupgradablemodule{IibccontractupgradablemoduleCaller: IibccontractupgradablemoduleCaller{contract: contract}, IibccontractupgradablemoduleTransactor: IibccontractupgradablemoduleTransactor{contract: contract}, IibccontractupgradablemoduleFilterer: IibccontractupgradablemoduleFilterer{contract: contract}}, nil
}

// NewIibccontractupgradablemoduleCaller creates a new read-only instance of Iibccontractupgradablemodule, bound to a specific deployed contract.
func NewIibccontractupgradablemoduleCaller(address common.Address, caller bind.ContractCaller) (*IibccontractupgradablemoduleCaller, error) {
	contract, err := bindIibccontractupgradablemodule(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IibccontractupgradablemoduleCaller{contract: contract}, nil
}

// NewIibccontractupgradablemoduleTransactor creates a new write-only instance of Iibccontractupgradablemodule, bound to a specific deployed contract.
func NewIibccontractupgradablemoduleTransactor(address common.Address, transactor bind.ContractTransactor) (*IibccontractupgradablemoduleTransactor, error) {
	contract, err := bindIibccontractupgradablemodule(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IibccontractupgradablemoduleTransactor{contract: contract}, nil
}

// NewIibccontractupgradablemoduleFilterer creates a new log filterer instance of Iibccontractupgradablemodule, bound to a specific deployed contract.
func NewIibccontractupgradablemoduleFilterer(address common.Address, filterer bind.ContractFilterer) (*IibccontractupgradablemoduleFilterer, error) {
	contract, err := bindIibccontractupgradablemodule(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IibccontractupgradablemoduleFilterer{contract: contract}, nil
}

// bindIibccontractupgradablemodule binds a generic wrapper to an already deployed contract.
func bindIibccontractupgradablemodule(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IibccontractupgradablemoduleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Iibccontractupgradablemodule.Contract.IibccontractupgradablemoduleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Iibccontractupgradablemodule.Contract.IibccontractupgradablemoduleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Iibccontractupgradablemodule.Contract.IibccontractupgradablemoduleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Iibccontractupgradablemodule.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Iibccontractupgradablemodule.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Iibccontractupgradablemodule.Contract.contract.Transact(opts, method, params...)
}

// GetAppInfoProposal is a free data retrieval call binding the contract method 0xff6bb285.
//
// Solidity: function getAppInfoProposal(string version) view returns((address,bytes,bool))
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleCaller) GetAppInfoProposal(opts *bind.CallOpts, version string) (IIBCContractUpgradableModuleAppInfo, error) {
	var out []interface{}
	err := _Iibccontractupgradablemodule.contract.Call(opts, &out, "getAppInfoProposal", version)

	if err != nil {
		return *new(IIBCContractUpgradableModuleAppInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(IIBCContractUpgradableModuleAppInfo)).(*IIBCContractUpgradableModuleAppInfo)

	return out0, err

}

// GetAppInfoProposal is a free data retrieval call binding the contract method 0xff6bb285.
//
// Solidity: function getAppInfoProposal(string version) view returns((address,bytes,bool))
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleSession) GetAppInfoProposal(version string) (IIBCContractUpgradableModuleAppInfo, error) {
	return _Iibccontractupgradablemodule.Contract.GetAppInfoProposal(&_Iibccontractupgradablemodule.CallOpts, version)
}

// GetAppInfoProposal is a free data retrieval call binding the contract method 0xff6bb285.
//
// Solidity: function getAppInfoProposal(string version) view returns((address,bytes,bool))
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleCallerSession) GetAppInfoProposal(version string) (IIBCContractUpgradableModuleAppInfo, error) {
	return _Iibccontractupgradablemodule.Contract.GetAppInfoProposal(&_Iibccontractupgradablemodule.CallOpts, version)
}

// ProposeAppVersion is a paid mutator transaction binding the contract method 0x347501f8.
//
// Solidity: function proposeAppVersion(string version, (address,bytes,bool) appInfo) returns()
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleTransactor) ProposeAppVersion(opts *bind.TransactOpts, version string, appInfo IIBCContractUpgradableModuleAppInfo) (*types.Transaction, error) {
	return _Iibccontractupgradablemodule.contract.Transact(opts, "proposeAppVersion", version, appInfo)
}

// ProposeAppVersion is a paid mutator transaction binding the contract method 0x347501f8.
//
// Solidity: function proposeAppVersion(string version, (address,bytes,bool) appInfo) returns()
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleSession) ProposeAppVersion(version string, appInfo IIBCContractUpgradableModuleAppInfo) (*types.Transaction, error) {
	return _Iibccontractupgradablemodule.Contract.ProposeAppVersion(&_Iibccontractupgradablemodule.TransactOpts, version, appInfo)
}

// ProposeAppVersion is a paid mutator transaction binding the contract method 0x347501f8.
//
// Solidity: function proposeAppVersion(string version, (address,bytes,bool) appInfo) returns()
func (_Iibccontractupgradablemodule *IibccontractupgradablemoduleTransactorSession) ProposeAppVersion(version string, appInfo IIBCContractUpgradableModuleAppInfo) (*types.Transaction, error) {
	return _Iibccontractupgradablemodule.Contract.ProposeAppVersion(&_Iibccontractupgradablemodule.TransactOpts, version, appInfo)
}
