import type { Metadata } from "next";
import { ShieldAlert } from "lucide-react";
import { VerificationStatusBadge } from "@/components/shared/StatusBadges";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { fetchCurrentOrganization } from "@/lib/api/organization";
import { auth0 } from "@/lib/auth0";
import { countryName } from "@/lib/countries";
import {
  organizationStateLabels,
  organizationTypeLabels,
  productCategoryLabels,
  registryRejectionExplanations,
} from "@/lib/labels";
import type { ProductCategory } from "@/lib/types";

export const metadata: Metadata = { title: "Organization" };

const GATED_CAPABILITIES = [
  "Submit sustainability claims",
  "Receive credit issuance",
  "List credits for sale",
];

const percentOf = (similarity: number) => `${Math.round(similarity * 100)}%`;

export default async function SettingsOrganizationPage() {
  const { token } = await auth0.getAccessToken();
  const organization = await fetchCurrentOrganization(token);

  const verified = organization.verificationStatus === "verified";
  const { outcome } = organization;

  return (
    <>
      {organization.state === "active" ? null : (
        <div
          role="status"
          className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3"
        >
          <p className="flex items-center gap-2 font-medium text-warning-700">
            <ShieldAlert className="size-4 shrink-0" aria-hidden />
            This organization is{" "}
            {organizationStateLabels[organization.state].toLowerCase()}
          </p>
          <p className="mt-1 text-caption text-pretty text-warning-700">
            Some actions are unavailable while the account is in this state.
          </p>
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Organization details</CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-caption text-neutral-600">Legal name</dt>
              <dd className="font-medium">{organization.name}</dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">Type</dt>
              <dd className="font-medium">
                {organizationTypeLabels[organization.type]}
              </dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">
                Country of incorporation
              </dt>
              <dd className="font-medium">
                {countryName(organization.countryOfIncorporation)}
              </dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">
                Business registration number
              </dt>
              <dd className="font-medium tabular-nums">
                {organization.businessRegistrationNumber}
              </dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">Your role</dt>
              <dd className="font-medium">{organization.role}</dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">Registered</dt>
              <dd className="font-medium">
                <TimestampDisplay value={organization.createdAt} dateOnly />
              </dd>
            </div>
          </dl>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <CardTitle>Verification</CardTitle>
          <VerificationStatusBadge status={organization.verificationStatus} />
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-caption text-pretty text-neutral-600">
            {verified
              ? "Your registration number matched an active entity in the business registry, so every capability your plan includes is available."
              : outcome.registryMatchFound
                ? "Your registration number matched a registry entry, but the match was not accepted."
                : "Your registration number did not match any entity in the business registry. You can use the product, but three capabilities are gated until that is resolved."}
          </p>

          {outcome.rejection ? (
            <p className="rounded-md border border-danger-600 bg-danger-50 px-4 py-3 text-caption text-pretty text-danger-700">
              {registryRejectionExplanations[outcome.rejection]}
            </p>
          ) : null}

          <dl className="space-y-3 border-t border-neutral-200 pt-4">
            <div className="flex flex-wrap items-baseline justify-between gap-4">
              <dt className="text-caption text-neutral-600">Registry match</dt>
              <dd className="font-medium">
                {outcome.registryMatchFound ? "Found" : "No match"}
              </dd>
            </div>
            {outcome.registeredLegalName ? (
              <div className="flex flex-wrap items-baseline justify-between gap-4">
                <dt className="text-caption text-neutral-600">
                  Registered legal name
                </dt>
                <dd className="font-medium">{outcome.registeredLegalName}</dd>
              </div>
            ) : null}
            {outcome.nameSimilarity === null ? null : (
              <div className="flex flex-wrap items-baseline justify-between gap-4">
                <dt className="text-caption text-neutral-600">
                  Name similarity
                </dt>
                <dd className="font-medium tabular-nums">
                  {percentOf(outcome.nameSimilarity)}
                </dd>
              </div>
            )}
          </dl>

          {verified ? null : (
            <>
              <ul className="space-y-2 border-t border-neutral-200 pt-4">
                {GATED_CAPABILITIES.map((capability) => (
                  <li key={capability} className="flex items-center gap-2">
                    <StatusPill
                      presentation={{ label: "Gated", variant: "warning" }}
                    />
                    <span className="text-caption">{capability}</span>
                  </li>
                ))}
              </ul>
              <div className="flex flex-wrap gap-2">
                <Button variant="outline" size="sm" disabled>
                  Correct the registration number
                </Button>
                <Button variant="ghost" size="sm" disabled>
                  Request manual verification
                </Button>
              </div>
              <p className="text-caption text-pretty text-neutral-600">
                Neither route is available yet. Correcting the number re-runs
                the registry check and can change this account&apos;s state, so
                it needs its own endpoint; manual verification needs the admin
                portal.
              </p>
            </>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Product categories</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-caption text-pretty text-neutral-600">
            These determine which claim types and checkpoint expectations you
            see by default. A batch belongs to exactly one category for its
            entire lifetime.
          </p>
          <div className="flex flex-wrap gap-2">
            {organization.productCategories.length === 0 ? (
              <span className="text-caption text-neutral-600">
                None declared. Credit buyers do not produce batches.
              </span>
            ) : (
              organization.productCategories.map((category) => (
                <StatusPill
                  key={category}
                  presentation={{
                    label:
                      productCategoryLabels[category as ProductCategory] ??
                      category,
                    variant: "primary",
                  }}
                  showDot={false}
                />
              ))
            )}
          </div>
        </CardContent>
      </Card>
    </>
  );
}
