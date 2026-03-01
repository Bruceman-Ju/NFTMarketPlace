// src/utils/listId.ts
import { encodeAbiParameters, keccak256, parseAbiParameters } from 'viem';

export function generateListId(nftAddress: string, tokenId: bigint): `0x${string}` {
    const encoded = encodeAbiParameters(parseAbiParameters(['address', 'uint256']), [
        nftAddress as `0x${string}`,
        tokenId,
    ]);
    return keccak256(encoded);
}