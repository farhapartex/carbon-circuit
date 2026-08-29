import { Check, TriangleAlert } from "lucide-react";
import Link from "next/link";

const guaranteed = [
  "A credit is issued exactly once, against exactly one approved claim, enforced on the ledger rather than by a database row.",
  "A credit permanently carries the facility that earned it, the vintage year, and the activity type, through every transfer.",
  "A retired credit can never be resold, re-transferred, or retired again by anyone.",
];

const notGuaranteed = [
  "That the same real-world reduction was not also registered with Verra, Gold Standard, or a national scheme.",
];

const mitigations = [
  "A mandatory exclusivity attestation recorded against the submitting user at claim time",
  "A duplicate-evidence check across the whole platform, by content hash and by semantic similarity",
  "A public retirement log that lets any third party audit what was claimed",
];

export function TrustBand() {
  return (
    <section className="border-y border-neutral-200 bg-neutral-100">
      <div className="mx-auto max-w-6xl space-y-8 px-6 py-20">
        <div className="max-w-2xl space-y-3">
          <h2 className="text-section-heading">
            What this does and does not guarantee
          </h2>
          <p className="text-body text-pretty text-neutral-600">
            Carbon markets have a credibility problem built on claims nobody
            could check. Stating the boundary plainly is the alternative.
          </p>
        </div>

        <div className="grid gap-4 lg:grid-cols-2">
          <div className="space-y-3 rounded-lg border border-success-600/30 bg-success-50 p-6">
            <h3 className="font-medium text-success-700">
              Enforced on the ledger
            </h3>
            <ul className="space-y-2">
              {guaranteed.map((item) => (
                <li
                  key={item}
                  className="flex gap-2 text-caption text-success-700"
                >
                  <Check className="mt-0.5 size-4 shrink-0" aria-hidden />
                  {item}
                </li>
              ))}
            </ul>
          </div>

          <div className="space-y-3 rounded-lg border border-warning-600/30 bg-warning-50 p-6">
            <h3 className="font-medium text-warning-700">
              What we cannot detect
            </h3>
            <ul className="space-y-2">
              {notGuaranteed.map((item) => (
                <li
                  key={item}
                  className="flex gap-2 text-caption text-warning-700"
                >
                  <TriangleAlert
                    className="mt-0.5 size-4 shrink-0"
                    aria-hidden
                  />
                  {item}
                </li>
              ))}
            </ul>
            <p className="text-caption font-medium text-warning-700">
              What we do about it
            </p>
            <ul className="space-y-1">
              {mitigations.map((item) => (
                <li key={item} className="text-caption text-warning-700">
                  {item}
                </li>
              ))}
            </ul>
          </div>
        </div>

        <p className="text-caption text-neutral-600">
          This disclosure appears on every marketplace listing, not only here.{" "}
          <Link
            href="/marketplace/retirements"
            className="rounded-sm font-medium text-primary-700 underline underline-offset-4"
          >
            Audit the public retirement log
          </Link>
          .
        </p>
      </div>
    </section>
  );
}
