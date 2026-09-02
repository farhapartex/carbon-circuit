import "server-only";
import { gatewayPost } from "@/lib/api/gateway";

type ApiNonce = {
  nonce: string;
  domain: string;
  chain_id: string;
  issued_at: string;
  expires_at: string;
};

type ApiTreasury = {
  address: string;
  designated_at: string;
};

export type TreasuryChallenge = {
  nonce: string;
  domain: string;
  chainId: number;
  issuedAt: string;
  expiresAt: string;
};

export type DesignatedTreasury = {
  address: string;
  designatedAt: string;
};

export const requestTreasuryChallenge = async (
  token: string,
  idempotencyKey: string,
): Promise<TreasuryChallenge> => {
  const nonce = await gatewayPost<ApiNonce>(
    "/v1/treasury/nonce",
    token,
    {},
    idempotencyKey,
  );

  return {
    nonce: nonce.nonce,
    domain: nonce.domain,
    chainId: Number(nonce.chain_id),
    issuedAt: nonce.issued_at,
    expiresAt: nonce.expires_at,
  };
};

export const designateTreasury = async (
  token: string,
  message: string,
  signature: string,
  idempotencyKey: string,
): Promise<DesignatedTreasury> => {
  const designated = await gatewayPost<ApiTreasury>(
    "/v1/treasury",
    token,
    { message, signature },
    idempotencyKey,
  );

  return {
    address: designated.address,
    designatedAt: designated.designated_at,
  };
};
