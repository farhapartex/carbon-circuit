"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useMemo, useState, useTransition } from "react";
import { useForm, useWatch } from "react-hook-form";
import { ActivityTypeSelector } from "@/components/features/claims/ActivityTypeSelector";
import { ClaimFiguresFormStep } from "@/components/features/claims/ClaimFiguresFormStep";
import { ClaimReviewStep } from "@/components/features/claims/ClaimReviewStep";
import { EvidenceUploadStep } from "@/components/features/claims/EvidenceUploadStep";
import { ExclusivityAttestationStep } from "@/components/features/claims/ExclusivityAttestationStep";
import {
  CLAIM_STEPS,
  claimDraftSchema,
  claimStepFields,
  claimStepLabels,
  type ClaimDraftValues,
} from "@/components/features/claims/claimDraft";
import { FormNavigationFooter } from "@/components/features/provenance/FormNavigationFooter";
import { StepIndicator } from "@/components/shared/StepIndicator";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { submitSustainabilityClaim } from "@/lib/actions/claims";
import type { EvidenceDocument } from "@/lib/api/evidence";
import type { FacilityRecord } from "@/lib/api/facilities";
import { useFormDraftStore } from "@/stores/form-drafts";

type ClaimWizardProps = {
  facilities: FacilityRecord[];
  userName: string;
  uploadedEvidence: EvidenceDocument[];
};

const today = () => new Date().toISOString().slice(0, 10);

const declaredFiguresOf = (draft: ClaimDraftValues): Record<string, string> => {
  if (draft.activityType === "reduced_emission_logistics") {
    return {
      tonne_kilometres: draft.tonneKilometres ?? "",
      actual_factor_kg_per_tonne_km: draft.actualFactorKgPerTonneKm ?? "",
    };
  }

  if (draft.activityType === "responsible_sourcing") {
    return {
      material: draft.material ?? "",
      verified_quantity: draft.verifiedQuantity ?? "",
      quantity_unit: draft.quantityUnit ?? "tonne",
    };
  }

  return {
    verified_kwh: draft.verifiedKwh ?? "",
    grid_region: draft.gridRegion ?? "",
  };
};

