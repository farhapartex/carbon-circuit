"use client";

import {
  FileText,
  Loader2,
  ShieldCheck,
  Trash2,
  TriangleAlert,
} from "lucide-react";
import { useState, useTransition } from "react";
import { FileDropzone } from "@/components/shared/FileDropzone";
import { Button } from "@/components/ui/button";
import { uploadEvidenceDocument } from "@/lib/actions/evidence";
import type { EvidenceDocument } from "@/lib/api/evidence";
import {
  ACCEPTED_EVIDENCE_MEDIA_TYPES,
  MAXIMUM_EVIDENCE_BYTES,
  MAXIMUM_EVIDENCE_DOCUMENTS,
  MAXIMUM_EVIDENCE_PAGES,
} from "@/lib/types";

const MEGABYTE = 1024 * 1024;

const MEDIA_TYPE_LABELS: Record<string, string> = {
  "application/pdf": "PDF",
  "image/png": "PNG",
  "image/jpeg": "JPEG",
  "text/csv": "CSV",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "XLSX",
};

const formatSize = (bytes: number) => {
  if (bytes < MEGABYTE) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  return `${(bytes / MEGABYTE).toFixed(1)} MB`;
};

type Refusal = {
  fileName: string;
  message: string;
};

type EvidenceUploadStepProps = {
  documents: EvidenceDocument[];
  onDocumentsChange: (documents: EvidenceDocument[]) => void;
  onBusyChange: (busy: boolean) => void;
};

export function EvidenceUploadStep({
  documents,
  onDocumentsChange,
  onBusyChange,
}: EvidenceUploadStepProps) {
  const [pending, startTransition] = useTransition();
  const [uploading, setUploading] = useState<string[]>([]);
  const [refusals, setRefusals] = useState<Refusal[]>([]);

  const formats = ACCEPTED_EVIDENCE_MEDIA_TYPES.map(
    (type) => MEDIA_TYPE_LABELS[type] ?? type,
  ).join(", ");

  const send = (files: File[]) => {
    setRefusals([]);
    setUploading(files.map((file) => file.name));
    onBusyChange(true);

    startTransition(async () => {
      const accepted: EvidenceDocument[] = [];
      const refused: Refusal[] = [];

      for (const file of files) {
        const form = new FormData();
        form.append("purpose", "claim_evidence");
        form.append("file", file, file.name);
        form.append("idempotencyKey", crypto.randomUUID());

        try {
          const result = await uploadEvidenceDocument(form);
          if (result.ok) {
            accepted.push(result.document);
          } else {
            refused.push({ fileName: file.name, message: result.message });
          }
        } catch {
          refused.push({
            fileName: file.name,
            message: "The upload could not be sent. Check your connection.",
          });
        }

        setUploading((names) => names.filter((name) => name !== file.name));
      }

      if (accepted.length > 0) {
        onDocumentsChange([...documents, ...accepted]);
      }
      setRefusals(refused);
      onBusyChange(false);
    });
  };

  const remove = (documentId: string) => {
    onDocumentsChange(
      documents.filter((document) => document.id !== documentId),
    );
  };

  return (
    <div className="space-y-4">
      <div>
        <p className="font-medium">Supporting evidence</p>
        <p className="mt-1 text-caption text-pretty text-neutral-600">
          Utility bills, third-party audit reports, meter or sensor exports, and
          supplier certificates — whatever independently corroborates the
          figures you entered. Each file is scanned and hashed before it is
          stored privately; the hash is what makes it auditable later.
        </p>
      </div>

      <FileDropzone
        onFilesAccepted={send}
        existingCount={documents.length}
        disabled={pending || documents.length >= MAXIMUM_EVIDENCE_DOCUMENTS}
      />

      {uploading.length > 0 ? (
        <ul className="space-y-1" aria-live="polite">
          {uploading.map((fileName) => (
            <li
              key={fileName}
              className="flex items-center gap-2 text-caption text-neutral-600"
            >
              <Loader2 className="size-3.5 shrink-0 animate-spin" aria-hidden />
              Scanning {fileName}…
            </li>
          ))}
        </ul>
      ) : null}

      {refusals.length > 0 ? (
        <ul
          className="space-y-2 rounded-md border border-danger-600 bg-danger-50 px-4 py-3"
          role="alert"
        >
          {refusals.map((refusal) => (
            <li key={refusal.fileName} className="text-caption text-danger-700">
              <span className="font-medium">{refusal.fileName}</span> —{" "}
              {refusal.message}
            </li>
          ))}
        </ul>
      ) : null}

      {documents.length > 0 ? (
        <ul className="divide-y divide-neutral-200 rounded-lg border border-neutral-200 bg-white">
          {documents.map((document) => (
            <li
              key={document.id}
              className="flex flex-wrap items-center gap-3 px-4 py-3"
            >
              <FileText
                className="size-4 shrink-0 text-neutral-600"
                aria-hidden
              />
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">{document.fileName}</p>
                <p className="text-caption text-neutral-600">
                  {MEDIA_TYPE_LABELS[document.mediaType] ?? document.mediaType}
                  {" · "}
                  {formatSize(document.byteSize)}
                  {document.pageCount === null
                    ? null
                    : ` · ${document.pageCount} ${document.pageCount === 1 ? "page" : "pages"}`}
                </p>
                {document.activeContentStripped ? (
                  <p className="mt-1 flex items-center gap-1.5 text-caption text-warning-700">
                    <TriangleAlert className="size-3.5 shrink-0" aria-hidden />
                    Embedded active content was removed before storing.
                  </p>
                ) : null}
              </div>
              <span className="flex items-center gap-1.5 text-caption text-success-700">
                <ShieldCheck className="size-3.5 shrink-0" aria-hidden />
                Scanned clean
              </span>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => remove(document.id)}
                aria-label={`Remove ${document.fileName} from this claim`}
              >
                <Trash2 className="size-4" aria-hidden />
              </Button>
            </li>
          ))}
        </ul>
      ) : null}

      <dl className="space-y-2 rounded-lg border border-neutral-200 bg-white px-4 py-4">
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">Documents attached</dt>
          <dd className="font-medium tabular-nums">
            {documents.length} of {MAXIMUM_EVIDENCE_DOCUMENTS}
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">Size per file</dt>
          <dd className="font-medium tabular-nums">
            up to {MAXIMUM_EVIDENCE_BYTES / MEGABYTE} MB
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">Pages per document</dt>
          <dd className="font-medium tabular-nums">
            up to {MAXIMUM_EVIDENCE_PAGES}
          </dd>
        </div>
        <div className="flex flex-wrap items-baseline justify-between gap-4">
          <dt className="text-caption text-neutral-600">Accepted formats</dt>
          <dd className="font-medium">{formats}</dd>
        </div>
      </dl>
    </div>
  );
}
