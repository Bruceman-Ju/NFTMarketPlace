// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import {NFTMarketPlace} from "./NFTMarketPlace.sol";

contract NFTMarketPlaceV2 is NFTMarketPlace{

   function testUpgrade() external pure returns (bool) {
        return true;
    }
}
