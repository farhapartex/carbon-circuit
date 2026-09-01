import "server-only";
import { Auth0Client } from "@auth0/nextjs-auth0/server";
import { fetchMe } from "@/lib/api/me";
import { auth0Config, serverConfig } from "@/lib/config/server";
import { TENANCY_KEY, tenancyFrom } from "@/lib/session";

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
  async beforeSessionSaved(session) {
    const accessToken = session.tokenSet.accessToken;

    if (!accessToken) {
      return session;
    }

    try {
      return {
        ...session,
        [TENANCY_KEY]: tenancyFrom(await fetchMe(accessToken)),
      };
    } catch {
      return session;
    }
  },
});
