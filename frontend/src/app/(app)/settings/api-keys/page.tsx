import type { Metadata } from "next";
import { ApiKeyManager } from "@/components/features/settings/ApiKeyManager";
import { fetchApiKeys } from "@/lib/api/apiKeys";
import { auth0 } from "@/lib/auth0";

export const metadata: Metadata = { title: "API keys" };

export default async function SettingsApiKeysPage() {
  const { token } = await auth0.getAccessToken();
  const keys = await fetchApiKeys(token);

  return <ApiKeyManager keys={keys} />;
}
