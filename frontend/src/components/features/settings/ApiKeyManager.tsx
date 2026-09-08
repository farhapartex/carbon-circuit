"use client";

import { KeyRound, TriangleAlert } from "lucide-react";
import { useState, useTransition } from "react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { CopyButton } from "@/components/shared/CopyButton";
import { EmptyState } from "@/components/shared/EmptyState";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { revokeApiKeyAction, submitApiKey } from "@/lib/actions/apiKeys";
import type { ApiKeyRecord } from "@/lib/api/apiKeys";

const failureMessage = (code: string) => {
  if (code === "FORBIDDEN") {
    return "Only an owner or admin can manage API keys.";
  }
  if (code === "VALIDATION_ERROR") {
    return "Give the key a name of up to 80 characters.";
  }
  if (code === "RATE_LIMITED") {
    return "You have created the maximum number of keys for today. Try again tomorrow.";
  }
  if (code === "RESOURCE_NOT_FOUND") {
    return "That key has already been revoked.";
  }
  return "We could not complete that. Please try again.";
};

export function ApiKeyManager({ keys }: { keys: ApiKeyRecord[] }) {
  const [naming, setNaming] = useState(false);
  const [name, setName] = useState("");
  const [issued, setIssued] = useState<string | null>(null);
  const [revoking, setRevoking] = useState<ApiKeyRecord | null>(null);
  const [failure, setFailure] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  const active = keys.filter((key) => key.revokedAt === null);

  const create = () => {
    setFailure(null);

    startTransition(async () => {
      try {
        const result = await submitApiKey(name.trim(), crypto.randomUUID());
        if (!result.ok) {
          setFailure(failureMessage(result.code));
          return;
        }
        setNaming(false);
        setName("");
        setIssued(result.issued.presentedKey);
      } catch (error) {
        setFailure(
          `The request could not be sent: ${error instanceof Error ? error.message : "unknown error"}`,
        );
      }
    });
  };

  const revoke = (key: ApiKeyRecord) => {
    setFailure(null);
    setRevoking(null);

    startTransition(async () => {
      try {
        const result = await revokeApiKeyAction(key.id, crypto.randomUUID());
        if (!result.ok) setFailure(failureMessage(result.code));
      } catch (error) {
        setFailure(
          `The request could not be sent: ${error instanceof Error ? error.message : "unknown error"}`,
        );
      }
    });
  };

  return (
    <>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <p className="max-w-xl text-caption text-pretty text-neutral-600">
          Keys let your ERP or warehouse system submit batches and checkpoints
          directly. Every submission carries your own external identifier, so
          replaying a day after an outage creates no duplicates. You have{" "}
          {active.length} active {active.length === 1 ? "key" : "keys"}.
        </p>
        <Button onClick={() => setNaming(true)} disabled={pending}>
          Create API key
        </Button>
      </div>

      {failure ? (
        <div
          role="alert"
          className="rounded-md border border-danger-600 bg-danger-50 px-4 py-3"
        >
          <p className="text-helper text-danger-700">{failure}</p>
        </div>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Keys</CardTitle>
        </CardHeader>
        <CardContent>
          {keys.length === 0 ? (
            <EmptyState
              icon={KeyRound}
              title="No API keys yet"
              description="Create one when you are ready to submit batches and checkpoints from your own systems rather than through this portal."
              action={
                <Button onClick={() => setNaming(true)}>Create API key</Button>
              }
            />
          ) : (
            <ul className="divide-y divide-neutral-200">
              {keys.map((key) => (
                <li
                  key={key.id}
                  className="flex flex-wrap items-center gap-3 py-4 first:pt-0 last:pb-0"
                >
                  <span className="min-w-0 flex-1">
                    <span className="flex flex-wrap items-center gap-2">
                      <span className="font-medium">{key.name}</span>
                      <StatusPill
                        presentation={
                          key.revokedAt
                            ? { label: "Revoked", variant: "neutral" }
                            : { label: "Active", variant: "success" }
                        }
                      />
                    </span>
                    <span className="flex flex-wrap items-center gap-2 text-caption text-neutral-600">
                      <span className="font-mono">cc_live_{key.prefix}…</span>
                      {key.lastUsedAt ? (
                        <>
                          {"· last used "}
                          <TimestampDisplay value={key.lastUsedAt} />
                        </>
                      ) : (
                        "· never used"
                      )}
                    </span>
                  </span>
                  {key.revokedAt ? null : (
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={pending}
                      onClick={() => setRevoking(key)}
                    >
                      Revoke
                    </Button>
                  )}
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Dialog open={naming} onOpenChange={setNaming}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create an API key</DialogTitle>
            <DialogDescription>
              Name it after the system that will use it, so you know what you
              are switching off if you ever revoke it.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-2">
            <Label htmlFor="api-key-name">Name</Label>
            <Input
              id="api-key-name"
              placeholder="ERP checkpoint ingest"
              value={name}
              onChange={(event) => setName(event.target.value)}
              maxLength={80}
            />
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setNaming(false)}>
              Cancel
            </Button>
            <Button
              disabled={name.trim().length === 0 || pending}
              onClick={create}
            >
              Create key
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={issued !== null}
        onOpenChange={(open) => {
          if (!open) setIssued(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Copy your key now</DialogTitle>
            <DialogDescription>
              This is the only time it will be shown. We store a hash of it, so
              we cannot show it to you again — if you lose it, revoke this key
              and create another.
            </DialogDescription>
          </DialogHeader>

          <div className="flex items-start gap-2 rounded-md border border-warning-600 bg-warning-50 px-4 py-3">
            <TriangleAlert
              className="mt-0.5 size-4 shrink-0 text-warning-700"
              aria-hidden
            />
            <code className="min-w-0 flex-1 font-mono text-helper break-all text-warning-700">
              {issued}
            </code>
            {issued ? <CopyButton value={issued} label="API key" /> : null}
          </div>

          <DialogFooter>
            <Button onClick={() => setIssued(null)}>I have copied it</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={revoking !== null}
        onOpenChange={(open) => {
          if (!open) setRevoking(null);
        }}
        title="Revoke this key?"
        description="Any system still using it will start being refused immediately. This cannot be undone, and the key stays on the record so you can see it was revoked."
        confirmLabel="Revoke key"
        destructive
        consequence={revoking ? <span>{revoking.name}</span> : undefined}
        onConfirm={() => {
          if (revoking) revoke(revoking);
        }}
      />
    </>
  );
}