export function ClaimWizard({
  facilities,
  userName,
  uploadedEvidence,
}: ClaimWizardProps) {
  const router = useRouter();
  const saveDraft = useFormDraftStore((state) => state.saveDraft);
  const clearDraft = useFormDraftStore((state) => state.clearDraft);
  const draft = useFormDraftStore((state) => state.drafts.claim);

  const [stepIndex, setStepIndex] = useState(draft?.step ?? 0);
  const [attachedIds, setAttachedIds] = useState<string[]>(
    draft?.evidenceIds ?? [],
  );
  const [freshlyUploaded, setFreshlyUploaded] = useState<EvidenceDocument[]>(
    [],
  );
  const [evidenceBusy, setEvidenceBusy] = useState(false);
  const [failure, setFailure] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();
  const [idempotencyKey] = useState(() => crypto.randomUUID());

  const knownEvidence = useMemo(
    () =>
      new Map(
        [...uploadedEvidence, ...freshlyUploaded].map((document) => [
          document.id,
          document,
        ]),
      ),
    [uploadedEvidence, freshlyUploaded],
  );

  const evidence = useMemo(
    () =>
      attachedIds.flatMap((id) => {
        const document = knownEvidence.get(id);
        return document ? [document] : [];
      }),
    [attachedIds, knownEvidence],
  );

  const form = useForm<ClaimDraftValues>({
    resolver: zodResolver(claimDraftSchema),
    defaultValues: {
      facilityId: String(draft?.values.facilityId ?? ""),
      activityType:
        (draft?.values.activityType as ClaimDraftValues["activityType"]) ??
        "renewable_energy",
      vintageYear: String(
        draft?.values.vintageYear ?? new Date().getFullYear(),
      ),
      periodStart: String(
        draft?.values.periodStart ?? `${new Date().getFullYear()}-01-01`,
      ),
      periodEnd: String(draft?.values.periodEnd ?? today()),
      verifiedKwh: "",
      gridRegion: undefined,
      tonneKilometres: "",
      actualFactorKgPerTonneKm: "",
      material: undefined,
      verifiedQuantity: "",
      quantityUnit: "tonne",
      requestedAmount: String(draft?.values.requestedAmount ?? ""),
      exclusivityAttested: false,
    },
  });

  const step = CLAIM_STEPS[stepIndex] ?? "activity";
  const values = useWatch({ control: form.control });
  const activityType = useWatch({
    control: form.control,
    name: "activityType",
  });
  const attested = useWatch({
    control: form.control,
    name: "exclusivityAttested",
  });

  const persist = (nextStep: number, documents = evidence) => {
    const current = form.getValues();
    saveDraft("claim", {
      step: nextStep,
      values: {
        facilityId: current.facilityId,
        activityType: current.activityType,
        vintageYear: current.vintageYear,
        periodStart: current.periodStart,
        periodEnd: current.periodEnd,
        requestedAmount: current.requestedAmount,
      },
      evidenceIds: documents.map((document) => document.id),
    });
  };

  const changeEvidence = (documents: EvidenceDocument[]) => {
    setFreshlyUploaded((existing) => {
      const seen = new Set(existing.map((document) => document.id));
      return [
        ...existing,
        ...documents.filter((document) => !seen.has(document.id)),
      ];
    });
    setAttachedIds(documents.map((document) => document.id));
    persist(stepIndex, documents);
  };

  const advance = async () => {
    if (evidenceBusy) return;

    const valid = await form.trigger(claimStepFields[step]);
    if (!valid) return;

    const nextStep = Math.min(stepIndex + 1, CLAIM_STEPS.length - 1);
    persist(nextStep);
    setStepIndex(nextStep);
  };

  const retreat = () => {
    const previous = Math.max(stepIndex - 1, 0);
    persist(previous);
    setStepIndex(previous);
  };

  const submit = (draft: ClaimDraftValues) => {
    setFailure(null);

    if (evidence.length === 0) {
      setFailure(
        "A claim needs at least one supporting document. Go back to Evidence and attach one.",
      );
      return;
    }

    startTransition(async () => {
      try {
        const result = await submitSustainabilityClaim(
          {
            facilityId: draft.facilityId,
            activityType: draft.activityType,
            vintageYear: Number(draft.vintageYear),
            periodStart: draft.periodStart,
            periodEnd: draft.periodEnd,
            declaredFigures: declaredFiguresOf(draft),
            requestedAmount: draft.requestedAmount,
            evidenceIds: evidence.map((document) => document.id),
            exclusivityAttested: draft.exclusivityAttested,
          },
          idempotencyKey,
        );

        if (!result.ok) {
          setFailure(result.message);
          return;
        }

        clearDraft("claim");
        router.push(`/claims/${result.claim.id}`);
      } catch (error) {
        setFailure(
          `The claim could not be sent: ${error instanceof Error ? error.message : "unknown error"}`,
        );
      }
    });
  };

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(submit)}
        className="max-w-2xl space-y-8"
      >
        <StepIndicator
          steps={CLAIM_STEPS.map((name) => claimStepLabels[name])}
          currentIndex={stepIndex}
        />

        {failure ? (
          <div
            role="alert"
            className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3"
          >
            <p className="text-helper text-warning-700">{failure}</p>
          </div>
        ) : null}

        {step === "activity" ? (
          <div className="space-y-6">
            <FormField
              control={form.control}
              name="facilityId"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Facility</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Select a facility" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {facilities.map((facility) => (
                        <SelectItem key={facility.id} value={facility.id}>
                          {facility.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <ActivityTypeSelector
              value={activityType}
              onValueChange={(activity) =>
                form.setValue("activityType", activity)
              }
            />

            <div className="grid gap-6 sm:grid-cols-3">
              <FormField
                control={form.control}
                name="vintageYear"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Vintage year</FormLabel>
                    <FormControl>
                      <Input inputMode="numeric" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="periodStart"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Period start</FormLabel>
                    <FormControl>
                      <Input type="date" max={today()} {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="periodEnd"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Period end</FormLabel>
                    <FormControl>
                      <Input type="date" max={today()} {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>
        ) : null}

        {step === "figures" ? (
          <ClaimFiguresFormStep
            control={form.control}
            activityType={activityType}
            facilityId={String(values.facilityId ?? "")}
            vintageYear={String(values.vintageYear ?? "")}
            periodStart={String(values.periodStart ?? "")}
            periodEnd={String(values.periodEnd ?? "")}
          />
        ) : null}

        {step === "evidence" ? (
          <EvidenceUploadStep
            documents={evidence}
            onDocumentsChange={changeEvidence}
            onBusyChange={setEvidenceBusy}
          />
        ) : null}

        {step === "attestation" ? (
          <FormField
            control={form.control}
            name="exclusivityAttested"
            render={() => (
              <FormItem>
                <ExclusivityAttestationStep
                  attested={attested}
                  onAttestedChange={(value) =>
                    form.setValue("exclusivityAttested", value, {
                      shouldValidate: true,
                    })
                  }
                  userName={userName}
                />
                <FormMessage />
              </FormItem>
            )}
          />
        ) : null}

        {step === "review" ? (
          <ClaimReviewStep
            values={values as ClaimDraftValues}
            facilities={facilities}
          />
        ) : null}

        <FormNavigationFooter
          onBack={stepIndex > 0 ? retreat : undefined}
          onNext={advance}
          isFinalStep={step === "review"}
          nextLabel={step === "review" ? "Submit claim" : "Continue"}
          submitting={pending}
          blockedReason={
            evidenceBusy
              ? "Every document has to finish scanning before you can move on."
              : step === "review" && evidence.length === 0
                ? "Attach at least one supporting document before submitting."
                : undefined
          }
        />
      </form>
    </Form>
  );
}
