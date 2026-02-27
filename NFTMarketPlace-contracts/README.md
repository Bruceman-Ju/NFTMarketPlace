safeTransferFrom 可能因接收方合约不支持而 revert
这是 ERC721 标准行为，无需修复，但需在文档中说明 buyer 必须能接收 NFT。

注意版本依赖，

nft 编程要验证是否存在，否则会捕获不到源码中的 ownerOf 的异常


hardhat 测试的时候，虽然不同 it 之间是相互隔离的，但是会造成 nft mint 的污染，要为每个 it 使用不同的tokenId进行测试。