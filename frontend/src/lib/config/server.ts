import "server-only";
import { z } from "zod";

const coreEnvironmentSchema = z.object({
  API_GATEWAY_URL: z.url(),
  APP_BASE_URL: z.url(),
});

const coreEnvironment = coreEnvironmentSchema.safeParse({
  API_GATEWAY_URL: process.env.API_GATEWAY_URL,
  APP_BASE_URL: process.env.APP_BASE_URL,
});

if (!coreEnvironment.success) {
  throw new Error(
    `Server environment configuration is invalid.\n${z.prettifyError(coreEnvironment.error)}`,
  );
}

export const serverConfig = {
  apiGatewayUrl: coreEnvironment.data.API_GATEWAY_URL,
  appBaseUrl: coreEnvironment.data.APP_BASE_URL,
} as const;

const auth0EnvironmentSchema = z.object({
  AUTH0_DOMAIN: z.string().min(1),
  AUTH0_CLIENT_ID: z.string().min(1),
  AUTH0_CLIENT_SECRET: z.string().min(1),
  AUTH0_AUDIENCE: z.string().min(1),
  AUTH0_SECRET: z.string().min(32, "Generate with: openssl rand -hex 32"),
});

export const auth0Config = () => {
  const parsed = auth0EnvironmentSchema.safeParse({
    AUTH0_DOMAIN: process.env.AUTH0_DOMAIN,
    AUTH0_CLIENT_ID: process.env.AUTH0_CLIENT_ID,
    AUTH0_CLIENT_SECRET: process.env.AUTH0_CLIENT_SECRET,
    AUTH0_AUDIENCE: process.env.AUTH0_AUDIENCE,
    AUTH0_SECRET: process.env.AUTH0_SECRET,
  });

  if (!parsed.success) {
    throw new Error(
      `Auth0 configuration is invalid.\n${z.prettifyError(parsed.error)}`,
    );
  }

  return {
    domain: parsed.data.AUTH0_DOMAIN,
    clientId: parsed.data.AUTH0_CLIENT_ID,
    clientSecret: parsed.data.AUTH0_CLIENT_SECRET,
    audience: parsed.data.AUTH0_AUDIENCE,
    sessionSecret: parsed.data.AUTH0_SECRET,
  } as const;
};
