// src/App.tsx
import { ConnectButton } from '@rainbow-me/rainbowkit';
import { useAccount } from 'wagmi';
import { useState, useEffect } from 'react';
import { useMarketplace } from './hooks/useMarketplace';
import { parseEther } from 'viem';

function App() {
    const { isConnected } = useAccount();
    const marketplace = useMarketplace();

    // List Form
    const [nftAddress, setNftAddress] = useState('');
    const [tokenId, setTokenId] = useState('');
    const [price, setPrice] = useState('');

    // Buy / Cancel
    const [inputListId, setInputListId] = useState('');
    const [buyPriceEth, setBuyPriceEth] = useState<string>('');

    // Auto load listed NFT when inputListId changes
    useEffect(() => {
        if (inputListId.startsWith('0x') && inputListId.length === 66) {
            console.log('Setting listId to:', inputListId);
            marketplace.setListId(inputListId as `0x${string}`);
        }
    }, [inputListId]);

    // 调试：监听listedNFT的变化
    useEffect(() => {
        console.log('listedNFT changed:', marketplace.listedNFT);
        console.log('listId:', marketplace.listId);
    }, [marketplace.listedNFT, marketplace.listId]);

    const handleList = async () => {
        try {
            await marketplace.listNFT(
                nftAddress as `0x${string}`,
                BigInt(tokenId),
                price
            );
            alert('NFT listed!');
        } catch (err) {
            console.error(err);
            alert('Listing failed');
        }
    };

    const handleBuy = async () => {
        try {
            const priceInWei = parseEther(buyPriceEth);
            console.log("Buying with price:", priceInWei);
            await marketplace.buyNFT(inputListId as `0x${string}`, priceInWei);
            alert('Purchase successful!');
        } catch (err) {
            console.error(err);
            alert('Purchase failed');
        }
    };

    const handleCancel = async () => {
        try {
            await marketplace.cancelListing(inputListId as `0x${string}`);
            alert('Listing canceled');
            setInputListId("");
        } catch (err) {
            console.error(err);
            alert('Cancel failed');
        }
    };
    return (
        <div style={{ padding: '2rem', maxWidth: '600px', margin: '0 auto' }}>
            <header style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '1rem' }}>
                <ConnectButton />
            </header>

            {!isConnected ? (
                <p>Please connect your wallet</p>
            ) : (
                <>
                    {/* List NFT */}
                    <section style={{ marginBottom: '2rem' }}>
                        <h2>List Your NFT</h2>
                        <input
                            placeholder="NFT Contract Address"
                            value={nftAddress}
                            onChange={(e) => setNftAddress(e.target.value)}
                            style={{ display: 'block', width: '100%', marginBottom: '8px' }}
                        />
                        <input
                            placeholder="Token ID"
                            value={tokenId}
                            onChange={(e) => setTokenId(e.target.value)}
                            style={{ display: 'block', width: '100%', marginBottom: '8px' }}
                        />
                        <input
                            placeholder="Price (ETH)"
                            value={price}
                            onChange={(e) => setPrice(e.target.value)}
                            style={{ display: 'block', width: '100%', marginBottom: '8px' }}
                        />
                        <button onClick={handleList} disabled={marketplace.isListing}>
                            {marketplace.isListing ? 'Listing...' : 'List NFT'}
                        </button>
                        {/*{marketplace.listId && (*/}
                        {/*    <p>*/}
                        {/*        <strong>Your List ID:</strong> {marketplace.listId}*/}
                        {/*    </p>*/}
                        {/*)}*/}
                    </section>

                    {/* Buy */}
                    <section>
                        <h2>Buy or Cancel Listing</h2>
                        <input
                            placeholder="List ID (bytes32)"
                            value={inputListId}
                            onChange={(e) => setInputListId(e.target.value)}
                            style={{ display: 'block', width: '100%', marginBottom: '8px' }}
                        />

                        {/* 调试信息 */}
                        <div style={{ backgroundColor: '#f0f0f0', padding: '10px', marginBottom: '10px' }}>
                            <p><strong>NFT Info:</strong></p>
                            <p>listId: {marketplace.listId || 'null'}</p>
                            <p>listedNFT exists: {marketplace.listedNFT?.price ? 'Yes' : 'No'}</p>
                            {marketplace.listedNFT && (
                                <>
                                    <p>Price (wei): {marketplace.listedNFT.price?.toString()}</p>
                                    <p>Price (ETH): {marketplace.listedNFT.price ? Number(marketplace.listedNFT.price) / 1e18 : 'N/A'} ETH</p>
                                    <p>Seller: {marketplace.listedNFT.seller}</p>
                                </>
                            )}
                        </div>
                        <button
                            onClick={handleCancel}
                            disabled={marketplace.isCanceling}
                            style={{ marginLeft: '8px' }}
                        >
                            {marketplace.isCanceling ? 'Canceling...' : 'Cancel Listing'}
                        </button>
                        <div></div>
                        {marketplace.listedNFT?.price && (
                            <p>
                                Price: {Number(marketplace.listedNFT.price) / 1e18} ETH
                                <button
                                    onClick={() => setBuyPriceEth(String(Number(marketplace.listedNFT!.price) / 1e18))}
                                    style={{ marginLeft: '8px' }}
                                >
                                    Use This Price
                                </button>
                            </p>
                        )}
                        <input
                            placeholder="Buy Price (ETH)"
                            value={buyPriceEth}
                            onChange={(e) => setBuyPriceEth(e.target.value)}
                            style={{ display: 'block', width: '100%', marginBottom: '8px' }}
                        />
                        <button onClick={handleBuy} disabled={marketplace.isBuying}>
                            {marketplace.isBuying ? 'Buying...' : 'Buy NFT'}
                        </button>
                    </section>
                </>
            )}
        </div>
    );
}

export default App;
