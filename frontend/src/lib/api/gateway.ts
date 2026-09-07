import "server-only";
import { headers } from "next/headers";
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

const deviceHeaders = async (): Promise<Record<string, string>> => {
  const incoming = await headers();

  const forwarded: Record<string, string> = {};

  const userAgent = incoming.get("user-agent");
  if (userAgent) forwarded["User-Agent"] = userAgent;

  const address = incoming.get("x-forwarded-for") ?? incoming.get("x-real-ip");
  if (address) forwarded["X-Forwarded-For"] = address;

  return forwarded;
};

export const gatewayGet = async <T>(
  path: string,
  token: string,
): Promise<T> => {
  const response = await fetch(new URL(path, serverConfig.apiGatewayUrl), {
    headers: {
      Accept: "application/json",
      Authorization: `Bearer ${token}`,
      ...(await deviceHeaders()),
    },
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
      ...(await deviceHeaders()),
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

const mutate = async (
  method: "PATCH" | "DELETE",
  path: string,
  token: string,
  body: unknown,
  idempotencyKey: string,
): Promise<Response> => {
  const requestHeaders: Record<string, string> = {
    Accept: "application/json",
    "Idempotency-Key": idempotencyKey,
    Authorization: `Bearer ${token}`,
    ...(await deviceHeaders()),
  };

  if (body !== undefined) {
    requestHeaders["Content-Type"] = "application/json";
  }

  const response = await fetch(new URL(path, serverConfig.apiGatewayUrl), {
    method,
    headers: requestHeaders,
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    cache: "no-store",
  });

  if (!response.ok) {
    throw new GatewayError(response.status, await errorCodeFrom(response));
  }

  return response;
};

export const gatewayPatch = async <T>(
  path: string,
  token: string,
  body: unknown,
  idempotencyKey: string,
): Promise<T> => {
  const response = await mutate("PATCH", path, token, body, idempotencyKey);
  const payload = (await response.json()) as { data: T };
  return payload.data;
};

export const gatewayDelete = async (
  path: string,
  token: string,
  idempotencyKey: string,
): Promise<void> => {
  await mutate("DELETE", path, token, undefined, idempotencyKey);
};
