"use client";

import { Button } from "@/components/ui/button";

type FormNavigationFooterProps = {
  onBack?: (() => void) | undefined;
  onNext?: (() => void) | undefined;
  nextLabel?: string;
  submitting?: boolean;
  isFinalStep?: boolean;
};

export function FormNavigationFooter({
  onBack,
  onNext,
  nextLabel = "Continue",
  submitting = false,
  isFinalStep = false,
}: FormNavigationFooterProps) {
  return (
    <div className="flex items-center justify-between gap-4 border-t border-neutral-200 pt-6">
      {onBack ? (
        <Button type="button" variant="outline" onClick={onBack}>
          Back
        </Button>
      ) : (
        <span />
      )}

      {isFinalStep ? (
        <Button type="submit" size="lg" disabled={submitting}>
          {nextLabel}
        </Button>
      ) : (
        <Button type="button" size="lg" onClick={onNext}>
          {nextLabel}
        </Button>
      )}
    </div>
  );
}
