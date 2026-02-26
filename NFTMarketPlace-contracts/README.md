safeTransferFrom 可能因接收方合约不支持而 revert
这是 ERC721 标准行为，无需修复，但需在文档中说明 buyer 必须能接收 NFT。