module.exports = async ({getNamedAccounts, deployments}) => {

    const {deployer} = await getNamedAccounts();

    const {deploy, log} = deployments;

    log("Deploying Mock NFT...");

    await deploy("MockNFT",{
        contract: "MockNFT",
        from: deployer,
        log: true,
        args: ["MockNFT","MNT"]
    });

    log("Mock NFT deployed successfully");
}

module.exports.tags = ["all","MockNFT"];