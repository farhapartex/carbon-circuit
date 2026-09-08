import { randomUUID } from "node:crypto";
import { NextResponse } from "next/server";
import { createEvidenceDownload } from "@/lib/api/evidence";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";

type RouteContext = {
  params: Promise<{ documentId: string }>;
};

const refuse = (status: number, reason: string) =>
  NextResponse.json({ error: reason }, { status });

export async function GET(_request: Request, context: RouteContext) {
  const session = await auth0.getSession();
  if (!session) {
    return refuse(401, "Authentication is required.");
  }

  const { documentId } = await context.params;

  try {
    const { token } = await auth0.getAccessToken();
    const link = await createEvidenceDownload(token, documentId, randomUUID());

    const stored = await fetch(link.url, { cache: "no-store" });
    if (!stored.ok || !stored.body) {
      return refuse(502, "The document could not be retrieved from storage.");
    }

    return new NextResponse(stored.body, {
      headers: {
        "Content-Type": link.mediaType,
        "Content-Disposition": `attachment; filename="${link.fileName.replaceAll('"', "")}"`,
        "Cache-Control": "no-store",
        "X-Content-Type-Options": "nosniff",
      },
    });
  } catch (error) {
    if (error instanceof GatewayError) {
      if (error.status === 404) {
        return refuse(404, "No such document.");
      }
      if (error.code === "EVIDENCE_NOT_READY") {
        return refuse(409, "This document has not passed scanning.");
      }
      return refuse(502, "The evidence service could not issue a download.");
    }
    throw error;
  }
}
