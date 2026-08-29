"use client";

import { RainbowKitProvider, lightTheme } from "@rainbow-me/rainbowkit";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";
import { WagmiProvider } from "wagmi";
import { wagmiConfig } from "@/lib/wallet/config";

import "@rainbow-me/rainbowkit/styles.css";

const carbonCircuitWalletTheme = lightTheme({
  accentColor: "var(--color-primary-700)",
  accentColorForeground: "var(--color-white)",
  borderRadius: "medium",
  fontStack: "system",
});

export const WalletProvider = ({ children }: { children: ReactNode }) => {
  const [queryClient] = useState(() => new QueryClient());

  return (
    <WagmiProvider config={wagmiConfig}>
      <QueryClientProvider client={queryClient}>
        <RainbowKitProvider
          theme={carbonCircuitWalletTheme}
          modalSize="compact"
        >
          {children}
        </RainbowKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  );
};
