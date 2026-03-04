safeTransferFrom 可能因接收方合约不支持而 revert
这是 ERC721 标准行为，无需修复，但需在文档中说明 buyer 必须能接收 NFT。

注意版本依赖，

nft 编程要验证是否存在，否则会捕获不到源码中的 ownerOf 的异常


hardhat 测试的时候，虽然不同 it 之间是相互隔离的，但是会造成 nft mint 的污染，要为每个 it 使用不同的tokenId进行测试。

部署了两个合约到 sepolia 上
0x71927479bfcD4361a7Bc39F38aAB565abd88E369
0x9A0e6bEdf97657FB93dAB575334d76eC500e58CA
v2 版本：bug-fix 重新编写 listed NFT id 生成函数，支持统一个 NFT 重复上架。
0x3BD7E1039686B774E0c297528c4b7d0CFd439922
0x0559F2a6055Ac64aBEb3feFE4F36417C6676aF4b
