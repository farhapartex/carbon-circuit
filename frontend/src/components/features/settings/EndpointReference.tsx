import { CopyButton } from "@/components/shared/CopyButton";
import { StatusPill } from "@/components/shared/StatusPill";
import {
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import type {
  DocumentedEndpoint,
  DocumentedField,
  FieldRequirement,
} from "@/lib/apiDocs/types";
import type { StatusVariant } from "@/lib/status";

const methodVariant: Record<DocumentedEndpoint["method"], StatusVariant> = {
  GET: "info",
  POST: "success",
  PATCH: "warning",
  DELETE: "danger",
};

const requirementVariant: Record<FieldRequirement, StatusVariant> = {
  required: "danger",
  conditional: "warning",
  optional: "neutral",
};

function FieldTable({
  caption,
  fields,
}: {
  caption: string;
  fields: DocumentedField[];
}) {
  return (
    <div className="space-y-2">
      <p className="font-medium">{caption}</p>
      <div className="overflow-x-auto">
        <table className="w-full text-left text-caption">
          <thead>
            <tr className="border-b border-neutral-200 text-neutral-600">
              <th scope="col" className="py-2 pr-4 font-medium">
                Field
              </th>
              <th scope="col" className="py-2 pr-4 font-medium">
                Type
              </th>
              <th scope="col" className="py-2 pr-4 font-medium">
                Required
              </th>
              <th scope="col" className="py-2 font-medium">
                Notes
              </th>
            </tr>
          </thead>
          <tbody>
            {fields.map((field) => (
              <tr
                key={field.name}
                className="border-b border-neutral-100 align-top last:border-0"
              >
                <td className="py-2 pr-4 font-mono whitespace-nowrap">
                  {field.name}
                </td>
                <td className="py-2 pr-4 text-neutral-600">{field.type}</td>
                <td className="py-2 pr-4">
                  <StatusPill
                    presentation={{
                      label: field.requirement,
                      variant: requirementVariant[field.requirement],
                    }}
                    showDot={false}
                  />
                </td>
                <td className="py-2 text-pretty text-neutral-600">
                  {field.detail}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function Snippet({ label, body }: { label: string; body: string }) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between gap-2">
        <p className="font-medium">{label}</p>
        <CopyButton value={body} label={label} />
      </div>
      <pre className="overflow-x-auto rounded-md border border-neutral-200 bg-neutral-50 p-4 text-helper">
        <code>{body}</code>
      </pre>
    </div>
  );
}

export function EndpointReference({
  endpoint,
}: {
  endpoint: DocumentedEndpoint;
}) {
  return (
    <AccordionItem value={endpoint.id} id={endpoint.id}>
      <AccordionTrigger>
        <span className="flex flex-wrap items-center gap-3">
          <StatusPill
            presentation={{
              label: endpoint.method,
              variant: methodVariant[endpoint.method],
            }}
            showDot={false}
          />
          <code className="font-mono text-helper break-all">
            {endpoint.path}
          </code>
        </span>
        <span className="mt-1 block font-medium">{endpoint.summary}</span>
      </AccordionTrigger>

      <AccordionContent className="space-y-6">
        <p className="text-caption text-pretty text-neutral-600">
          {endpoint.detail}
        </p>

        <dl className="grid gap-3 sm:grid-cols-2">
          <div>
            <dt className="text-caption text-neutral-600">Success</dt>
            <dd className="font-medium tabular-nums">
              {endpoint.successStatus}
            </dd>
          </div>
          <div>
            <dt className="text-caption text-neutral-600">Idempotency-Key</dt>
            <dd className="font-medium">
              {endpoint.idempotencyKey ? "Required" : "Not used"}
            </dd>
          </div>
        </dl>

        {endpoint.query ? (
          <FieldTable caption="Query parameters" fields={endpoint.query} />
        ) : null}

        {endpoint.body ? (
          <FieldTable caption="Request body" fields={endpoint.body} />
        ) : null}

        {endpoint.requestExample ? (
          <Snippet label="Request" body={endpoint.requestExample} />
        ) : null}

        <Snippet label="Response" body={endpoint.responseExample} />

        <div className="space-y-2">
          <p className="font-medium">Failures</p>
          <ul className="space-y-2">
            {endpoint.failures.map((failure) => (
              <li
                key={`${failure.status}-${failure.code}-${failure.when}`}
                className="flex flex-wrap items-baseline gap-2 text-caption"
              >
                <span className="font-mono text-neutral-600 tabular-nums">
                  {failure.status}
                </span>
                <span className="font-mono">{failure.code}</span>
                <span className="text-pretty text-neutral-600">
                  {failure.when}
                </span>
              </li>
            ))}
          </ul>
        </div>
      </AccordionContent>
    </AccordionItem>
  );
}
