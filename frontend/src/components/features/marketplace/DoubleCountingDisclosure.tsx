import { Info } from "lucide-react";

export function DoubleCountingDisclosure() {
  return (
    <section className="rounded-lg border border-info-600 bg-info-50 px-5 py-4">
      <h2 className="flex items-center gap-2 font-medium text-info-700">
        <Info className="size-4 shrink-0" aria-hidden />
        What this credit guarantees, and what it does not
      </h2>

      <p className="mt-2 text-caption text-pretty text-info-700">
        This credit was issued exactly once against exactly one approved
        sustainability claim. Once retired it can never be resold,
        re-transferred, or retired again by anyone, and that is enforced on an
        immutable ledger rather than by a database row.
      </p>

      <p className="mt-2 text-caption text-pretty text-info-700">
        What CarbonCircuit cannot guarantee on its own is that the underlying
        real-world reduction has not <em>also</em> been registered with an
        external registry such as Verra, Gold Standard, or a national scheme. We
        prevent double-retirement of a CarbonCircuit credit; we cannot prevent
        double-issuance of the underlying reduction across unconnected
        registries. Our mitigations are the seller&apos;s mandatory exclusivity
        attestation, a duplicate-evidence check across the whole platform, and
        the public retirement log that lets any third party audit what was
        claimed.
      </p>
    </section>
  );
}
