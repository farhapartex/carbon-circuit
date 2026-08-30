"use client";

import { CheckCircle2, Info, ShieldX } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { NAME_SIMILARITY_THRESHOLD, verifyRegistration } from "@/lib/fixtures";
import type {
  RegistryRejectionReason,
  RegistryVerificationOutcome,
} from "@/lib/types";
import { useFormDraftStore } from "@/stores/form-drafts";

const GATED_CAPABILITIES = [
  "Submit sustainability claims",
  "Receive credit issuance",
  "List credits for sale",
];

const REJECTION_EXPLANATIONS: Record<RegistryRejectionReason, string> = {
  entity_dissolved:
    "The registry lists this entity as dissolved. We cannot issue credits against a company that no longer legally exists.",
  sanctions_flag:
    "The registry carries a sanctions or restricted-party flag against this entity.",
  name_mismatch:
    "The name you entered does not closely enough match the registered legal name.",
};

export function VerificationOutcome() {
  const draft = useFormDraftStore((state) => state.drafts.organization);
  const [outcome, setOutcome] = useState<RegistryVerificationOutcome | null>(
    null,
  );

  const declaredName = String(draft?.values.name ?? "");
  const countryCode = String(draft?.values.countryOfIncorporation ?? "");
  const registrationNumber = String(
    draft?.values.businessRegistrationNumber ?? "",
  );

  useEffect(() => {
    if (!registrationNumber) return;
    let cancelled = false;
    verifyRegistration(countryCode, registrationNumber, declaredName).then(
      (result) => {
        if (!cancelled) setOutcome(result);
      },
    );
    return () => {
      cancelled = true;
    };
  }, [countryCode, registrationNumber, declaredName]);

  if (!draft) {
    return (
      <Card>
        <CardContent className="space-y-4">
          <p>We do not have your organization details yet.</p>
          <Button asChild>
            <Link href="/onboarding/organization">Start again</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (!outcome) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Checking the business registry…</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
        </CardContent>
      </Card>
    );
  }

  if (outcome.status === "verified") {
    return (
      <Card className="border-success-600/30 bg-success-50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-success-700">
            <CheckCircle2 className="size-5" aria-hidden />
            Verified
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-caption text-pretty text-success-700">
            {registrationNumber} matched an active entity in the {countryCode}{" "}
            register. Every capability your plan includes is available to you.
          </p>
          <dl className="grid gap-3 sm:grid-cols-2">
            <div>
              <dt className="text-caption text-success-700/80">
                Registered legal name
              </dt>
              <dd className="font-medium text-success-700">
                {outcome.matchedRecord?.legalName}
              </dd>
            </div>
            <div>
              <dt className="text-caption text-success-700/80">
                Registered address
              </dt>
              <dd className="font-medium text-success-700">
                {outcome.matchedRecord?.registeredAddress}
              </dd>
            </div>
          </dl>
          <Button asChild size="lg">
            <Link href="/onboarding/plan">Choose a plan</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (outcome.status === "unverified") {
    return (
      <Card className="border-warning-600/30 bg-warning-50">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-warning-700">
            <Info className="size-5" aria-hidden />
            Not matched
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-caption text-pretty text-warning-700">
            We could not find {registrationNumber} in the {countryCode}{" "}
            register. You can carry on and use the product, but three
            capabilities stay gated until this is resolved.
          </p>
          <ul className="space-y-1">
            {GATED_CAPABILITIES.map((capability) => (
              <li key={capability} className="text-caption text-warning-700">
                · {capability}
              </li>
            ))}
          </ul>
          <p className="text-caption text-warning-700">
            Facilities you register will also be treated as self-declared, which
            halves the credit ceiling any claim can reach.
          </p>
          <div className="flex flex-wrap gap-2">
            <Button asChild size="lg">
              <Link href="/onboarding/plan">Continue anyway</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link href="/onboarding/organization">Correct the number</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-danger-600/30 bg-danger-50">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-danger-700">
          <ShieldX className="size-5" aria-hidden />
          Account suspended pending review
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-caption text-pretty text-danger-700">
          {outcome.rejectionReason
            ? REJECTION_EXPLANATIONS[outcome.rejectionReason]
            : null}
        </p>
        {outcome.rejectionReason === "name_mismatch" ? (
          <dl className="space-y-2">
            <div>
              <dt className="text-caption text-danger-700/80">You entered</dt>
              <dd className="font-medium text-danger-700">{declaredName}</dd>
            </div>
            <div>
              <dt className="text-caption text-danger-700/80">
                The register says
              </dt>
              <dd className="font-medium text-danger-700">
                {outcome.matchedRecord?.legalName}
              </dd>
            </div>
            <div>
              <dt className="text-caption text-danger-700/80">Similarity</dt>
              <dd className="font-medium text-danger-700 tabular-nums">
                {Math.round((outcome.nameSimilarity ?? 0) * 100)}%, below the{" "}
                {Math.round(NAME_SIMILARITY_THRESHOLD * 100)}% threshold
              </dd>
            </div>
          </dl>
        ) : null}
        <p className="text-caption text-danger-700">
          Your account exists but no platform actions are available. Our team
          reviews suspended registrations manually and will be in touch.
        </p>
        <div className="flex flex-wrap gap-2">
          <Button size="lg" variant="outline">
            Contact support
          </Button>
          <Button asChild size="lg" variant="ghost">
            <Link href="/onboarding/organization">Review what you entered</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
