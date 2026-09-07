import "server-only";
import { gatewayDelete, gatewayGet } from "@/lib/api/gateway";

type ApiSession = {
  id: string;
  user_agent: string;
  ip_address: string;
  started_at: string;
  last_seen_at: string;
  current: boolean;
};

export type SessionRecord = {
  id: string;
  userAgent: string;
  ipAddress: string;
  startedAt: string;
  lastSeenAt: string;
  current: boolean;
};

const toSession = (session: ApiSession): SessionRecord => ({
  id: session.id,
  userAgent: session.user_agent,
  ipAddress: session.ip_address,
  startedAt: session.started_at,
  lastSeenAt: session.last_seen_at,
  current: session.current,
});

export const fetchSessions = async (
  token: string,
): Promise<SessionRecord[]> => {
  const listed = await gatewayGet<{ sessions: ApiSession[] }>(
    "/v1/sessions",
    token,
  );
  return listed.sessions.map(toSession);
};

export const revokeSession = async (
  token: string,
  sessionId: string,
  idempotencyKey: string,
): Promise<void> => {
  await gatewayDelete(
    `/v1/sessions/${encodeURIComponent(sessionId)}`,
    token,
    idempotencyKey,
  );
};
