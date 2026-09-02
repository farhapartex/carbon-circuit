"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
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
import type { FacilityRecord } from "@/lib/api/facilities";
import { useFormDraftStore } from "@/stores/form-drafts";

type ClaimWizardProps = {
  facilities: FacilityRecord[];
  userName: string;
};

const today = () => new Date().toISOString().slice(0, 10);

export function ClaimWizard({ facilities, userName }: ClaimWizardProps) {
  const router = useRouter();
  const saveDraft = useFormDraftStore((state) => state.saveDraft);
  const draft = useFormDraftStore((state) => state.drafts.claim);

  const [stepIndex, setStepIndex] = useState(draft?.step ?? 0);
  const [failure, setFailure] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

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

  const persist = (nextStep: number) => {
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
      evidenceIds: [],
    });
  };

  const advance = async () => {
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

  const submit = () => {
    setFailure(null);
    startTransition(() => {
      setFailure(
        "Claims cannot be submitted yet — the sustainability and evidence services are not built.",
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
          />
        ) : null}

        {step === "evidence" ? <EvidenceUploadStep /> : null}

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
        />
      </form>
    </Form>
  );
}
