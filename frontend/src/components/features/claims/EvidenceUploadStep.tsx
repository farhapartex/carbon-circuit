import { FileWarning } from "lucide-react";
import {
  ACCEPTED_EVIDENCE_MEDIA_TYPES,
  MAXIMUM_EVIDENCE_BYTES,
  MAXIMUM_EVIDENCE_DOCUMENTS,
  MAXIMUM_EVIDENCE_PAGES,
} from "@/lib/types";

const MEDIA_TYPE_LABELS: Record<string, string> = {
  "application/pdf": "PDF",
  "image/png": "PNG",
  "image/jpeg": "JPEG",
  "text/csv": "CSV",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "XLSX",
};

export function EvidenceUploadStep() {
  const megabytes = MAXIMUM_EVIDENCE_BYTES / (1024 * 1024);
  const formats = ACCEPTED_EVIDENCE_MEDIA_TYPES.map(
    (type) => MEDIA_TYPE_LABELS[type] ?? type,
  ).join(", ");

  return (
    <div className="space-y-4">
      <div className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3">
        <p className="flex items-center gap-2 font-medium text-warning-700">
          <FileWarning className="size-4 shrink-0" aria-hidden />
          Evidence upload is not available yet
        </p>
        <p className="mt-1 text-caption text-pretty text-warning-700">
          Documents are scanned, hashed, and stored privately by the evidence
          service, which is not built. Rather than take files this page cannot
          safely keep, uploading is disabled until it exists — so a claim cannot
          actually be submitted yet.
        </p>
      </div>

      <div className="rounded-lg border border-neutral-200 bg-white px-4 py-4">
        <p className="font-medium">What you will attach here</p>
        <p className="mt-1 text-caption text-pretty text-neutral-600">
          Utility bills, third-party audit reports, meter or sensor exports, and
          supplier certificates — whatever independently corroborates the
          figures you entered.
        </p>
        <dl className="mt-4 space-y-2 border-t border-neutral-200 pt-4">
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Documents</dt>
            <dd className="font-medium tabular-nums">
              up to {MAXIMUM_EVIDENCE_DOCUMENTS}
            </dd>
          </div>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Size per file</dt>
            <dd className="font-medium tabular-nums">up to {megabytes} MB</dd>
          </div>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">
              Pages per document
            </dt>
            <dd className="font-medium tabular-nums">
              up to {MAXIMUM_EVIDENCE_PAGES}
            </dd>
          </div>
          <div className="flex flex-wrap items-baseline justify-between gap-4">
            <dt className="text-caption text-neutral-600">Formats</dt>
            <dd className="font-medium">{formats}</dd>
          </div>
        </dl>
        <p className="mt-3 text-caption text-pretty text-neutral-600">
          These limits exist because evidence volume drives the cost of AI
          review directly.
        </p>
      </div>
    </div>
  );
}
