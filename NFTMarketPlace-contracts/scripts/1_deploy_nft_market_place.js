
const { ethers, upgrades } = require("hardhat");

async function main() {
    const [deployer,pauserAddr] = await ethers.getSigners();

    console.log("Deploying contracts with the admin address ", deployer.address ," and pauser address", pauserAddr.address);

    const NFTMarketPlace = await ethers.getContractFactory("NFTMarketPlace");

    const defaultAdmin = deployer.address;
    const pauser = pauserAddr.address;
    const unpauser = deployer.address;
    const upgrader = deployer.address;
    const logicOperator = deployer.address;
    const platformWallet = deployer.address;
    const platformFee = 100;

    console.log("Deploying NFTMarketPlace proxy...");

    const marketplace = await upgrades.deployProxy(NFTMarketPlace, [
        defaultAdmin,
        pauser,
        unpauser,
        upgrader,
        logicOperator,
        platformWallet,
        platformFee
    ], {
        initializer: "initialize",
        kind: "uups"
    });

    await marketplace.waitForDeployment();
    console.log("NFTMarketPlace deployed to:", marketplace.target);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });