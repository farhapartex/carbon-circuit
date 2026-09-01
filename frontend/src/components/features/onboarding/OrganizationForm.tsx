"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { submitOrganization } from "@/lib/actions/organization";
import { Button } from "@/components/ui/button";
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
import { useFormDraftStore } from "@/stores/form-drafts";

const ORGANIZATION_TYPES = [
  {
    value: "manufacturer",
    label: "Manufacturer",
    detail: "Produces raw components and claims credit for its practices",
  },
  {
    value: "assembler",
    label: "Assembler",
    detail: "Buys components and assembles finished goods",
  },
  {
    value: "logistics",
    label: "Logistics partner",
    detail: "Moves goods between facilities and reports checkpoints",
  },
  {
    value: "credit_buyer",
    label: "Credit buyer",
    detail: "Only purchases and retires credits, and pays nothing",
  },
] as const;

const COUNTRIES = ["TW", "SG", "UK", "MY", "JP", "KR", "DE", "VN", "US"];

const organizationSchema = z.object({
  name: z.string().min(2, "Enter your registered legal name."),
  type: z.enum(["manufacturer", "assembler", "logistics", "credit_buyer"]),
  countryOfIncorporation: z.string().min(2, "Select a country."),
  businessRegistrationNumber: z
    .string()
    .min(
      4,
      "Enter the registration number exactly as it appears on the register.",
    ),
});

type OrganizationValues = z.infer<typeof organizationSchema>;

export function OrganizationForm() {
  const router = useRouter();
  const saveDraft = useFormDraftStore((state) => state.saveDraft);
  const draft = useFormDraftStore((state) => state.drafts.organization);

  const form = useForm<OrganizationValues>({
    resolver: zodResolver(organizationSchema),
    defaultValues: {
      name: String(draft?.values.name ?? ""),
      type:
        (draft?.values.type as OrganizationValues["type"]) ?? "manufacturer",
      countryOfIncorporation: String(
        draft?.values.countryOfIncorporation ?? "TW",
      ),
      businessRegistrationNumber: String(
        draft?.values.businessRegistrationNumber ?? "",
      ),
    },
  });

  const [pending, startTransition] = useTransition();
  const [failure, setFailure] = useState<string | null>(null);
  const [idempotencyKey] = useState(() => crypto.randomUUID());

  const submit = (values: OrganizationValues) => {
    setFailure(null);
    saveDraft("organization", { step: 1, values, evidenceIds: [] });

    startTransition(async () => {
      const result = await submitOrganization(
        {
          name: values.name,
          type: values.type,
          countryOfIncorporation: values.countryOfIncorporation,
          businessRegistrationNumber: values.businessRegistrationNumber,
          productCategories: ["electronics"],
        },
        idempotencyKey,
      );

      if (!result.ok) {
        setFailure(failureMessage(result.code));
        return;
      }

      router.push("/onboarding/verification");
    });
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(submit)} className="space-y-6">
        {failure ? (
          <div
            role="alert"
            className="rounded-md border border-danger-600 bg-danger-50 px-4 py-3"
          >
            <p className="font-700 text-body text-danger-700">
              We could not create your organization
            </p>
            <p className="text-helper text-danger-700">{failure}</p>
          </div>
        ) : null}

        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Registered legal name</FormLabel>
              <FormControl>
                <Input
                  placeholder="Formosa Precision Semiconductor Co., Ltd."
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Exactly as it appears on your business register.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="type"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Organization type</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {ORGANIZATION_TYPES.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormDescription>
                {
                  ORGANIZATION_TYPES.find(
                    (option) => option.value === field.value,
                  )?.detail
                }
                . Permanent, and it decides which plans you can choose.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="countryOfIncorporation"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Country of incorporation</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {COUNTRIES.map((country) => (
                    <SelectItem key={country} value={country}>
                      {country}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="businessRegistrationNumber"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Business registration number</FormLabel>
              <FormControl>
                <Input
                  placeholder="TW-28419377"
                  autoComplete="off"
                  className="tabular-nums"
                  {...field}
                />
              </FormControl>
              <FormDescription>
                Checked against the register when you continue.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit" size="lg" disabled={pending}>
          {pending
            ? "Checking the register…"
            : "Create organization and verify"}
        </Button>
      </form>
    </Form>
  );
}

function failureMessage(code: string): string {
  switch (code) {
    case "CONFLICT":
      return "That registration number is already registered, or you already belong to an organization.";
    case "REQUEST_IN_PROGRESS":
      return "This submission is already being processed. Give it a moment.";
    case "VALIDATION_ERROR":
      return "Check the details above and try again.";
    default:
      return "Something went wrong on our side. Try again shortly.";
  }
}
