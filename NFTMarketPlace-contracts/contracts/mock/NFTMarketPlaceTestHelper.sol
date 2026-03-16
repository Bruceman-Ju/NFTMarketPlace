// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import {NFTMarketPlace} from "../NFTMarketPlace.sol";

/**
 * @title NFTMarketPlaceTestHelper
 * @dev This contract is used to test the upgradeability of the NFTMarketPlace contract.
 */
contract NFTMarketPlaceTestHelper is NFTMarketPlace {

    function testUpgrade() public pure returns (bool) {
        return true;
    }

    function testSetInvalidSellerAddress(bytes32 _listId,address _unReceiveAddress) public onlyRole(DEFAULT_ADMIN_ROLE){
        ListedNFT storage nft_ = listedNFTs[_listId];
        nft_.seller = _unReceiveAddress;
    }
}
