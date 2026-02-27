const { upgrades,ethers } = require("hardhat");

module.exports = async ({getNamedAccounts, deployments}) => {
    const {deployer, pauseUser, upgradeUser,logicOperator,platformWalletAddress} = await getNamedAccounts();
    const {deploy, log} = deployments;

    log("Deploying NFT Marketplace...");

    const nftMarketPlace = await ethers.getContractFactory("NFTMarketPlace");

    const nftMarketPlaceProxy = await upgrades.deployProxy(
        nftMarketPlace,
        [deployer, pauseUser, upgradeUser,logicOperator, platformWalletAddress,100]
    );

    await nftMarketPlaceProxy.waitForDeployment();

    await deployments.save("NFTMarketPlace", {
        abi: (await deployments.getExtendedArtifact("NFTMarketPlace")).abi,
        address: nftMarketPlaceProxy.target,
    });

    log("NFT Marketplace deployed successfully with address "+nftMarketPlaceProxy.target);
}

module.exports.tags = ["all", "marketplace"];