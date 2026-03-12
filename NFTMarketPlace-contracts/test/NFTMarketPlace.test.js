const {expect} = require("chai");
const { ethers, upgrades } = require("hardhat");

// Helper to fast-forward time
const timeTravel = async (seconds) => {
    await ethers.provider.send("evm_increaseTime", [seconds]);
    await ethers.provider.send("evm_mine", []);
};

describe("NFTMarketPlace", function () {
    let marketplace;
    let mockNFT;
    let owner;
    let pauser;
    let upgrader;
    let logicOperator;
    let seller;
    let buyer;
    let platformWallet;
    let random;
    let userMap = new Map();

    const PLATFORM_FEE = 100; // 1%
    const LISTING_DURATION = 30 * 24 * 60 * 60; // 30 days

    before(async () => {
        [owner, pauser, upgrader, logicOperator, seller, buyer, platformWallet, random] = await ethers.getSigners();

        userMap.set(seller.address, 0);
        userMap.set(buyer.address, 0);

        // Deploy Mock ERC721
        const MockNFT = await ethers.getContractFactory("MockNFT");
        mockNFT = await MockNFT.deploy("MockNFT", "MFT");
        await mockNFT.waitForDeployment();

        // Deploy Marketplace (UUPS proxy)
        const MarketPlace = await ethers.getContractFactory("NFTMarketPlace");
        marketplace = await upgrades.deployProxy(MarketPlace, [
            owner.address,
            pauser.address,
            upgrader.address,
            logicOperator.address,
            platformWallet.address,
            PLATFORM_FEE
        ], { initializer: 'initialize' });
        await marketplace.waitForDeployment();
    });

    async function getListId(userAddress, nftAddress, tokenId) {

        let userNonce = userMap.get(userAddress);
        userMap.set(userAddress, userNonce+1);
        return ethers.keccak256(
            ethers.AbiCoder.defaultAbiCoder().encode(
                ["address", "uint256","address","uint256"],
                [userAddress,userNonce, nftAddress, tokenId]));
    }

    describe("Initialization", () => {
        it("Should set correct initial values", async () => {
            expect(await marketplace.platformWalletAddress()).to.equal(platformWallet.address);
            expect(await marketplace.platformFee()).to.equal(PLATFORM_FEE);
            expect(await marketplace.listingDuration()).to.equal(LISTING_DURATION);
        });

        it("Should revert if wallet is zero address", async () => {
            const MarketPlace = await ethers.getContractFactory("NFTMarketPlace");
            await expect(
                upgrades.deployProxy(MarketPlace, [
                    owner.address, pauser.address, upgrader.address, logicOperator.address, ethers.ZeroAddress, PLATFORM_FEE
                ], { initializer: 'initialize' })
            ).to.be.revertedWith("Invalid wallet address");
        });

        it("Should revert if fee > 1000", async () => {
            const MarketPlace = await ethers.getContractFactory("NFTMarketPlace");
            await expect(
                upgrades.deployProxy(MarketPlace, [
                    owner.address, pauser.address, upgrader.address, logicOperator.address, platformWallet.address, 1001
                ], { initializer: 'initialize' })
            ).to.be.revertedWith("Fee too high");
        });
    });

    describe("List NFT", () => {

        it("Should list NFT successfully", async () => {
            const tokenId = 1;
            await mockNFT.mint(seller.address, tokenId);

            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);

            await expect(marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("1")))
                .to.emit(marketplace, "NFTListed");
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);

            const listed = await marketplace.listedNFTs(listId);

            expect(listed.nftAddress).to.equal(await mockNFT.getAddress());
            expect(listed.tokenId).to.equal(tokenId);
            expect(listed.seller).to.equal(seller.address);
            expect(listed.price).to.equal(ethers.parseEther("1"));
            const block = await ethers.provider.getBlock("latest");
            expect(listed.expiredAt).to.equal(block.timestamp+LISTING_DURATION);
        });

        it("Should revert if NFT address is zero", async () => {
            const tokenId = 2;
            await expect(marketplace.connect(seller).listNFT(ethers.ZeroAddress, tokenId, ethers.parseEther("1")))
                .to.be.revertedWith("NFT address invalid");
        });

        it("Should revert if price <= 0", async () => {
            await expect(marketplace.connect(seller).listNFT(await mockNFT.getAddress(), 3, 0))
                .to.be.revertedWith("NFT price less than 0");
        });

        it("Should revert if NFT doesn't exist", async () => {
            await mockNFT.mint(seller.address, 4);
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), 4);
            await expect(marketplace.connect(seller).listNFT(await mockNFT.getAddress(), 5, ethers.parseEther("1")))
                .to.be.revertedWith("NFT does not exist");
        });

        it("Should revert if not owner", async () => {
            await mockNFT.mint(seller.address, 6);
            await mockNFT.connect(buyer).mint(buyer.address, 7)
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), 6);
            await expect(marketplace.connect(seller).listNFT(await mockNFT.getAddress(), 7, ethers.parseEther("1")))
                .to.be.revertedWith("Target NFT not belong to owner");
        });

        it("Should revert if not approved", async () => {
            const tokenId = 8;
            await mockNFT.mint(seller.address, tokenId);
            await expect(marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("1")))
                .to.be.revertedWith("NFT not approved for marketplace");
        });

        it("Should revert if not ERC721", async () => {
            const tokenId = 9;
            await mockNFT.mint(seller.address, tokenId);
            const NotNFT = await ethers.getContractFactory("NFTMarketPlace"); // dummy non-ERC721
            const fake = await NotNFT.deploy();
            await expect(marketplace.connect(seller).listNFT(await fake.getAddress(), tokenId, ethers.parseEther("1")))
                .to.be.revertedWith("Not ERC721");
        });

        it("Should revert if already listed", async () => {
            const tokenId = 10;
            await mockNFT.mint(seller.address, tokenId);
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("1"));
            await getListId(seller.address, await mockNFT.getAddress(), tokenId);
            await expect(marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("1")))
                .to.be.revertedWith("Target NFT not belong to owner");
        });
    });

    describe("Buy NFT", () => {

        it("Should buy NFT successfully", async () => {

            // mint
            const tokenId = 100;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            // 除了 revert tests 之外，每次 listNFT,就计算一次 listId，这样就让 userNonce 与合约中的基数相同了。
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);
            const sellerBalanceBefore = await ethers.provider.getBalance(seller.address);
            const platformBalanceBefore = await ethers.provider.getBalance(platformWallet.address);

            // buy
            await expect(marketplace.connect(buyer).buyNFT(listId, { value: ethers.parseEther("2") }))
                .to.emit(marketplace, "NFTSold");

            // Check balances
            const sellerBalanceAfter = await ethers.provider.getBalance(seller.address);
            const platformBalanceAfter = await ethers.provider.getBalance(platformWallet.address);

            const fee = ethers.parseEther("2") * BigInt(PLATFORM_FEE) / BigInt(10000); // 0.02 ETH
            const sellerAmount = ethers.parseEther("2") - fee;

            expect(sellerBalanceAfter - sellerBalanceBefore).to.equal(sellerAmount);
            expect(platformBalanceAfter - platformBalanceBefore).to.equal(fee);

            // NFT ownership
            expect(await mockNFT.ownerOf(tokenId)).to.equal(buyer.address);
            expect((await marketplace.listedNFTs(listId)).nftAddress).to.equal(ethers.ZeroAddress);
        });

        it("Should revert if NFT doesn't exist when buy NFT", async () => {
            // mint
            let tokenId = 110;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            const tx = await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
            await tx.wait();
            tokenId = 111;
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);
            await expect(marketplace.connect(buyer).buyNFT(listId, { value: ethers.parseEther("2") }))
                .to.be.revertedWith("Target NFT not exist");
        });

        it("Should revert if not exact amount", async () => {
            // mint
            const tokenId = 120;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            const tx = await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
            await tx.wait();
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);

            await expect(marketplace.connect(buyer).buyNFT(listId, { value: ethers.parseEther("1.9") }))
                .to.be.revertedWith("Must send exact amount");
        });

        it("Should revert if NFT not listed", async () => {
            const fakeId = ethers.keccak256(ethers.toUtf8Bytes("fake"));
            await expect(marketplace.connect(buyer).buyNFT(fakeId, { value: ethers.parseEther("2") }))
                .to.be.revertedWith("Target NFT not exist");
        });

        it("Should revert if expired when buy NFT", async () => {
            // mint
            const tokenId = 130;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));

            await timeTravel(LISTING_DURATION + 1);
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);
            await expect(marketplace.connect(buyer).buyNFT(listId, { value: ethers.parseEther("2") }))
                .to.be.revertedWith("NFT expired");
        });
    });

    describe("Cancel Listing", () => {

        it("Should cancel listing successfully", async () => {
            // mint
            const tokenId = 200;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));

            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);

            await expect(marketplace.connect(seller).cancelListing(listId))
                .to.emit(marketplace, "NFTCanceled");

            expect(await mockNFT.ownerOf(tokenId)).to.equal(seller.address);
            expect((await marketplace.listedNFTs(listId)).nftAddress).to.equal(ethers.ZeroAddress);
        });

        it("Should revert if not seller", async () => {
            // mint
            let tokenId = 210;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
            await getListId(seller.address, await mockNFT.getAddress(), tokenId)
            // mint
            tokenId = 220;
            await mockNFT.mint(buyer.address, tokenId);
            // approve
            await mockNFT.connect(buyer).approve(await marketplace.getAddress(), tokenId);
            // list
            await marketplace.connect(buyer).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));

            const listId = await getListId(buyer.address, await mockNFT.getAddress(), tokenId);
            await expect(marketplace.connect(seller).cancelListing(listId))
                .to.be.revertedWith("Only seller can cancel NFT");
        });

        it("Should revert if expired", async () => {
            // mint
            let tokenId = 230;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);
            await timeTravel(LISTING_DURATION + 1);
            await expect(marketplace.connect(seller).cancelListing(listId))
                .to.be.revertedWith("NFT expired");
        });
    });

    describe("Cleanup Expired Batch", () => {

        it("Should cleanup expired listings", async () => {
            // mint
            let tokenId = 300;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);

            await timeTravel(LISTING_DURATION + 1);

            await expect(marketplace.connect(logicOperator).cleanupExpiredBatch([listId]))
                .to.emit(marketplace, "NFTExpired");

            expect(await mockNFT.ownerOf(tokenId)).to.equal(seller.address);
            expect((await marketplace.listedNFTs(listId)).nftAddress).to.equal(ethers.ZeroAddress);
        });

        it("Should skip non-expired or non-existent", async () => {
            // mint
            let tokenId = 310;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);

            await timeTravel(LISTING_DURATION + 1);
            const fakeId = ethers.keccak256(ethers.toUtf8Bytes("fake"));
            await marketplace.connect(logicOperator).cleanupExpiredBatch([fakeId, listId]);
            expect(await marketplace.listedNFTs(listId).nftAddress).to.equal(undefined);
        });

        it("Should revert if not LOGIC_ROLE", async () => {
            // mint
            let tokenId = 320;
            await mockNFT.mint(seller.address, tokenId);
            // approve
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);
            // list
            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));
            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);

            await timeTravel(LISTING_DURATION + 1);
            await expect(marketplace.connect(seller).cleanupExpiredBatch([listId])).to.be.revertedWithCustomError(
                marketplace,
                "AccessControlUnauthorizedAccount"
            ).withArgs(seller.address, await marketplace.LOGIC_ROLE());
        });
    });

    describe("Settings", () => {
        it("Should update platform wallet", async () => {
            const radomAddress = random.address;
            await marketplace.connect(logicOperator).setPlatformWalletAddress(radomAddress);
            expect(await marketplace.platformWalletAddress()).to.equal(radomAddress);
        });

        it("Should revert on zero wallet", async () => {
            await expect(marketplace.connect(logicOperator).setPlatformWalletAddress(ethers.ZeroAddress))
                .to.be.revertedWith("Invalid wallet address");
        });

        it("Should update platform fee", async () => {
            await marketplace.connect(logicOperator).setPlatformFee(500);
            expect(await marketplace.platformFee()).to.equal(500);
        });

        it("Should reject fee > 1000", async () => {
            await expect(marketplace.connect(logicOperator).setPlatformFee(1001))
                .to.be.revertedWith("Fee too high");
        });

        it("Should update listing duration", async () => {
            await marketplace.connect(logicOperator).setListingDuration(60);
            expect(await marketplace.listingDuration()).to.equal(60);
        });

        it("Non-LOGIC_ROLE cannot change settings", async () => {
            await expect(marketplace.connect(seller).setPlatformWalletAddress(random.address))
                .to.be.revertedWithCustomError(
                    marketplace,
                    "AccessControlUnauthorizedAccount"
                ).withArgs(seller.address, await marketplace.LOGIC_ROLE());
        });
    });

    describe("Pause/Unpause", () => {
        it("PAUSER_ROLE can pause/unpause", async () => {
            await marketplace.connect(pauser).pause();
            expect(await marketplace.paused()).to.be.true;

            await marketplace.connect(pauser).unpause();
            expect(await marketplace.paused()).to.be.false;
        });

        it("Non-PAUSER cannot pause", async () => {
            await expect(marketplace.connect(seller).pause()).to.be.revertedWithCustomError(
                marketplace,
                "AccessControlUnauthorizedAccount"
            ).withArgs(seller.address, await marketplace.PAUSER_ROLE());
        });

        it("Paused blocks listing", async () => {
            await marketplace.connect(pauser).pause();

            const tokenId = 400;

            await mockNFT.mint(seller.address, tokenId);
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);

            await expect(marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2")))
                .to.be.revertedWithCustomError(
                    marketplace,
                    "EnforcedPause"
                );

            await marketplace.connect(pauser).unpause();
        });

        it("Paused blocks buying", async () => {
            const tokenId = 410;

            await mockNFT.mint(seller.address, tokenId);
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);

            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));

            await marketplace.connect(pauser).pause();

            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);
            await expect(marketplace.connect(buyer).buyNFT(listId, { value: ethers.parseEther("2") }))
                .to.be.revertedWithCustomError(
                    marketplace,
                    "EnforcedPause"
                );
            await marketplace.connect(pauser).unpause();

        });

        it("Paused blocks cancel", async () => {

            const tokenId = 420;

            await mockNFT.mint(seller.address, tokenId);
            await mockNFT.connect(seller).approve(await marketplace.getAddress(), tokenId);

            await marketplace.connect(seller).listNFT(await mockNFT.getAddress(), tokenId, ethers.parseEther("2"));

            const listId = await getListId(seller.address, await mockNFT.getAddress(), tokenId);
            await marketplace.connect(buyer).buyNFT(listId, { value: ethers.parseEther("2") });

            await marketplace.connect(pauser).pause();

            await expect(marketplace.connect(seller).cancelListing(listId))
                .to.be.revertedWithCustomError(
                    marketplace,
                    "EnforcedPause"
                );
            await marketplace.connect(pauser).unpause();

        });
    });

    describe("Upgradeability", () => {
        it("UPGRADER_ROLE can upgrade", async () => {
            const NFTMarketPlaceV2 = await ethers.getContractFactory("NFTMarketPlaceV2",upgrader);

            const marketplaceV2 = await upgrades.upgradeProxy(await marketplace.getAddress(), NFTMarketPlaceV2);
            expect(await marketplaceV2.testUpgrade()).to.equal(true);
        });

        it("Non-UPGRADER cannot upgrade", async () => {
            const NFTMarketPlaceV2 = await ethers.getContractFactory("NFTMarketPlaceV2",seller);

            await expect(upgrades.upgradeProxy(await marketplace.getAddress(), NFTMarketPlaceV2))
                .to.be.revertedWithCustomError(
                    marketplace,
                    "AccessControlUnauthorizedAccount"
                ).withArgs(seller.address, await marketplace.UPGRADER_ROLE());
        });
    });

    describe("onERC721Received", () => {
        it("Should return correct selector when not paused", async () => {
            const selector = await marketplace.onERC721Received.staticCall(ethers.ZeroAddress, ethers.ZeroAddress, 0, "0x");
            expect(selector).to.equal("0x150b7a02");
        });

        it("Should revert if paused (view function, so no revert in practice, but modifier is applied)", async () => {
            await marketplace.connect(pauser).pause();
            await expect(marketplace.onERC721Received(ethers.ZeroAddress, ethers.ZeroAddress, 0, "0x"))
                .to.be.revertedWith("Can't receive NFT when contract paused.");
        });
    });
});