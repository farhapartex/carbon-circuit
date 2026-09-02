import { Lock } from "lucide-react";

export function EscrowExplainer() {
  return (
    <section className="rounded-lg border border-warning-600 bg-warning-50 px-5 py-4">
      <h2 className="flex items-center gap-2 font-medium text-warning-700">
        <Lock className="size-4 shrink-0" aria-hidden />
        Listed credits leave your balance immediately
      </h2>
      <p className="mt-2 text-caption text-pretty text-warning-700">
        The moment this listing goes live, the quantity you list moves into
        escrow. It stops counting toward your available balance, so you cannot
        sell, retire, or list it anywhere else — which is what makes overselling
        structurally impossible rather than merely something we check for.
      </p>
      <p className="mt-2 text-caption text-pretty text-warning-700">
        Cancelling the listing, or letting it expire, returns whatever is unsold
        to your available balance.
      </p>
    </section>
  );
}
