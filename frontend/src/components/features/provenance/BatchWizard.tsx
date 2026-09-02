"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
import { StepIndicator } from "@/components/shared/StepIndicator";
import { submitBatch } from "@/lib/actions/batches";
import { Form } from "@/components/ui/form";
import { useFormDraftStore } from "@/stores/form-drafts";
import type { FacilityRecord } from "@/lib/api/facilities";
import type { ProductCategory } from "@/lib/types";

const failureMessage = (code: string) => {
  if (code === "FORBIDDEN") {
    return "Your organization cannot create batches on its current plan or state.";
  }
  if (code === "VALIDATION_ERROR") {
    return "Check the quantity, production date, and component references.";
  }
  if (code === "CONFLICT") {
    return "A batch with that reference already exists for your organization.";
  }
  return "We could not create this batch. Please try again.";
};

const STEP_LABELS: Record<BatchStep, string> = {
  category: "Category",
  details: "Details",
  review: "Review",
};

type BatchWizardProps = {
  facilities: FacilityRecord[];
  availableCategories: ProductCategory[];
};

export function BatchWizard({
  facilities,
  availableCategories,
}: BatchWizardProps) {
  const router = useRouter();
  const saveDraft = useFormDraftStore((state) => state.saveDraft);
  const clearDraft = useFormDraftStore((state) => state.clearDraft);
  const draft = useFormDraftStore((state) => state.drafts.batch);

  const [stepIndex, setStepIndex] = useState(draft?.step ?? 0);
  const [failure, setFailure] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();
  const [idempotencyKey] = useState(() => crypto.randomUUID());

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

  const submit = (values: BatchDraftValues) => {
    setFailure(null);
    startTransition(async () => {
      const result = await submitBatch(
        {
          originatingFacilityId: values.originatingFacilityId,
          productCategory: values.productCategory,
          componentType: values.componentType,
          lotNumber: values.lotNumber ?? "",
          quantity: values.quantity,
          unit: values.unit,
          producedAt: new Date(`${values.producedAt}T00:00:00Z`).toISOString(),
          externalId: values.externalId ?? "",
          parentReferences: values.parentReferences
            .map((parent) => parent.value)
            .filter((value) => value.length > 0),
        },
        idempotencyKey,
      );

      if (!result.ok) {
        setFailure(failureMessage(result.code));
        return;
      }

      clearDraft("batch");
      router.push(`/batches/${result.detail.batch.id}`);
    });
  };

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(submit)}
        className="max-w-2xl space-y-8"
      >
        <StepIndicator
          steps={BATCH_STEPS.map((name) => STEP_LABELS[name])}
          currentIndex={stepIndex}
        />

        {failure ? (
          <div
            role="alert"
            className="rounded-md border border-danger-600 bg-danger-50 px-4 py-3"
          >
            <p className="text-helper text-danger-700">{failure}</p>
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
