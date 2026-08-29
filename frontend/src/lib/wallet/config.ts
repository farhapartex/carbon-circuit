import { getDefaultConfig } from "@rainbow-me/rainbowkit";
import {
  coinbaseWallet,
  metaMaskWallet,
  walletConnectWallet,
} from "@rainbow-me/rainbowkit/wallets";
import { http } from "wagmi";
import { clientConfig } from "@/lib/config/client";
import { expectedChain } from "@/lib/wallet/chain";

export const wagmiConfig = getDefaultConfig({
  appName: "CarbonCircuit",
  projectId: clientConfig.walletConnectProjectId,
  chains: [expectedChain],
  transports: { [expectedChain.id]: http(clientConfig.rpcUrl) },
  wallets: [
    {
      groupName: "Supported",
      wallets: [metaMaskWallet, walletConnectWallet, coinbaseWallet],
    },
  ],
  ssr: true,
});
