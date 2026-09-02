"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { useForm, useWatch } from "react-hook-form";
import { z } from "zod";
import { submitCheckpoint } from "@/lib/actions/batches";
import { Button } from "@/components/ui/button";
import { Combobox } from "@/components/ui/combobox";
import {
  Form,
  FormControl,
  FormDescription,
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
import { allCountryOptions } from "@/lib/countries";
import {
  checkpointTypeLabels,
  movesGoods,
  shippingMethodLabels,
} from "@/lib/labels";
import type { CheckpointType, ShippingMethod } from "@/lib/types";

const countryChoices = allCountryOptions().map((country) => ({
  value: country.code,
  label: country.name,
}));

const checkpointTypes = Object.keys(checkpointTypeLabels) as CheckpointType[];
const shippingMethods = Object.keys(shippingMethodLabels) as ShippingMethod[];

const toLocalInput = (iso: string) => iso.slice(0, 16);

const failureMessage = (code: string) => {
  if (code === "FORBIDDEN") {
    return "Only the organization that owns this batch can log checkpoints on it.";
  }
  if (code === "VALIDATION_ERROR") {
    return "Check the event time and shipping method, then try again.";
  }
  if (code === "CONFLICT") {
    return "That checkpoint was already logged. Reload the batch to see it.";
  }
  if (code === "RESOURCE_NOT_FOUND") {
    return "This batch no longer exists.";
  }
  return "We could not log this checkpoint. Please try again.";
};

type CheckpointFormProps = {
  batchId: string;
  batchLabel: string;
  producedAt: string;
};

export function CheckpointForm({
  batchId,
  batchLabel,
  producedAt,
}: CheckpointFormProps) {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [failure, setFailure] = useState<string | null>(null);
  const [idempotencyKey] = useState(() => crypto.randomUUID());

  const earliest = toLocalInput(producedAt);
  const latest = toLocalInput(new Date().toISOString());

  const schema = z
    .object({
      type: z.enum(checkpointTypes as [CheckpointType, ...CheckpointType[]]),
      locationLabel: z
        .string()
        .min(2, "Name the place this happened.")
        .max(120),
      countryCode: z.string().length(2, "Select a country."),
      occurredAt: z
        .string()
        .min(1, "When did this happen?")
        .refine((value) => value <= latest, {
          message: "A checkpoint cannot be logged for a future time.",
        })
        .refine((value) => value >= earliest, {
          message: "A checkpoint cannot precede the batch's production date.",
        }),
      shippingMethod: z.string().optional(),
    })
    .refine(
      (values) => !movesGoods(values.type) || Boolean(values.shippingMethod),
      {
        path: ["shippingMethod"],
        message:
          "A movement needs a shipping method to carry an emissions factor.",
      },
    );

  type CheckpointValues = z.infer<typeof schema>;

  const form = useForm<CheckpointValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      type: "departed_origin",
      locationLabel: "",
      countryCode: "TW",
      occurredAt: latest,
      shippingMethod: "",
    },
  });

  const selectedType = useWatch({ control: form.control, name: "type" });

  const submit = (values: CheckpointValues) => {
    setFailure(null);
    startTransition(async () => {
      const result = await submitCheckpoint(
        batchId,
        {
          type: values.type,
          locationLabel: values.locationLabel,
          countryCode: values.countryCode,
          shippingMethod: values.shippingMethod ?? "",
          occurredAt: new Date(values.occurredAt).toISOString(),
        },
        idempotencyKey,
      );

      if (!result.ok) {
        setFailure(failureMessage(result.code));
        return;
      }

      router.push(`/batches/${batchId}`);
    });
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(submit)} className="max-w-xl space-y-6">
        {failure ? (
          <div
            role="alert"
            className="rounded-md border border-danger-600 bg-danger-50 px-4 py-3"
          >
            <p className="text-helper text-danger-700">{failure}</p>
          </div>
        ) : null}

        <FormField
          control={form.control}
          name="type"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Checkpoint type</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {checkpointTypes.map((type) => (
                    <SelectItem key={type} value={type}>
                      {checkpointTypeLabels[type]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormDescription>
                Recorded against {batchLabel}. Checkpoints are append-only — a
                mistake is corrected by a superseding entry, never edited.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="locationLabel"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Location</FormLabel>
              <FormControl>
                <Input placeholder="Port of Kaohsiung" {...field} />
              </FormControl>
              <FormDescription>
                Where the event happened, as it would appear on a shipping
                document.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="countryCode"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Country</FormLabel>
              <FormControl>
                <Combobox
                  options={countryChoices}
                  value={field.value}
                  onValueChange={field.onChange}
                  placeholder="Select a country"
                  searchPlaceholder="Search countries"
                  emptyMessage="No country matches that."
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="occurredAt"
          render={({ field }) => (
            <FormItem>
              <FormLabel>When it happened</FormLabel>
              <FormControl>
                <Input
                  type="datetime-local"
                  min={earliest}
                  max={latest}
                  {...field}
                />
              </FormControl>
              <FormDescription>
                The event time, not the time you are reporting it.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {movesGoods(selectedType) ? (
          <FormField
            control={form.control}
            name="shippingMethod"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Shipping method</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value ?? ""}
                >
                  <FormControl>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select a method" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {shippingMethods.map((method) => (
                      <SelectItem key={method} value={method}>
                        {shippingMethodLabels[method]}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormDescription>
                  This determines the emissions factor applied to the leg, so a
                  movement cannot be logged without it.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        ) : null}

        <input type="hidden" name="batchId" value={batchId} />

        <Button type="submit" size="lg" disabled={pending}>
          Log checkpoint
        </Button>
      </form>
    </Form>
  );
}
