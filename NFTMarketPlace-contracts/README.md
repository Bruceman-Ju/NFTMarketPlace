# NFTMarketPlace 智能合约

一个安全、可升级、可暂停的 NFT 交易市场，基于以太坊使用 Solidity 和 OpenZeppelin 构建。

该智能合约允许用户上架、购买、取消以及批量清理已过期的 ERC-721 NFT 列表，同时平台运营方可收取可配置的手续费。
合约采用 UUPS 代理模式实现逻辑升级，并集成了基于角色的访问控制（RBAC）和暂停机制，以增强安全性。

合约采用 safeTransferFrom 对 NFT 进行所属权操作，这是 ERC721 标准行为，请确保购买者必须能接收 NFT。

## 核心功能

- 上架 NFT：卖家在授权市场后，可将自己拥有的 ERC-721 NFT 上架出售。
- 购买 NFT：买家可通过发送精确金额的 ETH 购买已上架的 NFT。
- 取消上架：卖家可随时取消自己的有效上架。
- 过期处理：每个上架项在设定时间（默认 30 天）后代表失效。
- 批量清理：授权操作员可批量回收已过期的 NFT 并归还给原卖家。
- 平台手续费：每笔交易可收取最高 10% 的手续费，发送至指定钱包地址。
- 可升级性：采用 UUPS 代理模式，支持由授权角色升级合约逻辑。
- 可暂停性：紧急情况下，可由指定角色暂停全部交易功能。
- 安全转账：使用 safeTransferFrom 并正确实现 IERC721Receiver 接口。
- 角色权限分离：管理员、暂停员、升级员、逻辑操作员职责分明。

## 角色说明
| 角色 | 权限说明 |
|------|----------|
| `DEFAULT_ADMIN_ROLE` | 初始化时指定，拥有所有角色的管理权限。 |
| `PAUSER_ROLE` | 可暂停/恢复整个市场交易功能。 |
| `UPGRADER_ROLE` | 可通过 UUPS 模式升级合约逻辑。 |
| `LOGIC_ROLE` | 可修改平台参数（手续费、收款地址、上架时长）并执行过期 NFT 批量清理。 |

⚠️ 注意：LOGIC_ROLE 不能执行上架或购买操作，仅用于配置与维护。

## 关键参数
| 参数 | 默认值 | 说明 |
|------|--------|------|
| `platformFee` | 初始化设置 | 平台手续费（单位：基点，例如 `100` = 1%，最大 `1000` = 10%）。 |
| `listingDuration` | `30 days` | 上架有效期（秒），超过后自动过期。 |
| `platformWalletAddress` | 初始化设置 | 手续费收款地址。 |

所有参数均可在部署后由 LOGIC_ROLE 动态更新。

## 主要函数

### 用户功能
- listNFT(address nftAddress, uint256 tokenId, uint256 price)  
上架 NFT（需先授权市场）。
- buyNFT(bytes32 listId)  
购买指定上架项（必须发送精确 ETH 金额）。
- cancelListing(bytes32 listId)  
取消自己的有效上架。

### 运营功能（需 LOGIC_ROLE）
- cleanupExpiredBatch(bytes32[] memory listIds)  
批量清理过期上架，将 NFT 归还卖家。
- setPlatformWalletAddress(address)
- setPlatformFee(uint256)
- setListingDuration(uint256)

### 管理与安全
- pause() / unpause() — 由 PAUSER_ROLE 调用
- _authorizeUpgrade(...) — 由 UPGRADER_ROLE 调用

### 事件（Events）

- NFTListed：NFT 上架
- NFTSold：NFT 成交
- NFTCanceled：上架取消
- NFTExpired：上架过期（由批量清理触发）

所有事件均包含索引字段，便于链下高效查询。

## 技术细节

Solidity 版本：^0.8.27

代理模式：UUPS（通过 UUPSUpgradeable 实现）

继承结构：
- Initializable
- PausableUpgradeable
- AccessControlUpgradeable
- IERC721Receiver

遵循标准：ERC-721、ERC-165（接口检测）

重入防护：采用“检查-生效-交互”模式，ETH 转账使用带返回值校验的 .call{value:}

## 部署与初始化

合约必须作为代理合约部署。

部署后，仅一次调用 initialize 函数：

``` solidity
initialize(
    address defaultAdmin,     // 默认管理员
    address pauser,           // 暂停角色
    address upgrader,         // 升级角色
    address logicOperator,    // 逻辑操作员
    address platformWallet,   // 平台收款地址
    uint256 platformFee       // 手续费率（≤1000）
)
```

❗ initialize 受 initializer 修饰符保护，只能调用一次。

## 安全注意事项
所有 ETH 转账均使用低级 call 并验证返回值，防止转账失败。

上架前严格校验 NFT 所有权及市场授权状态。

合约暂停时拒绝接收任何 NFT（onERC721Received 中强制检查）。

手续费率上限为 10%，避免恶意配置。

过期上架不会自动清理，需显式调用清理函数，防止 Gas 拒绝服务攻击。

## 依赖库

- OpenZeppelin Contracts Upgradeable v5+
- AccessControlUpgradeable
- PausableUpgradeable
- UUPSUpgradeable
- Initializable
- IERC721, IERC721Receiver, IERC165

## 合约地址
已经部署一个合约，地址：
0x0559F2a6055Ac64aBEb3feFE4F36417C6676aF4b
