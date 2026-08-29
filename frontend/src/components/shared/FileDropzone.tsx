"use client";

import { UploadCloud } from "lucide-react";
import { useRef, useState, type ChangeEvent, type DragEvent } from "react";
import {
  ACCEPTED_EVIDENCE_MEDIA_TYPES,
  MAXIMUM_EVIDENCE_BYTES,
  MAXIMUM_EVIDENCE_DOCUMENTS,
} from "@/lib/types";
import { cn } from "@/lib/utils";

export type RejectedFile = {
  fileName: string;
  reason: string;
};

const MEGABYTE = 1024 * 1024;

const acceptAttribute = ACCEPTED_EVIDENCE_MEDIA_TYPES.join(",");

const describeAcceptedTypes = "PDF, PNG, JPEG, CSV, or XLSX";

type FileDropzoneProps = {
  onFilesAccepted: (files: File[]) => void;
  existingCount?: number | undefined;
  disabled?: boolean | undefined;
  className?: string | undefined;
};

export function FileDropzone({
  onFilesAccepted,
  existingCount = 0,
  disabled = false,
  className,
}: FileDropzoneProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);
  const [rejected, setRejected] = useState<RejectedFile[]>([]);

  const remainingSlots = MAXIMUM_EVIDENCE_DOCUMENTS - existingCount;

  const partition = (candidates: File[]) => {
    const accepted: File[] = [];
    const refused: RejectedFile[] = [];

    for (const file of candidates) {
      if (accepted.length >= remainingSlots) {
        refused.push({
          fileName: file.name,
          reason: `Only ${MAXIMUM_EVIDENCE_DOCUMENTS} documents are allowed per claim.`,
        });
        continue;
      }
      if (!ACCEPTED_EVIDENCE_MEDIA_TYPES.includes(file.type as never)) {
        refused.push({
          fileName: file.name,
          reason: `Must be ${describeAcceptedTypes}.`,
        });
        continue;
      }
      if (file.size > MAXIMUM_EVIDENCE_BYTES) {
        refused.push({
          fileName: file.name,
          reason: `Exceeds the ${MAXIMUM_EVIDENCE_BYTES / MEGABYTE} MB limit.`,
        });
        continue;
      }
      accepted.push(file);
    }

    setRejected(refused);
    if (accepted.length > 0) onFilesAccepted(accepted);
  };

  const onDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragging(false);
    if (disabled) return;
    partition(Array.from(event.dataTransfer.files));
  };

  const onSelect = (event: ChangeEvent<HTMLInputElement>) => {
    partition(Array.from(event.target.files ?? []));
    event.target.value = "";
  };

  return (
    <div className={cn("space-y-2", className)}>
      <div
        data-slot="file-dropzone"
        onDragOver={(event) => {
          event.preventDefault();
          if (!disabled) setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={onDrop}
        className={cn(
          "rounded-lg border border-dashed border-neutral-200 bg-white px-6 py-10 text-center transition-colors",
          dragging && "border-primary-600 bg-primary-50",
          disabled && "opacity-60",
        )}
      >
        <UploadCloud
          className="mx-auto size-6 text-muted-foreground"
          aria-hidden
        />
        <p className="mt-3 text-body">
          Drag evidence here, or{" "}
          <button
            type="button"
            disabled={disabled}
            onClick={() => inputRef.current?.click()}
            className="rounded-sm font-medium text-primary-700 underline underline-offset-4"
          >
            browse your files
          </button>
        </p>
        <p className="mt-1 text-caption text-muted-foreground">
          {describeAcceptedTypes} · up to {MAXIMUM_EVIDENCE_BYTES / MEGABYTE} MB
          each · {remainingSlots} of {MAXIMUM_EVIDENCE_DOCUMENTS} slots
          remaining
        </p>
        <input
          ref={inputRef}
          type="file"
          multiple
          accept={acceptAttribute}
          onChange={onSelect}
          disabled={disabled}
          className="sr-only"
          aria-label="Upload evidence documents"
        />
      </div>

      {rejected.length > 0 ? (
        <ul className="space-y-1" role="alert">
          {rejected.map((file) => (
            <li key={file.fileName} className="text-caption text-danger-700">
              {file.fileName} — {file.reason}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
