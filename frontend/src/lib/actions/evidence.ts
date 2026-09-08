"use server";

import {
  uploadEvidence,
  type EvidenceDocument,
  type EvidencePurpose,
} from "@/lib/api/evidence";
import { GatewayError } from "@/lib/api/gateway";
import { auth0 } from "@/lib/auth0";
import { MAXIMUM_EVIDENCE_BYTES } from "@/lib/types";

export type EvidenceUploadResult =
  { ok: true; document: EvidenceDocument } | { ok: false; message: string };

const refusalMessages: Record<string, string> = {
  REJECTED_MEDIA_TYPE: "That file type is not accepted as evidence.",
  REJECTED_CONTENT_MISMATCH:
    "The file's contents do not match the type it claims to be.",
  REJECTED_TOO_LARGE: `The file is larger than the ${MAXIMUM_EVIDENCE_BYTES / (1024 * 1024)} MB limit.`,
  REJECTED_PAGE_COUNT: "The document has more pages than a claim allows.",
  REJECTED_ACTIVE_CONTENT:
    "The file carries embedded scripts or macros that could not be removed.",
  REJECTED_MALWARE: "The file was refused by the malware scanner.",
  REJECTED_UNREADABLE: "The file could not be read as its declared format.",
};

const codeMessages: Record<string, string> = {
  PAYLOAD_TOO_LARGE: `The file is larger than the ${MAXIMUM_EVIDENCE_BYTES / (1024 * 1024)} MB limit.`,
  RATE_LIMITED: "Too many uploads just now. Wait a moment and try again.",
  ORGANIZATION_READ_ONLY:
    "This organization cannot upload evidence in its current state.",
  IDEMPOTENCY_KEY_REUSED:
    "A different file was already uploaded under this attempt. Retry the upload.",
  REQUEST_IN_PROGRESS: "This file is still being processed. Try again shortly.",
  DEPENDENCY_UNAVAILABLE:
    "The evidence service is unreachable right now. Try again shortly.",
};

const describe = (error: GatewayError): string => {
  const verdict = error.codeForField("file");
  if (verdict && refusalMessages[verdict]) {
    return refusalMessages[verdict];
  }
  return codeMessages[error.code] ?? "The file could not be uploaded.";
};

export const uploadEvidenceDocument = async (
  form: FormData,
): Promise<EvidenceUploadResult> => {
  const file = form.get("file");
  const purpose = form.get("purpose");
  const idempotencyKey = form.get("idempotencyKey");

  if (!(file instanceof File) || file.size === 0) {
    return { ok: false, message: "No file was received." };
  }
  if (typeof purpose !== "string" || typeof idempotencyKey !== "string") {
    return { ok: false, message: "The upload was missing its metadata." };
  }

  try {
    const { token } = await auth0.getAccessToken();
    const document = await uploadEvidence(
      token,
      purpose as EvidencePurpose,
      file,
      idempotencyKey,
    );

    return { ok: true, document };
  } catch (error) {
    if (error instanceof GatewayError) {
      return { ok: false, message: describe(error) };
    }
    throw error;
  }
};
