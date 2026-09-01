import "server-only";
import { serverConfig } from "@/lib/config/server";

export class GatewayError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
  ) {
    super(`Gateway responded ${status} ${code}`);
  }
}

const errorCodeFrom = async (response: Response): Promise<string> => {
  try {
    const body = (await response.json()) as { error?: { code?: string } };
    return body.error?.code ?? "UNKNOWN";
  } catch {
    return "UNPARSEABLE";
  }
};

export const gatewayGet = async <T>(
  path: string,
  token: string,
): Promise<T> => {
  const response = await fetch(new URL(path, serverConfig.apiGatewayUrl), {
    headers: { Accept: "application/json", Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  if (!response.ok) {
    throw new GatewayError(response.status, await errorCodeFrom(response));
  }

  const body = (await response.json()) as { data: T };
  return body.data;
};

export const gatewayPost = async <T>(
  path: string,
  token: string,
  body: unknown,
  idempotencyKey: string,
): Promise<T> => {
  const response = await fetch(new URL(path, serverConfig.apiGatewayUrl), {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      "Idempotency-Key": idempotencyKey,
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
    cache: "no-store",
  });

  if (!response.ok) {
    throw new GatewayError(response.status, await errorCodeFrom(response));
  }

  const payload = (await response.json()) as { data: T };
  return payload.data;
};
