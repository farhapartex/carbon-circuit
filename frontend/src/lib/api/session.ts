import "server-only";
import { authorizedGet } from "@/lib/api/gateway";

type ApiSession = {
  subject: string;
  issued_at: string;
};

export type VerifiedSession = {
  subject: string;
  issuedAt: string;
};

export const fetchVerifiedSession = async (): Promise<VerifiedSession> => {
  const session = await authorizedGet<ApiSession>("/v1/session");
  return { subject: session.subject, issuedAt: session.issued_at };
};
