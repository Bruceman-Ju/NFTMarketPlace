import {useMemo, useState} from 'react';
import {
    useWriteContract,
    useReadContract,
    useAccount,
} from 'wagmi';
import { parseEther } from 'viem';
import { marketplaceContract } from '../contracts/marketplace';


// 定义 ListedNFT 类型
interface ListedNFT {
    nftAddress: `0x${string}`;
    tokenId: bigint;
    price: bigint;
    listedTime: bigint;
    seller: `0x${string}`;
    expiredAt: bigint;
}

export function useMarketplace() {
    const { address } = useAccount();
    const [listId, setListId] = useState<`0x${string}` | null>(null);

    // --- List NFT ---
    const {
        writeContractAsync: listNFT,
        isPending: isListing,
    } = useWriteContract();

    const listNFTWithApprove = async (
        nftAddress: `0x${string}`,
        tokenId: bigint,
        priceInEth: string
    ) => {
        if (!address) throw new Error('Not connected');

        const price = parseEther(priceInEth); // ETH → Wei
        // const id = generateListId(nftAddress, tokenId);
        // setListId(id);

        // Step 1: Approve marketplace
        await listNFT({
            address: nftAddress,
            abi: [
                {
                    name: 'approve',
                    type: 'function',
                    inputs: [{ name: 'to', type: 'address' }, { name: 'tokenId', type: 'uint256' }],
                    outputs: [],
                    stateMutability: 'nonpayable',
                },
            ],
            functionName: 'approve',
            args: [marketplaceContract.address, tokenId],
        });

        // Step 2: List on marketplace
        await listNFT({
            ...marketplaceContract,
            functionName: 'listNFT',
            args: [nftAddress, tokenId, price],
            gas: 1_000_000n, // 通常 500k~1M 足够
        });
    };

    // --- Buy NFT ---
    const {
        writeContractAsync: buyNFT,
        isPending: isBuying,
    } = useWriteContract();

    const buyNFTById = async (listId: `0x${string}`, price: bigint) => {
        await buyNFT({
            ...marketplaceContract,
            functionName: 'buyNFT',
            args: [listId],
            value: price, // 必须发送 exact amount
        });
    };

    // --- Cancel Listing ---
    const {
        writeContractAsync: cancelListing,
        isPending: isCanceling,
    } = useWriteContract();

    const cancelListingById = async (listId: `0x${string}`) => {
        await cancelListing({
            ...marketplaceContract,
            functionName: 'cancelListing',
            args: [listId],
        });
    };

    // --- Read listed NFT ---
    const { data: rawData, refetch, isError, error } = useReadContract({
        ...marketplaceContract,
        functionName: 'listedNFTs',
        args: listId ? [listId] : undefined,
        query: {
            enabled: !!listId,
            refetchOnWindowFocus: true,
            refetchInterval: 5000 // 每5秒刷新一次
        },
    }) as {
        data: ListedNFT | undefined;
        refetch: () => void;
        isError: boolean;
        error: Error | null;
    };

// 关键：手动将数组数据映射到 ListedNFT 对象
    const listedNFT = useMemo((): ListedNFT | undefined => {
        // 检查原始数据是否存在且为数组
        if (!rawData || !Array.isArray(rawData)) {
            console.log('Raw data is not array or undefined:', rawData);
            return undefined;
        }

        console.log('Raw data received:', rawData);

        // 解构数组元素（按照 Solidity struct 的顺序）
        const [nftAddress, tokenId, price, listedTime, seller, expiredAt] = rawData;
        console.log(nftAddress,tokenId,price,listedTime,seller,expiredAt);

        try {
            // 创建并返回映射后的对象
            const mappedNFT: ListedNFT = {
                nftAddress: nftAddress as `0x${string}`,
                tokenId: BigInt(tokenId.toString()),
                price: BigInt(price.toString()),
                listedTime: BigInt(listedTime.toString()),
                seller: seller as `0x${string}`,
                expiredAt: BigInt(expiredAt.toString())
            };

            console.log('Successfully mapped NFT data:', mappedNFT);
            return mappedNFT;
        } catch (error) {
            console.error('Error mapping NFT data:', error);
            return undefined;
        }
    }, [rawData]); // 依赖 rawData 变化

    // 调试：打印错误信息
    if (isError) {
        console.error('Read contract error:', error);
    }

    console.log('useMarketplace hook state:', {
        listId,
        listedNFT,
        hasListId: !!listId,
        hasData: !!listedNFT
    });

    return {
        listNFT: listNFTWithApprove,
        buyNFT: buyNFTById,
        cancelListing: cancelListingById,
        listedNFT,
        listId,
        setListId,
        isListing,
        isBuying,
        isCanceling,
        refetchListedNFT: refetch,
    };
}
