import "server-only";
import { Auth0Client } from "@auth0/nextjs-auth0/server";
import { auth0Config, serverConfig } from "@/lib/config/server";

const config = auth0Config();

export const auth0 = new Auth0Client({
  domain: config.domain,
  clientId: config.clientId,
  clientSecret: config.clientSecret,
  secret: config.sessionSecret,
  appBaseUrl: serverConfig.appBaseUrl,
  authorizationParameters: {
    audience: config.audience,
    scope: "openid profile email offline_access",
  },
});
