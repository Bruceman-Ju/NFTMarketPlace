import MarketplaceABI from './NFTMarketPlace.json'; // 直接导入默认导出
import { MARKETPLACE_ADDRESS } from '../config';

export { MarketplaceABI };

export const marketplaceContract = {
    address: MARKETPLACE_ADDRESS as `0x${string}`,
    abi: MarketplaceABI,
} as const;
