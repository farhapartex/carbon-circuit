import "server-only";
import { z } from "zod";

const serverEnvironmentSchema = z.object({
  API_GATEWAY_URL: z.url(),
});

const serverEnvironment = serverEnvironmentSchema.safeParse({
  API_GATEWAY_URL: process.env.API_GATEWAY_URL,
});

if (!serverEnvironment.success) {
  throw new Error(
    `Server environment configuration is invalid.\n${z.prettifyError(serverEnvironment.error)}`,
  );
}

export const serverConfig = {
  apiGatewayUrl: serverEnvironment.data.API_GATEWAY_URL,
} as const;
