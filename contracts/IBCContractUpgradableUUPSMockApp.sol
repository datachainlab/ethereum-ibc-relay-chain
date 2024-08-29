// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {ERC1967Utils} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Utils.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

import {IBCContractUpgradableModuleBase} from "./IBCContractUpgradableModule.sol";

import {IBCMockApp} from "@hyperledger-labs/yui-ibc-solidity/contracts/apps/mock/IBCMockApp.sol";
import {IBCAppBase} from "@hyperledger-labs/yui-ibc-solidity/contracts/apps/commons/IBCAppBase.sol";
import {IBCChannelUpgradableModuleBase} from "@hyperledger-labs/yui-ibc-solidity/contracts/apps/commons/IBCChannelUpgradableModule.sol";
import {IIBCHandler} from "@hyperledger-labs/yui-ibc-solidity/contracts/core/25-handler/IIBCHandler.sol";

contract IBCContractUpgradableUUPSMockApp is
    UUPSUpgradeable,
    IBCMockApp,
    IBCContractUpgradableModuleBase
{
    address immutable _self;     // implementation's address, which is set by the constructor
    address immutable _deployer; // contract deployer, which is set by the constructor
    constructor(IIBCHandler ibcHandler_) IBCMockApp(ibcHandler_) {
        _self = address(this);
        _deployer = msg.sender;
    }

    // ------------------- Functions ------------------- //

    function self() external view returns(address) {
        return _self;
    }

    /**
     * @dev See {Ownable-owner}.
     */
    function owner() public view virtual override(Ownable) returns (address) {
        return _deployer;
    }

    /**
     * @dev See {IERC165-supportsInterface}.
     */
    function supportsInterface(bytes4 interfaceId)
        public
        view
        virtual
        override(IBCAppBase, IBCContractUpgradableModuleBase)
        returns (bool)
    {
        return super.supportsInterface(interfaceId) ||
            interfaceId == type(UUPSUpgradeable).interfaceId;
    }

    // ------------------- Internal Functions ------------------- //

    function __IBCContractUpgradableUUPSMockApp_init(string memory initialVersion) internal initializer {
        __UUPSUpgradeable_init();

        AppInfo storage appInfo = appInfos[initialVersion];
        appInfo.implementation = _self;
        appInfo.consumed = true;
    }

    /**
     * @dev See{IBCChannelUpgradableModuleBase-_isAuthorizedUpgrader}
     */
    function _isAuthorizedUpgrader(string calldata, string calldata, address msgSender)
        internal
        view
        virtual
        override(IBCChannelUpgradableModuleBase)
        returns (bool)
    {
        return _isContractUpgrader(msgSender);
    }

    /**
     * @dev See{IBCContractUpgradableModuleBase-_isContractUpgrader}
     */
    function _isContractUpgrader(address msgSender)
        internal
        view
        virtual
        override(IBCContractUpgradableModuleBase)
        returns (bool)
    {
        return msgSender == owner() || msgSender == address(this);
    }

    /**
     * @dev See{IBCContractUpgradableModuleBase-_upgradeContract}
     */
    function _doUpgradeContract(address implementation, bytes memory initialCalldata)
        internal
        virtual
        override(IBCContractUpgradableModuleBase)
    {
        // if there is no implementation update, nothing happens here
        if (ERC1967Utils.getImplementation() != implementation) {
            ERC1967Utils.upgradeToAndCall(implementation, initialCalldata);
        }
    }

    /**
     * @dev See {UUPSupgradeable-_authorizeupgrade}
     */
    function _authorizeUpgrade(address msgSender) internal view virtual override(UUPSUpgradeable) {
        if (!_isContractUpgrader(msgSender)) {
            revert IBCContractUpgradableModuleUnauthorizedUpgrader(msgSender);
        }
    }
}
