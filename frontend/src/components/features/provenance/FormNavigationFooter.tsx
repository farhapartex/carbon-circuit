"use client";

import { Button } from "@/components/ui/button";

type FormNavigationFooterProps = {
  onBack?: (() => void) | undefined;
  onNext?: (() => void) | undefined;
  nextLabel?: string;
  submitting?: boolean;
  isFinalStep?: boolean;
  blockedReason?: string | undefined;
};

export function FormNavigationFooter({
  onBack,
  onNext,
  nextLabel = "Continue",
  submitting = false,
  isFinalStep = false,
  blockedReason,
}: FormNavigationFooterProps) {
  const blocked = submitting || Boolean(blockedReason);

  return (
    <div className="space-y-2 border-t border-neutral-200 pt-6">
      <div className="flex items-center justify-between gap-4">
        {onBack ? (
          <Button
            type="button"
            variant="outline"
            onClick={onBack}
            disabled={blocked}
          >
            Back
          </Button>
        ) : (
          <span />
        )}

        {isFinalStep ? (
          <Button type="submit" size="lg" disabled={blocked}>
            {nextLabel}
          </Button>
        ) : (
          <Button type="button" size="lg" onClick={onNext} disabled={blocked}>
            {nextLabel}
          </Button>
        )}
      </div>

      {blockedReason ? (
        <p
          className="text-right text-caption text-neutral-600"
          aria-live="polite"
        >
          {blockedReason}
        </p>
      ) : null}
    </div>
  );
}
