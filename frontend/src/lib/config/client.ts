import { z } from "zod";

const publicEnvironmentSchema = z.object({
  NEXT_PUBLIC_APP_URL: z.url(),
  NEXT_PUBLIC_CHAIN_ID: z.coerce.number().int().positive(),
  NEXT_PUBLIC_RPC_URL: z.url(),
  NEXT_PUBLIC_EXPLORER_BASE_URL: z.url(),
  NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID: z.string().min(1),
});

const publicEnvironment = publicEnvironmentSchema.safeParse({
  NEXT_PUBLIC_APP_URL: process.env.NEXT_PUBLIC_APP_URL,
  NEXT_PUBLIC_CHAIN_ID: process.env.NEXT_PUBLIC_CHAIN_ID,
  NEXT_PUBLIC_RPC_URL: process.env.NEXT_PUBLIC_RPC_URL,
  NEXT_PUBLIC_EXPLORER_BASE_URL: process.env.NEXT_PUBLIC_EXPLORER_BASE_URL,
  NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID:
    process.env.NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID,
});

if (!publicEnvironment.success) {
  throw new Error(
    `Public environment configuration is invalid.\n${z.prettifyError(publicEnvironment.error)}`,
  );
}

export const clientConfig = {
  appUrl: publicEnvironment.data.NEXT_PUBLIC_APP_URL,
  chainId: publicEnvironment.data.NEXT_PUBLIC_CHAIN_ID,
  rpcUrl: publicEnvironment.data.NEXT_PUBLIC_RPC_URL,
  explorerBaseUrl: publicEnvironment.data.NEXT_PUBLIC_EXPLORER_BASE_URL,
  walletConnectProjectId:
    publicEnvironment.data.NEXT_PUBLIC_WALLETCONNECT_PROJECT_ID,
} as const;
