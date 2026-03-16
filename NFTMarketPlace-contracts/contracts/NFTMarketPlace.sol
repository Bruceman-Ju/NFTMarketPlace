// SPDX-License-Identifier: MIT
pragma solidity ^0.8.27;

import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {Initializable} from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts/proxy/utils/UUPSUpgradeable.sol";
import {IERC721} from "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import {IERC721Receiver} from "@openzeppelin/contracts/token/ERC721/IERC721Receiver.sol";
import {IERC165} from "@openzeppelin/contracts/utils/introspection/IERC165.sol";

contract NFTMarketPlace is Initializable, PausableUpgradeable, AccessControlUpgradeable, UUPSUpgradeable, IERC721Receiver {
    // --- Roles ---
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
    bytes32 public constant UN_PAUSER_ROLE = keccak256("UN_PAUSER_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    bytes32 public constant LOGIC_ROLE = keccak256("LOGIC_ROLE");

    // --- Constants ---
    bytes4 private constant ERC721_INTERFACE_ID = 0x80ac58cd;

    // --- State variables ---
    address public platformWalletAddress;
    // e.g. 100 = 1%
    uint256 public platformFee;
    uint256 public listingDuration;
    mapping(address => uint256) private userNonce;

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

    function initialize(
        address _defaultAdmin, address _pauser, address _unPauser,
        address _upgrader, address _logicOperator, address _wallet,
        uint256 _platformFee)
    public
    initializer
    {
        require(_wallet != address(0), "Invalid wallet address");
        require(_platformFee <= 1000, "Fee too high");

        __Pausable_init();
        __AccessControl_init();

        _grantRole(DEFAULT_ADMIN_ROLE, _defaultAdmin);
        _grantRole(PAUSER_ROLE, _pauser);
        _grantRole(UN_PAUSER_ROLE, _unPauser);
        _grantRole(UPGRADER_ROLE, _upgrader);
        _grantRole(LOGIC_ROLE, _logicOperator);

        platformWalletAddress = _wallet;
        platformFee = _platformFee;
        listingDuration = 30 days;
    }

    // --- Logic functions ---
    function listNFT(address _nftAddress, uint256 tokenId, uint256 price)
    public
    whenNotPaused
    {
        require(_nftAddress != address(0), "NFT address invalid");
        require(price > 0, "NFT price less than 0");
        // check whether _nftAddress is instance of IERC721
        require(
            IERC165(_nftAddress).supportsInterface(ERC721_INTERFACE_ID),
            "Not ERC721"
        );

        // check whether nft exist
        address owner;
        try IERC721(_nftAddress).ownerOf(tokenId) returns (address _owner) {
            owner = _owner;
        } catch {
            revert("NFT does not exist");
        }

        require(msg.sender == owner, "Target NFT not belong to owner");
        bool isApproved =
            IERC721(_nftAddress).getApproved(tokenId) == address(this) ||
            IERC721(_nftAddress).isApprovedForAll(owner, address(this));
        require(isApproved, "NFT not approved for marketplace");

        uint256 expiredAt = block.timestamp + listingDuration;

        bytes32 listId = _getListedNFTId(_nftAddress, tokenId);
        listedNFTs[listId] = ListedNFT(_nftAddress, tokenId, price, block.timestamp, owner, expiredAt);

        IERC721(_nftAddress).safeTransferFrom(owner, address(this), tokenId);

        emit NFTListed(listId, _nftAddress, tokenId, price, block.timestamp, owner);
    }

    function buyNFT(bytes32 listId)
    public
    whenNotPaused
    whenNFTExistAndNotExpired(listId)
    payable
    {
        ListedNFT storage nft = listedNFTs[listId];

        address nftAddress_ = nft.nftAddress;
        uint256 price_ = nft.price;
        address seller_ = nft.seller;
        uint256 tokenId_ = nft.tokenId;

        delete listedNFTs[listId];

        uint256 feeAmount = (price_ * platformFee) / 10000;
        uint256 amount = price_ - feeAmount;
        require(msg.value == price_, "Must send exact amount");

        (bool stateTransfer,) = payable(seller_).call{value: amount}("");
        require(stateTransfer, "Failed to transfer eth amount");

        (bool stateFee,) = payable(platformWalletAddress).call{value: feeAmount}("");
        require(stateFee, "Failed to collect fee");

        IERC721(nftAddress_).safeTransferFrom(address(this), msg.sender, tokenId_);

        emit NFTSold(listId, nftAddress_, tokenId_, price_, block.timestamp, seller_, msg.sender);
    }

    function cancelListing(bytes32 listId)
    public
    whenNotPaused
    whenNFTExistAndNotExpired(listId)
    {
        ListedNFT storage nft = listedNFTs[listId];
        require(nft.seller == msg.sender, "Only seller can cancel NFT");

        address nftAddress_ = nft.nftAddress;
        address seller_ = nft.seller;
        uint256 tokenId_ = nft.tokenId;

        delete listedNFTs[listId];

        IERC721(nftAddress_).safeTransferFrom(address(this), seller_, tokenId_);

        emit NFTCanceled(listId, nftAddress_, tokenId_, block.timestamp, msg.sender);
    }

    function cleanupExpiredBatch(bytes32[] memory listIds)
    public
    onlyRole(LOGIC_ROLE)
    whenNotPaused
    {
        address operator = msg.sender;
        for (uint256 i = 0; i < listIds.length; i++) {
            bytes32 listId = listIds[i];
            ListedNFT storage nft = listedNFTs[listId];
            if (nft.nftAddress != address(0) && nft.expiredAt < block.timestamp) {
                address nftAddress_ = nft.nftAddress;
                address seller_ = nft.seller;
                uint256 tokenId_ = nft.tokenId;
                delete listedNFTs[listId];
                IERC721(nftAddress_).safeTransferFrom(address(this), seller_, tokenId_);
                emit NFTExpired(listId, nftAddress_, tokenId_, block.timestamp, operator);
            }
        }
    }

    // --- Internal functions & modifier---
    function _getListedNFTId(address _nftAddress, uint256 tokenId)
    internal
    returns (bytes32)
    {
        uint256 userNon = userNonce[msg.sender]++;
        bytes32 listId = keccak256(abi.encode(msg.sender, userNon, _nftAddress, tokenId));
        return listId;
    }

    modifier whenNFTExistAndNotExpired(bytes32 listId) {
        ListedNFT storage nft = listedNFTs[listId];
        require(nft.nftAddress != address(0), "Target NFT not exist");
        require(nft.expiredAt >= block.timestamp, "NFT expired");
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
        require(_fee > 0, "Fee must greater than 0");
        platformFee = _fee;
    }

    function setListingDuration(uint256 _duration)
    public
    onlyRole(LOGIC_ROLE)
    {
        listingDuration = _duration;
    }

    function pause() public onlyRole(PAUSER_ROLE) {
        _pause();
    }

    function unpause() public onlyRole(UN_PAUSER_ROLE) {
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
    ) external view returns (bytes4) {
        require(!paused(), "Can't receive NFT when contract paused.");
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
