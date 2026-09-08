import "server-only";
import { gatewayDelete, gatewayGet, gatewayPost } from "@/lib/api/gateway";

type ApiApiKey = {
  id: string;
  name: string;
  prefix: string;
  created_at: string;
  last_used_at: string | null;
  revoked_at: string | null;
};

export type ApiKeyRecord = {
  id: string;
  name: string;
  prefix: string;
  createdAt: string;
  lastUsedAt: string | null;
  revokedAt: string | null;
};

export type IssuedApiKey = {
  key: ApiKeyRecord;
  presentedKey: string;
};

const toApiKey = (key: ApiApiKey): ApiKeyRecord => ({
  id: key.id,
  name: key.name,
  prefix: key.prefix,
  createdAt: key.created_at,
  lastUsedAt: key.last_used_at,
  revokedAt: key.revoked_at,
});

export const fetchApiKeys = async (token: string): Promise<ApiKeyRecord[]> => {
  const listed = await gatewayGet<{ keys: ApiApiKey[] }>("/v1/api-keys", token);
  return listed.keys.map(toApiKey);
};

export const createApiKey = async (
  token: string,
  name: string,
  idempotencyKey: string,
): Promise<IssuedApiKey> => {
  const issued = await gatewayPost<{
    key: ApiApiKey;
    presented_key: string;
  }>("/v1/api-keys", token, { name }, idempotencyKey);

  return { key: toApiKey(issued.key), presentedKey: issued.presented_key };
};

export const revokeApiKey = async (
  token: string,
  keyId: string,
  idempotencyKey: string,
): Promise<void> => {
  await gatewayDelete(`/v1/api-keys/${keyId}`, token, idempotencyKey);
};
