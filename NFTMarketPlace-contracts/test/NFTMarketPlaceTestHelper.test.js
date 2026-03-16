const {expect} = require("chai");
const { ethers, upgrades } = require("hardhat");

describe("NFTMarketPlaceTestHelper tests", function () {
    let marketplace;
    let mockNFT;
    let unReceiveETH;
    let owner;
    let pauser;
    let unPauser;
    let upgrader;
    let logicOperator;
    let seller;
    let buyer;
    let platformWallet;
    let random;
    let userMap = new Map();

    const PLATFORM_FEE = 100; // 1%

    before(async () => {
        [owner, pauser,unPauser, upgrader, logicOperator, seller, buyer, platformWallet, random]
            = await ethers.getSigners();

        userMap.set(seller.address, 0);
        userMap.set(buyer.address, 0);

        // Deploy Mock ERC721
        const MockNFT = await ethers.getContractFactory("MockNFT");
        mockNFT = await MockNFT.deploy("MockNFT", "MFT");
        await mockNFT.waitForDeployment();

        // Deploy Marketplace (UUPS proxy)
        const MarketPlace = await ethers.getContractFactory("NFTMarketPlaceTestHelper");
        marketplace = await upgrades.deployProxy(MarketPlace, [
            owner.address,
            pauser.address,
            unPauser.address,
            upgrader.address,
            logicOperator.address,
            platformWallet.address,
            PLATFORM_FEE
        ], { initializer: 'initialize' });
        await marketplace.waitForDeployment();

        // Deploy UnReceiveETH to test transfer ETH
        const UnReceiveETH = await ethers.getContractFactory("UnReceiveETH");
        unReceiveETH = await UnReceiveETH.deploy();
        await unReceiveETH.waitForDeployment();
    });

    async function getListId(userAddress, nftAddress, tokenId) {

        let userNonce = userMap.get(userAddress);
        userMap.set(userAddress, userNonce+1);
        return ethers.keccak256(
            ethers.AbiCoder.defaultAbiCoder().encode(
                ["address", "uint256","address","uint256"],
                [userAddress,userNonce, nftAddress, tokenId]));
    }

    it("Should reverted if testSetInvalidSellerAddress executed with wrong role",async () => {
        // mint
        const tokenId = 1;
        await mockNFT.mint(seller.address, tokenId);
        // approve
        await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
        // list
        // 除了 revert tests 之外，每次 listNFT,就计算一次 listId，这样就让 userNonce 与合约中的基数相同了。
        await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
        const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);

        await expect(marketplace.connect(seller).testSetInvalidSellerAddress(listId,unReceiveETH))
            .to.be.revertedWithCustomError(
                marketplace,
                "AccessControlUnauthorizedAccount"
            ).withArgs(seller.address, await marketplace.DEFAULT_ADMIN_ROLE());
    });
});