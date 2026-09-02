import { Calculator } from "lucide-react";

export function CalculatedCeilingPreview() {
  return (
    <div className="rounded-lg border border-neutral-200 bg-neutral-50 px-4 py-4">
      <p className="flex items-center gap-2 font-medium">
        <Calculator className="size-4 shrink-0 text-neutral-600" aria-hidden />
        Your credit ceiling
      </p>
      <p className="mt-2 text-caption text-pretty text-neutral-600">
        The ceiling is computed by the server from your facility&apos;s attested
        capacity, the pinned reference factor for your region and vintage, and
        your facility&apos;s verification discount. It is not estimated here: a
        client-side figure that disagreed with the authoritative one would be
        worse than showing none, and the reference table versions involved are
        not the browser&apos;s to hold.
      </p>
      <p className="mt-2 text-caption text-pretty text-neutral-600">
        It will appear on this step once the sustainability service can compute
        it.
      </p>
    </div>
  );
}
