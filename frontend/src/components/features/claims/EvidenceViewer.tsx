import {
  FileText,
  ShieldAlert,
  ShieldCheck,
  ShieldQuestion,
} from "lucide-react";
import { CopyButton } from "@/components/shared/CopyButton";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { Evidence, EvidenceScanStatus } from "@/lib/types";

const scanPresentation: Record<
  EvidenceScanStatus,
  { icon: typeof ShieldCheck; label: string; className: string }
> = {
  clean: {
    icon: ShieldCheck,
    label: "Scanned clean",
    className: "text-success-700",
  },
  pending: {
    icon: ShieldQuestion,
    label: "Scan pending",
    className: "text-warning-700",
  },
  failed: {
    icon: ShieldAlert,
    label: "Scan failed",
    className: "text-danger-700",
  },
};

const byteFormat = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 1,
});

const readableSize = (bytes: number) =>
  bytes >= 1024 * 1024
    ? `${byteFormat.format(bytes / (1024 * 1024))} MB`
    : `${byteFormat.format(bytes / 1024)} KB`;

export function EvidenceViewer({ evidence }: { evidence: Evidence[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Evidence</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {evidence.length === 0 ? (
          <p className="text-caption text-neutral-600">
            No documents were attached to this claim.
          </p>
        ) : (
          evidence.map((document) => {
            const scan = scanPresentation[document.scanStatus];
            const ScanIcon = scan.icon;

            return (
              <div
                key={document.id}
                className="rounded-md border border-neutral-200 px-4 py-3"
              >
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <span className="flex min-w-0 items-start gap-2">
                    <FileText
                      className="mt-0.5 size-4 shrink-0 text-neutral-600"
                      aria-hidden
                    />
                    <span className="min-w-0">
                      <span className="block truncate font-medium">
                        {document.fileName}
                      </span>
                      <span className="block text-caption text-neutral-600">
                        {readableSize(document.byteSize)}
                        {document.pageCount === null
                          ? ""
                          : ` · ${document.pageCount} pages`}{" "}
                        · uploaded{" "}
                        <TimestampDisplay
                          value={document.uploadedAt}
                          dateOnly
                        />
                      </span>
                    </span>
                  </span>
                  <span
                    className={`flex shrink-0 items-center gap-1.5 text-caption ${scan.className}`}
                  >
                    <ScanIcon className="size-3.5" aria-hidden />
                    {scan.label}
                  </span>
                </div>

                <div className="mt-3 flex items-center gap-2 border-t border-neutral-200 pt-3">
                  <span className="text-caption text-neutral-600">Hash</span>
                  <code className="min-w-0 truncate text-helper">
                    {document.contentHash}
                  </code>
                  <CopyButton
                    value={document.contentHash}
                    label="Copy content hash"
                  />
                </div>
              </div>
            );
          })
        )}

        <p className="text-caption text-pretty text-neutral-600">
          Documents stay in private storage. Only their content hashes are ever
          published, and the files themselves are fetched through the server
          rather than linked directly from this page.
        </p>
      </CardContent>
    </Card>
  );
}
