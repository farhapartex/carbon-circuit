import { CheckCircle2, Info, ShieldX } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { RegistryRejection } from "@/lib/api/organizations";
import type { VerificationStatus } from "@/lib/status";

const NAME_SIMILARITY_THRESHOLD = 0.85;

const GATED_CAPABILITIES = [
  "Submit sustainability claims",
  "Receive credit issuance",
  "List credits for sale",
];

const REJECTION_EXPLANATIONS: Record<RegistryRejection, string> = {
  entity_dissolved:
    "The registry lists this entity as dissolved. We cannot issue credits against a company that no longer legally exists.",
  sanctions_flag:
    "The registry carries a sanctions or restricted-party flag against this entity.",
  name_mismatch:
    "The name you entered does not closely enough match the registered legal name.",
};

type VerificationOutcomeProps = {
  declaredName: string;
  countryCode: string;
  registrationNumber: string;
  status: VerificationStatus;
  rejection: RegistryRejection | null;
  nameSimilarity: number | null;
  registeredLegalName: string | null;
};

export function VerificationOutcome({
  declaredName,
  countryCode,
  registrationNumber,
  status,
  rejection,
  nameSimilarity,
  registeredLegalName,
}: VerificationOutcomeProps) {
  if (status === "verified") {
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
          {registeredLegalName ? (
            <dl>
              <dt className="text-caption text-success-700/80">
                Registered legal name
              </dt>
              <dd className="font-medium text-success-700">
                {registeredLegalName}
              </dd>
            </dl>
          ) : null}
          <Button asChild size="lg">
            <Link href="/onboarding/plan">Choose a plan</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (status === "unverified") {
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
            halves the credit ceiling any claim can reach. To resolve it,
            contact support with your registration documents and we will verify
            the organization manually.
          </p>
          <Button asChild size="lg">
            <Link href="/onboarding/plan">Continue to plan selection</Link>
          </Button>
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
        {rejection ? (
          <p className="text-caption text-pretty text-danger-700">
            {REJECTION_EXPLANATIONS[rejection]}
          </p>
        ) : null}

        {rejection === "name_mismatch" ? (
          <dl className="space-y-2">
            <div>
              <dt className="text-caption text-danger-700/80">You entered</dt>
              <dd className="font-medium text-danger-700">{declaredName}</dd>
            </div>
            {registeredLegalName ? (
              <div>
                <dt className="text-caption text-danger-700/80">
                  The register says
                </dt>
                <dd className="font-medium text-danger-700">
                  {registeredLegalName}
                </dd>
              </div>
            ) : null}
            {nameSimilarity !== null ? (
              <div>
                <dt className="text-caption text-danger-700/80">Similarity</dt>
                <dd className="font-medium text-danger-700 tabular-nums">
                  {Math.round(nameSimilarity * 100)}%, below the{" "}
                  {Math.round(NAME_SIMILARITY_THRESHOLD * 100)}% threshold
                </dd>
              </div>
            ) : null}
          </dl>
        ) : null}

        <p className="text-caption text-danger-700">
          Your account exists but no platform actions are available. Our team
          reviews suspended registrations manually and will be in touch.
        </p>
      </CardContent>
    </Card>
  );
}
