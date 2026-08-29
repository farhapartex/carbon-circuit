import type { Metadata } from "next";
import { VerificationStatusBadge } from "@/components/shared/StatusBadges";
import { StatusPill } from "@/components/shared/StatusPill";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getCurrentOrganization } from "@/lib/fixtures";
import type { OrganizationType, ProductCategory } from "@/lib/types";

export const metadata: Metadata = { title: "Organization" };

const TYPE_LABELS: Record<OrganizationType, string> = {
  manufacturer: "Manufacturer",
  assembler: "Assembler",
  logistics: "Logistics partner",
  credit_buyer: "Credit buyer",
};

const CATEGORY_LABELS: Record<ProductCategory, string> = {
  electronics: "Electronics",
  agriculture: "Agriculture",
  pharma: "Pharma",
  textiles: "Textiles",
};

const GATED_CAPABILITIES = [
  "Submit sustainability claims",
  "Receive credit issuance",
  "List credits for sale",
];

export default async function SettingsOrganizationPage() {
  const organization = await getCurrentOrganization();
  const verified = organization.verificationStatus === "verified";

  return (
    <>
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
              <dd className="font-medium">{TYPE_LABELS[organization.type]}</dd>
            </div>
            <div>
              <dt className="text-caption text-neutral-600">
                Country of incorporation
              </dt>
              <dd className="font-medium">
                {organization.countryOfIncorporation}
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
              : "Your registration number did not match an active entity in the business registry. You can use the product, but three capabilities are gated until that is resolved."}
          </p>

          {verified ? null : (
            <>
              <ul className="space-y-2">
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
                <Button variant="outline" size="sm">
                  Correct the registration number
                </Button>
                <Button variant="ghost" size="sm">
                  Request manual verification
                </Button>
              </div>
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
                    label: CATEGORY_LABELS[category],
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
