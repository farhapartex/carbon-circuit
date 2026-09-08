import "server-only";
import { gatewayGet, gatewayPost, gatewayUpload } from "@/lib/api/gateway";
import type { Evidence, EvidenceScanStatus } from "@/lib/types";

export type EvidencePurpose = "claim_evidence" | "batch_certification";

export type EvidenceVerdict =
  | "accepted"
  | "rejected_media_type"
  | "rejected_content_mismatch"
  | "rejected_too_large"
  | "rejected_page_count"
  | "rejected_active_content"
  | "rejected_malware"
  | "rejected_unreadable";

type ApiDocument = {
  id: string;
  purpose: EvidencePurpose;
  file_name: string;
  media_type: string;
  byte_size: number;
  page_count: number | null;
  content_hash: string;
  scan_status: EvidenceScanStatus;
  scan_verdict: EvidenceVerdict;
  scanned_by: string;
  scanned_at: string;
  active_content_stripped: boolean;
  created_at: string;
};

export type EvidenceDocument = Evidence & {
  purpose: EvidencePurpose;
  scanVerdict: EvidenceVerdict;
  scannedBy: string;
  activeContentStripped: boolean;
};

type ApiDownloadLink = {
  url: string;
  expires_at: string;
  file_name: string;
  media_type: string;
};

export type EvidenceDownload = {
  url: string;
  expiresAt: string;
  fileName: string;
  mediaType: string;
};

const toEvidenceDocument = (document: ApiDocument): EvidenceDocument => ({
  id: document.id,
  purpose: document.purpose,
  fileName: document.file_name,
  mediaType: document.media_type,
  byteSize: document.byte_size,
  pageCount: document.page_count,
  contentHash: document.content_hash,
  scanStatus: document.scan_status,
  scanVerdict: document.scan_verdict,
  scannedBy: document.scanned_by,
  activeContentStripped: document.active_content_stripped,
  uploadedAt: document.created_at,
});

export const uploadEvidence = async (
  token: string,
  purpose: EvidencePurpose,
  file: File,
  idempotencyKey: string,
): Promise<EvidenceDocument> => {
  const form = new FormData();
  form.append("purpose", purpose);
  form.append("file", file, file.name);

  const uploaded = await gatewayUpload<ApiDocument>(
    "/v1/evidence",
    token,
    form,
    idempotencyKey,
  );

  return toEvidenceDocument(uploaded);
};

export const fetchEvidenceDocuments = async (
  token: string,
  purpose: EvidencePurpose,
): Promise<EvidenceDocument[]> => {
  const listed = await gatewayGet<{ documents: ApiDocument[] }>(
    `/v1/evidence?purpose=${purpose}`,
    token,
  );

  return listed.documents.map(toEvidenceDocument);
};

export const fetchEvidenceDocument = async (
  token: string,
  documentId: string,
): Promise<EvidenceDocument> => {
  const found = await gatewayGet<ApiDocument>(
    `/v1/evidence/${documentId}`,
    token,
  );

  return toEvidenceDocument(found);
};

export const createEvidenceDownload = async (
  token: string,
  documentId: string,
  idempotencyKey: string,
): Promise<EvidenceDownload> => {
  const link = await gatewayPost<ApiDownloadLink>(
    `/v1/evidence/${documentId}/download`,
    token,
    {},
    idempotencyKey,
  );

  return {
    url: link.url,
    expiresAt: link.expires_at,
    fileName: link.file_name,
    mediaType: link.media_type,
  };
};
