// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.20;

interface IIBCContractUpgradableModuleErrors {
    // ------------------- Errors ------------------- //

    error IBCContractUpgradableModuleUnauthorizedUpgrader(address msgSender);
    error IBCContractUpgradableModuleAppInfoProposedWithZeroImpl();
    error IBCContractUpgradableModuleAppInfoNotProposedYet();
    error IBCContractUpgradableModuleAppInfoIsAlreadySet();
    error IBCContractUpgradableModuleChannelNotFound(string portId, string channelId);
    error IBCContractUpgradableModuleAlreadyConsumedAppInfo();
}

interface IIBCContractUpgradableModule {
    // ------------------- Data Structures ------------------- //

    /**
     * @dev Proposed AppInfo data
     * @param implemantation the new implementation address
     * @param initialCalldata the first function call's calldata
     * @param consumed the flag that specifies if this implementation is already deployed
     */
    struct AppInfo {
        address implementation;
        bytes initialCalldata;
        bool consumed;
    }

    // ------------------- Functions ------------------- //

    /**
     * @dev Returns the proposed AppInfo for the given version
     */
    function getAppInfoProposal(string calldata version) external view returns (AppInfo memory);

    /**
     * @dev Propose an Appinfo for the given version
     * @notice This function is only callable by an authorized upgrader.
     *         To upgrade the IBC module contract along with a channel upgrade, the upgrader must
     *         call this function before calling `channelUpgradeInit` or `channelUpgradeTry` of the IBC handler.
     */
    function proposeAppVersion(string calldata version, AppInfo calldata appInfo) external;

}

