// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {Initializable} from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721Receiver.sol";

contract NFTMarketPlace is Initializable, ReentrancyGuard, PausableUpgradeable, AccessControlUpgradeable, UUPSUpgradeable, IERC721Receiver {
    // --- Roles ---
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    bytes32 public constant LOGIC_ROLE = keccak256("LOGIC_ROLE");

    // --- Constants ---
    bytes4 private constant ERC721_INTERFACE_ID = 0x80ac58cd;

    // --- State variables ---
    address public platformWalletAddress;
    // e.g. 100 = 1%
    uint256 public platformFee;
    uint256 public LISTING_DURATION = 30 days;

    // --- Struct & enum & mappings---
    struct ListedNFT {
        address nftAddress;
        uint256 tokenId;
        uint256 price;
        uint256 listedTime;
        address seller;
        uint256 expiredAt;
    }

    mapping(bytes32 => ListedNFT) public listedNFTs;

    // --- Constructor ---
    constructor() {
        _disableInitializers();
    }

    function initialize(address defaultAdmin, address pauser, address upgrader, address _wallet,uint256 _platformFee)
    public
    initializer
    {
        require(_wallet != address(0), "Invalid wallet address");
        require(_platformFee <= 1000, "Fee too high");

        __Pausable_init();
        __AccessControl_init();

        _grantRole(DEFAULT_ADMIN_ROLE, defaultAdmin);
        _grantRole(PAUSER_ROLE, pauser);
        _grantRole(UPGRADER_ROLE, upgrader);

        platformWalletAddress = _wallet;
        platformFee = _platformFee;
    }

    // --- Logic functions ---
    function listNFT(address _nftAddress, uint256 tokenId, uint256 price)
    public
    whenNotPaused
    nonReentrant
    {
        require(_nftAddress != address(0),"NFT address invalid");
        require(price > 0,"NFT price less than 0");
        address owner = msg.sender;
        require(IERC721(_nftAddress).ownerOf(tokenId) == owner,"Target NFT not belong to owner");
        bool isApproved =
            IERC721(_nftAddress).getApproved(tokenId) == address(this) ||
            IERC721(_nftAddress).isApprovedForAll(owner, address(this));
        require(isApproved,"NFT not approved for marketplace");
        require(
            IERC165(_nftAddress).supportsInterface(ERC721_INTERFACE_ID),
            "Not ERC721"
        );

        uint256 expiredAt = block.timestamp + LISTING_DURATION;

        bytes32 listId = _getListedNFTId(_nftAddress, tokenId);
        require(listedNFTs[listId].nftAddress == address(0),"NFT already listed");

        listedNFTs[listId] = ListedNFT(_nftAddress, tokenId, price, block.timestamp, owner,expiredAt);

        IERC721(_nftAddress).safeTransferFrom(owner, address(this), tokenId);

        emit NFTListed(listId, _nftAddress, tokenId, price, block.timestamp, owner);
    }

    function buyNFT(bytes32 listId)
    public
    whenNotExpired(listId)
    whenNotPaused
    nonReentrant
    payable
    {
        ListedNFT storage nft = listedNFTs[listId];

        uint256 feeAmount = (nft.price * platformFee) / 10000;
        uint256 amount = nft.price - feeAmount;
        require(msg.value == nft.price,"Must send exact amount");

        delete listedNFTs[listId];

        if (amount > 0){
            (bool stateTransfer,) = payable(nft.seller).call{value: amount}("");
            require(stateTransfer, "Failed to transfer eth amount");
        }

        if(feeAmount > 0){
            (bool stateFee,) = payable(platformWalletAddress).call{value: feeAmount}("");
            require(stateFee, "Failed to collect fee");
        }

        IERC721(nft.nftAddress).safeTransferFrom(address(this), msg.sender, nft.tokenId);

        emit NFTSold(listId, nft.nftAddress, nft.tokenId, nft.price, block.timestamp, nft.seller, msg.sender);
    }

    function cancelNFT(bytes32 listId)
    public
    whenNotExpired(listId)
    whenNotPaused
    nonReentrant
    {
        ListedNFT storage nft = listedNFTs[listId];
        require(nft.seller == msg.sender,"Only seller can cancel NFT");

        delete listedNFTs[listId];
        IERC721(nft.nftAddress).safeTransferFrom(address(this), nft.seller, nft.tokenId);

        emit NFTCanceled(listId, nft.nftAddress, nft.tokenId, block.timestamp, msg.sender);
    }

    function cleanupExpiredBatch(bytes32[] memory listIds)
    public
    onlyRole(LOGIC_ROLE)
    whenNotPaused
    nonReentrant
    {
        address operator = msg.sender;
        for (uint256 i = 0; i < listIds.length; i++) {
            bytes32 listId = listIds[i];
            ListedNFT storage nft = listedNFTs[listId];
            if (nft.nftAddress != address(0) && nft.expiredAt < block.timestamp) {
                delete listedNFTs[listId];
                IERC721(nft.nftAddress).safeTransferFrom(address(this), nft.seller, nft.tokenId);
                emit NFTExpired(listId, nft.nftAddress, nft.tokenId, block.timestamp, operator);
            }
        }
    }

    // --- Internal functions & modifier---
    function _getListedNFTId(address _nftAddress, uint256 tokenId)
    internal
    pure
    returns (bytes32)
    {
        bytes32 listId = keccak256(abi.encode(_nftAddress, tokenId));
        return listId;
    }

    modifier whenNotExpired(bytes32 listId) {
        ListedNFT storage nft = listedNFTs[listId];
        require(nft.nftAddress != address(0),"Target NFT not exist");
        require(nft.expiredAt >= block.timestamp,"NFT expired");
        _;
    }


    // --- Setting functions ---
    function setPlatformWalletAddress(address _wallet)
    public
    onlyRole(LOGIC_ROLE)
    {
        require(_wallet != address(0), "Invalid wallet address");
        platformWalletAddress = _wallet;
    }

    function setPlatformFee(uint256 _fee)
    public
    onlyRole(LOGIC_ROLE)
    {
        // max fee must less than 10%
        require(_fee <= 1000, "Fee too high");
        platformFee = _fee;
    }

    function setListingDuration(uint256 _duration)
    public
    onlyRole(LOGIC_ROLE)
    {
        LISTING_DURATION = _duration;
    }

    function pause() public onlyRole(PAUSER_ROLE) {
        _pause();
    }

    function unpause() public onlyRole(PAUSER_ROLE) {
        _unpause();
    }

    function _authorizeUpgrade(address newImplementation)
    internal
    override
    onlyRole(UPGRADER_ROLE)
    {}

    function onERC721Received(
        address,
        address,
        uint256,
        bytes memory
    ) public virtual whenNotPaused view returns (bytes4) {
        return IERC721Receiver.onERC721Received.selector;
    }

    event NFTListed(
        bytes32 listId,
        address indexed nftAddress,
        uint256 indexed tokenId,
        uint256 price,
        uint256 listedTime,
        address indexed seller
    );

    event NFTSold(
        bytes32 listId,
        address indexed nftAddress,
        uint256 indexed tokenId,
        uint256 price,
        uint256 soldTime,
        address seller,
        address indexed buyer
    );

    event NFTCanceled(
        bytes32 listId,
        address indexed  nftAddress,
        uint256 indexed tokenId,
        uint256 cancelTime,
        address indexed operator
    );

    event NFTExpired(
        bytes32 listId,
        address indexed nftAddress,
        uint256 indexed tokenId,
        uint256 expiredTime,
        address indexed operator
    );
}
