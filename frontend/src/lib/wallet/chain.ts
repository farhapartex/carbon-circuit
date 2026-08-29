import { defineChain } from "viem";
import { base, baseSepolia } from "viem/chains";
import { clientConfig } from "@/lib/config/client";

const anvil = defineChain({
  id: 31337,
  name: "Anvil",
  nativeCurrency: { name: "Ether", symbol: "ETH", decimals: 18 },
  rpcUrls: { default: { http: [clientConfig.rpcUrl] } },
});

const supportedChains = [base, baseSepolia, anvil];

const resolveChain = () => {
  const match = supportedChains.find(
    (candidate) => candidate.id === clientConfig.chainId,
  );
  if (!match) {
    throw new Error(
      `NEXT_PUBLIC_CHAIN_ID ${clientConfig.chainId} is not a supported chain. Supported: ${supportedChains.map((c) => c.id).join(", ")}`,
    );
  }
  return match;
};

export const expectedChain = resolveChain();

export const explorerTransactionUrl = (transactionHash: string) =>
  `${clientConfig.explorerBaseUrl.replace(/\/$/, "")}/tx/${transactionHash}`;

export const explorerAddressUrl = (address: string) =>
  `${clientConfig.explorerBaseUrl.replace(/\/$/, "")}/address/${address}`;
