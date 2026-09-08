import type { Metadata } from "next";
import Link from "next/link";
import { EndpointReference } from "@/components/features/settings/EndpointReference";
import { CopyButton } from "@/components/shared/CopyButton";
import { Accordion } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { apiDocGroups } from "@/lib/apiDocs/endpoints";

export const metadata: Metadata = { title: "API documentation" };

const AUTH_EXAMPLE = `curl https://api.carboncircuit.dev/v1/batches \\
  -H "Authorization: Bearer cc_live_a1b2c3d4_<secret>"`;

const ERROR_EXAMPLE = `{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "One or more fields are invalid.",
    "request_id": "7e0d7715-4c91-4c47-bb80-194289b9f13c",
    "details": [
      { "field": "external_id", "code": "REQUIRED_FOR_API_SUBMISSION" }
    ]
  }
}`;

export default function SettingsApiDocsPage() {
  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Submitting data from your own systems</CardTitle>
        </CardHeader>
        <CardContent className="space-y-5">
          <p className="text-caption text-pretty text-neutral-600">
            These are the endpoints an API key can reach. They are the same
            endpoints this portal uses, so a batch your ERP submits is
            indistinguishable from one entered here by hand — only the
            credential differs.
          </p>

          <div className="space-y-2">
            <p className="font-medium">Authentication</p>
            <p className="text-caption text-pretty text-neutral-600">
              Send your key as a bearer token. The key is shown once when you
              create it and is stored only as a hash, so it cannot be recovered
              — if you lose it, revoke it and issue another.
            </p>
            <div className="flex items-start justify-between gap-2 rounded-md border border-neutral-200 bg-neutral-50 p-4">
              <pre className="overflow-x-auto text-helper">
                <code>{AUTH_EXAMPLE}</code>
              </pre>
              <CopyButton value={AUTH_EXAMPLE} label="example request" />
            </div>
            <Button asChild variant="outline" size="sm">
              <Link href="/settings/api-keys">Manage API keys</Link>
            </Button>
          </div>

          <div className="space-y-2 border-t border-neutral-200 pt-5">
            <p className="font-medium">Rules that apply to every request</p>
            <ul className="space-y-2 text-caption text-pretty text-neutral-600">
              <li>
                <span className="font-medium text-neutral-900">
                  Successful responses are wrapped in{" "}
                  <code className="font-mono">data</code>
                </span>{" "}
                and failures in <code className="font-mono">error</code>. Every
                failure carries a <code className="font-mono">request_id</code>{" "}
                — quote it if you ask us about a specific call.
              </li>
              <li>
                <span className="font-medium text-neutral-900">
                  Writes need an{" "}
                  <code className="font-mono">Idempotency-Key</code> header
                </span>{" "}
                of 8 to 255 printable characters. Retrying with the same key
                returns the original outcome instead of acting twice.
              </li>
              <li>
                <span className="font-medium text-neutral-900">
                  Every record you submit needs an{" "}
                  <code className="font-mono">external_id</code>
                </span>{" "}
                — its identifier in your system. That is what makes replaying a
                day of events after an outage safe, and it gives you a stable
                key to reconcile against.
              </li>
              <li>
                <span className="font-medium text-neutral-900">
                  Facilities are never created implicitly.
                </span>{" "}
                A submission naming a facility you have not registered is
                refused rather than guessed at.
              </li>
              <li>
                <span className="font-medium text-neutral-900">
                  Decimal amounts are strings.
                </span>{" "}
                Quantities travel as{" "}
                <code className="font-mono">&quot;5000.000000&quot;</code>{" "}
                rather than numbers so nothing is lost to floating point.
              </li>
            </ul>
          </div>

          <div className="space-y-2 border-t border-neutral-200 pt-5">
            <div className="flex items-center justify-between gap-2">
              <p className="font-medium">A failure looks like this</p>
              <CopyButton value={ERROR_EXAMPLE} label="error example" />
            </div>
            <pre className="overflow-x-auto rounded-md border border-neutral-200 bg-neutral-50 p-4 text-helper">
              <code>{ERROR_EXAMPLE}</code>
            </pre>
          </div>
        </CardContent>
      </Card>

      {apiDocGroups.map((group) => (
        <section key={group.id} className="space-y-4">
          <div className="space-y-1">
            <h2 className="text-section-heading">{group.title}</h2>
            <p className="max-w-2xl text-caption text-pretty text-neutral-600">
              {group.detail}
            </p>
          </div>
          <Accordion type="multiple">
            {group.endpoints.map((endpoint) => (
              <EndpointReference key={endpoint.id} endpoint={endpoint} />
            ))}
          </Accordion>
        </section>
      ))}

      <p className="max-w-2xl text-caption text-pretty text-neutral-600">
        Bulk submission is not available yet. When it arrives it will accept
        many checkpoints in one request and answer with a per-item result, so
        one malformed row cannot reject the rest of the batch.
      </p>
    </>
  );
}
