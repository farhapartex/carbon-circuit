import { AlertTriangle, Info } from "lucide-react";
import Link from "next/link";
import type { Organization } from "@/lib/types";

const gatedCapabilities = [
  "submit sustainability claims",
  "receive credit issuance",
  "list credits for sale",
];

type VerificationStatusBannerProps = {
  organization: Organization;
};

export function VerificationStatusBanner({
  organization,
}: VerificationStatusBannerProps) {
  if (
    organization.state === "active" &&
    organization.verificationStatus === "verified"
  ) {
    return null;
  }

  if (organization.state === "restricted") {
    return (
      <div
        role="alert"
        className="flex items-start gap-3 border-b border-danger-600/30 bg-danger-50 px-6 py-3"
      >
        <AlertTriangle
          className="mt-0.5 size-4 shrink-0 text-danger-600"
          aria-hidden
        />
        <p className="text-caption text-danger-700">
          Your organization is restricted following a fraud escalation. You
          cannot {gatedCapabilities.join(", ")}. You can still log in, view your
          data, log checkpoints, and export everything you are entitled to.
        </p>
      </div>
    );
  }

  if (organization.state === "read_only") {
    return (
      <div
        role="alert"
        className="flex items-start gap-3 border-b border-warning-600/30 bg-warning-50 px-6 py-3"
      >
        <AlertTriangle
          className="mt-0.5 size-4 shrink-0 text-warning-600"
          aria-hidden
        />
        <p className="text-caption text-warning-700">
          Your subscription lapsed past the 14 day grace period, so your
          organization is read-only. Your credits are untouched and existing
          listings are still honoured.{" "}
          <Link
            href="/settings/billing"
            className="font-medium underline underline-offset-4"
          >
            Update your payment method
          </Link>
          .
        </p>
      </div>
    );
  }

  return (
    <div className="flex items-start gap-3 border-b border-warning-600/30 bg-warning-50 px-6 py-3">
      <Info className="mt-0.5 size-4 shrink-0 text-warning-600" aria-hidden />
      <p className="text-caption text-warning-700">
        Your organization did not match the business registry, so you cannot{" "}
        {gatedCapabilities.join(", ")}. Everything else works normally.{" "}
        <Link
          href="/settings/organization"
          className="font-medium underline underline-offset-4"
        >
          Correct your registration number
        </Link>{" "}
        or contact support for manual verification.
      </p>
    </div>
  );
}
