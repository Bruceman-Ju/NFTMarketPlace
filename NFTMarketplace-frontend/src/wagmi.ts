import { getDefaultConfig } from '@rainbow-me/rainbowkit';
import { sepolia } from 'viem/chains';
import { http } from 'viem';

const WALLET_CONNECT_PROJECT_ID = 'ac9198dda017a633d6a69a8309722f3d'; // 可选，留空也可用 MetaMask

export const config = getDefaultConfig({
    appName: 'NFT Marketplace',
    projectId: WALLET_CONNECT_PROJECT_ID,
    chains: [sepolia],
    transports: {
        [sepolia.id]: http(), // 使用公共 RPC 或 Infura
    },
    ssr: true,
});

declare module 'wagmi' {
    interface Register {
        config: typeof config;
    }
}