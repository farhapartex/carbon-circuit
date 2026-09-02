"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Check } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { useForm, useWatch } from "react-hook-form";
import { BatchDetailsFormStep } from "@/components/features/provenance/BatchDetailsFormStep";
import { FormNavigationFooter } from "@/components/features/provenance/FormNavigationFooter";
import { ProductCategorySelector } from "@/components/features/provenance/ProductCategorySelector";
import { ReviewSummaryStep } from "@/components/features/provenance/ReviewSummaryStep";
import {
  BATCH_STEPS,
  batchDraftSchema,
  stepFields,
  type BatchDraftValues,
  type BatchStep,
} from "@/components/features/provenance/batchDraft";
import { Form } from "@/components/ui/form";
import { useFormDraftStore } from "@/stores/form-drafts";
import type { Facility, ProductCategory } from "@/lib/types";
import { cn } from "@/lib/utils";

const STEP_LABELS: Record<BatchStep, string> = {
  category: "Category",
  details: "Details",
  review: "Review",
};

type BatchWizardProps = {
  facilities: Facility[];
  availableCategories: ProductCategory[];
};

export function BatchWizard({
  facilities,
  availableCategories,
}: BatchWizardProps) {
  const router = useRouter();
  const saveDraft = useFormDraftStore((state) => state.saveDraft);
  const draft = useFormDraftStore((state) => state.drafts.batch);

  const [stepIndex, setStepIndex] = useState(draft?.step ?? 0);
  const [failure, setFailure] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  const form = useForm<BatchDraftValues>({
    resolver: zodResolver(batchDraftSchema),
    defaultValues: {
      productCategory:
        (draft?.values.productCategory as ProductCategory | undefined) ??
        availableCategories[0] ??
        "electronics",
      acknowledgedCategory: false,
      originatingFacilityId: String(draft?.values.originatingFacilityId ?? ""),
      componentType: String(draft?.values.componentType ?? ""),
      lotNumber: String(draft?.values.lotNumber ?? ""),
      externalId: String(draft?.values.externalId ?? ""),
      quantity: String(draft?.values.quantity ?? ""),
      unit: String(draft?.values.unit ?? ""),
      producedAt: String(
        draft?.values.producedAt ?? new Date().toISOString().slice(0, 10),
      ),
      parentReferences: [],
    },
  });

  const step = BATCH_STEPS[stepIndex] ?? "category";
  const values = useWatch({ control: form.control });
  const category = useWatch({ control: form.control, name: "productCategory" });
  const acknowledged = useWatch({
    control: form.control,
    name: "acknowledgedCategory",
  });

  const persist = (nextStep: number) => {
    const current = form.getValues();
    saveDraft("batch", {
      step: nextStep,
      values: {
        productCategory: current.productCategory,
        originatingFacilityId: current.originatingFacilityId,
        componentType: current.componentType,
        lotNumber: current.lotNumber ?? "",
        externalId: current.externalId ?? "",
        quantity: current.quantity,
        unit: current.unit,
        producedAt: current.producedAt,
      },
      evidenceIds: [],
    });
  };

  const advance = async () => {
    const valid = await form.trigger(stepFields[step]);
    if (!valid) return;

    const nextStep = Math.min(stepIndex + 1, BATCH_STEPS.length - 1);
    persist(nextStep);
    setStepIndex(nextStep);
  };

  const retreat = () => {
    const previous = Math.max(stepIndex - 1, 0);
    persist(previous);
    setStepIndex(previous);
  };

  const submit = () => {
    setFailure(null);
    startTransition(() => {
      setFailure(
        "Batches cannot be created yet — the provenance service is not built.",
      );
      router.refresh();
    });
  };

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(submit)}
        className="max-w-2xl space-y-8"
      >
        <ol className="flex flex-wrap items-center gap-x-2 gap-y-3">
          {BATCH_STEPS.map((name, index) => {
            const done = index < stepIndex;
            const active = index === stepIndex;

            return (
              <li key={name} className="flex items-center gap-2">
                <span
                  aria-current={active ? "step" : undefined}
                  className={cn(
                    "flex size-6 shrink-0 items-center justify-center rounded-full text-caption font-medium tabular-nums",
                    done && "bg-primary-700 text-white",
                    active &&
                      "bg-primary-50 text-primary-800 ring-2 ring-primary-600",
                    !done && !active && "bg-neutral-100 text-neutral-600",
                  )}
                >
                  {done ? (
                    <Check className="size-3.5" aria-hidden />
                  ) : (
                    index + 1
                  )}
                </span>
                <span
                  className={cn(
                    "text-caption",
                    active
                      ? "font-medium text-neutral-900"
                      : "text-neutral-600",
                  )}
                >
                  {STEP_LABELS[name]}
                </span>
                {index < BATCH_STEPS.length - 1 ? (
                  <span
                    className="ml-2 hidden h-px w-8 bg-neutral-200 sm:block"
                    aria-hidden
                  />
                ) : null}
              </li>
            );
          })}
        </ol>

        {failure ? (
          <div
            role="alert"
            className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3"
          >
            <p className="text-helper text-warning-700">{failure}</p>
          </div>
        ) : null}

        {step === "category" ? (
          <ProductCategorySelector
            available={availableCategories}
            value={category}
            onValueChange={(category) =>
              form.setValue("productCategory", category)
            }
            acknowledged={acknowledged}
            onAcknowledgedChange={(acknowledged) =>
              form.setValue("acknowledgedCategory", acknowledged, {
                shouldValidate: true,
              })
            }
          />
        ) : null}

        {step === "details" ? (
          <BatchDetailsFormStep
            control={form.control}
            facilities={facilities}
          />
        ) : null}

        {step === "review" ? (
          <ReviewSummaryStep
            values={values as BatchDraftValues}
            facilities={facilities}
          />
        ) : null}

        <FormNavigationFooter
          onBack={stepIndex > 0 ? retreat : undefined}
          onNext={advance}
          isFinalStep={step === "review"}
          nextLabel={step === "review" ? "Create batch" : "Continue"}
          submitting={pending}
        />
      </form>
    </Form>
  );
}
