require("@nomicfoundation/hardhat-toolbox");
require("@nomicfoundation/hardhat-ethers");
require("hardhat-deploy");
require("hardhat-deploy-ethers");
require("@openzeppelin/hardhat-upgrades");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.27",
    settings: {
      evmVersion: "cancun",
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  namedAccounts: {
    deployer: {
      default: 0
    },
    pauseUser: {
      default: 1
    },
    upgradeUser: {
      default: 2
    },
    platformWalletAddress: {
      default: 3
    },
    logicOperator: {
      default: 4
    },
    normalUser: {
      default: 5
    }
  },
  gasReporter: {
    enabled: false,  // 将 true 改为 false
    currency: 'USD',
    gasPrice: 21
  }
};
