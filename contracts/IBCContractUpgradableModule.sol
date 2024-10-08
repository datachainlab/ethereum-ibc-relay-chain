// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

import {IIBCContractUpgradableModule, IIBCContractUpgradableModuleErrors} from "./IIBCContractUpgradableModule.sol";

import {IBCChannelUpgradableModuleBase} from "@hyperledger-labs/yui-ibc-solidity/contracts/apps/commons/IBCChannelUpgradableModule.sol";
import {Channel, UpgradeFields} from "@hyperledger-labs/yui-ibc-solidity/contracts/proto/Channel.sol";
import {IIBCHandler} from "@hyperledger-labs/yui-ibc-solidity/contracts/core/25-handler/IIBCHandler.sol";

abstract contract IBCContractUpgradableModuleBase is
    IBCChannelUpgradableModuleBase,
    IIBCContractUpgradableModule,
    IIBCContractUpgradableModuleErrors
{
    // ------------------- Storage ------------------- //

    // NOTE: A module should set an initial appVersion struct in contract constructor or initializer
    mapping(string appVersion => AppInfo) internal appInfos;

    // ------------------- Modifiers ------------------- //

    modifier onlyContractUpgrader() {
        address msgSender = _msgSender();
        if (!_isContractUpgrader(msgSender)) {
            revert IBCContractUpgradableModuleUnauthorizedUpgrader(msgSender);
        }
        _;
    }

    // ------------------- Functions ------------------- //

    /**
     * @dev See {IIBCContractUpgradableModule-getAppinfoproposal}
     */
    function getAppInfoProposal(string calldata version)
        external
        view
        override(IIBCContractUpgradableModule)
        returns (AppInfo memory)
    {
        return appInfos[version];
    }

    /**
     * @dev See {IIBCContractUpgradableModule-proposeAppVersion}
     */
    function proposeAppVersion(string calldata version, address implementation, bytes calldata initialCalldata)
        external
        override(IIBCContractUpgradableModule)
        onlyContractUpgrader
    {
        if (implementation == address(0)) {
            revert IBCContractUpgradableModuleAppInfoProposedWithZeroImpl();
        }

        AppInfo storage appInfo = appInfos[version];
        if (appInfo.implementation != address(0)) {
            revert IBCContractUpgradableModuleAppInfoIsAlreadySet();
        }

        appInfos[version] = AppInfo({
            implementation: implementation,
            initialCalldata: initialCalldata,
            consumed: false
        });
    }

    /**
     * @dev See {IERC165-supportsInterface}.
     */
    function supportsInterface(bytes4 interfaceId)
        public
        view
        virtual
        override(IBCChannelUpgradableModuleBase)
        returns (bool)
    {
        return super.supportsInterface(interfaceId) ||
            interfaceId == type(IIBCContractUpgradableModule).interfaceId;
    }

    /**
     * @dev See {IIBCModuleUpgrade-onChanUpgradeInit}
     */
    function onChanUpgradeInit(
        string calldata portId,
        string calldata channelId,
        uint64 upgradeSequence,
        UpgradeFields.Data calldata proposedUpgradeFields
    )
        public
        view
        virtual
        override(IBCChannelUpgradableModuleBase)
        onlyIBC
        returns (string memory version)
    {
        version = super.onChanUpgradeInit(portId, channelId, upgradeSequence, proposedUpgradeFields);

        (Channel.Data memory channel, bool found) = IIBCHandler(ibcAddress()).getChannel(portId, channelId);
        if (!found) {
            revert IBCContractUpgradableModuleChannelNotFound(portId, channelId);
        }

        _prepareContractUpgrade(channel.version, version);
    }

    /**
     * @dev See {IIBCModuleUpgrade-onChanUpgradeTry}
     */
    function onChanUpgradeTry(
        string calldata portId,
        string calldata channelId,
        uint64 upgradeSequence,
        UpgradeFields.Data calldata proposedUpgradeFields
    )
        public
        view
        virtual
        override(IBCChannelUpgradableModuleBase)
        onlyIBC
        returns (string memory version)
    {
        version = super.onChanUpgradeTry(portId, channelId, upgradeSequence, proposedUpgradeFields);

        (Channel.Data memory channel, bool found) = IIBCHandler(ibcAddress()).getChannel(portId, channelId);
        if (!found) {
            revert IBCContractUpgradableModuleChannelNotFound(portId, channelId);
        }

        _prepareContractUpgrade(channel.version, version);
    }

    /**
     * @dev See {IIBCModuleUpgrade-onChanUpgradeOpen}
     */
    function onChanUpgradeOpen(
        string calldata portId,
        string calldata channelId,
        uint64 upgradeSequence
    )
        public
        virtual
        override(IBCChannelUpgradableModuleBase)
        onlyIBC
    {
        super.onChanUpgradeOpen(portId, channelId, upgradeSequence);

        (Channel.Data memory channel, bool found) = IIBCHandler(ibcAddress()).getChannel(portId, channelId);
        if (!found) {
            revert IBCContractUpgradableModuleChannelNotFound(portId, channelId);
        }

        _upgradeContract(channel.version);
    }

    // ------------------- Internal Functions ------------------- //

    function _isContractUpgrader(address msgSender) internal view virtual returns (bool);

    function _doUpgradeContract(address implementation, bytes memory initialCalldata) internal virtual;

    // ------------------- Private Functions ------------------- //

    function _prepareContractUpgrade(string memory version, string memory newVersion) private view {
        if (!_compareString(version, newVersion)) {
            AppInfo storage appInfo = appInfos[newVersion];
            if (appInfo.implementation == address(0)) {
                revert IBCContractUpgradableModuleAppInfoNotProposedYet();
            }
            if (appInfo.consumed) {
                revert IBCContractUpgradableModuleAlreadyConsumedAppInfo();
            }
        }
    }

    function _compareString(string memory a, string memory b) private pure returns (bool) {
        if (bytes(a).length != bytes(b).length) {
            return false;
        }
        return keccak256(abi.encodePacked(a)) == keccak256(abi.encodePacked(b));
    }

    function _upgradeContract(string memory newVersion) private {
        AppInfo storage appInfo = appInfos[newVersion];

        if (appInfo.implementation != address(0) && !appInfo.consumed) {
            appInfo.consumed = true;
            _doUpgradeContract(appInfo.implementation, appInfo.initialCalldata);
            delete appInfo.initialCalldata;
        }
    }
}
