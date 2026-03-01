
const { ethers, upgrades } = require("hardhat");

async function main() {
    const [deployer] = await ethers.getSigners();

    console.log("Deploying contracts with the account:", deployer.address);

    const NFTMarketPlace = await ethers.getContractFactory("NFTMarketPlace");

    const defaultAdmin = deployer.address;
    const pauser = deployer.address;
    const upgrader = deployer.address;
    const logicOperator = deployer.address;
    const platformWallet = deployer.address;
    const platformFee = 100;

    console.log("Deploying NFTMarketPlace proxy...");

    const marketplace = await upgrades.deployProxy(NFTMarketPlace, [
        defaultAdmin,
        pauser,
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