import "server-only";
import { z } from "zod";

const serverEnvironmentSchema = z.object({
  API_GATEWAY_URL: z.url(),
  APP_BASE_URL: z.url(),
  AUTH0_DOMAIN: z.string().min(1),
  AUTH0_CLIENT_ID: z.string().min(1),
  AUTH0_CLIENT_SECRET: z.string().min(1),
  AUTH0_AUDIENCE: z.string().min(1),
  AUTH0_SECRET: z.string().min(32, "Generate with: openssl rand -hex 32"),
});

const serverEnvironment = serverEnvironmentSchema.safeParse({
  API_GATEWAY_URL: process.env.API_GATEWAY_URL,
  APP_BASE_URL: process.env.APP_BASE_URL,
  AUTH0_DOMAIN: process.env.AUTH0_DOMAIN,
  AUTH0_CLIENT_ID: process.env.AUTH0_CLIENT_ID,
  AUTH0_CLIENT_SECRET: process.env.AUTH0_CLIENT_SECRET,
  AUTH0_AUDIENCE: process.env.AUTH0_AUDIENCE,
  AUTH0_SECRET: process.env.AUTH0_SECRET,
});

if (!serverEnvironment.success) {
  throw new Error(
    `Server environment configuration is invalid.\n${z.prettifyError(serverEnvironment.error)}`,
  );
}

export const serverConfig = {
  apiGatewayUrl: serverEnvironment.data.API_GATEWAY_URL,
  appBaseUrl: serverEnvironment.data.APP_BASE_URL,
  auth0: {
    domain: serverEnvironment.data.AUTH0_DOMAIN,
    clientId: serverEnvironment.data.AUTH0_CLIENT_ID,
    clientSecret: serverEnvironment.data.AUTH0_CLIENT_SECRET,
    audience: serverEnvironment.data.AUTH0_AUDIENCE,
    sessionSecret: serverEnvironment.data.AUTH0_SECRET,
  },
} as const;
