"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
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
import { submitFacility } from "@/lib/actions/facilities";
import { allCountryOptions } from "@/lib/countries";
import { facilityTypeLabels, gridRegionLabels } from "@/lib/labels";
import type { FacilityType, GridRegion } from "@/lib/types";

const countryChoices = allCountryOptions().map((country) => ({
  value: country.code,
  label: country.name,
}));

const gridRegions = Object.keys(gridRegionLabels) as GridRegion[];
const facilityTypes = Object.keys(facilityTypeLabels) as FacilityType[];

const positiveDecimal = (label: string) =>
  z
    .string()
    .regex(
      /^\d+(\.\d{1,6})?$/,
      `${label} must be a plain number with up to six decimal places.`,
    )
    .refine((value) => Number(value) > 0, {
      message: `${label} must be greater than zero.`,
    });

const facilitySchema = z.object({
  name: z.string().min(2, "Name this site as your organization refers to it."),
  facilityReference: z.string().max(64).optional(),
  address: z.string().min(6, "Enter the physical address of the site."),
  countryCode: z.string().length(2, "Select a country."),
  gridRegion: z.enum(gridRegions as [GridRegion, ...GridRegion[]]),
  type: z.enum(facilityTypes as [FacilityType, ...FacilityType[]]),
  declaredAnnualProductionCapacity: positiveDecimal("Production capacity"),
  declaredAnnualEnergyConsumptionKwh: positiveDecimal("Energy consumption"),
});

type FacilityValues = z.infer<typeof facilitySchema>;

const failureMessage = (code: string) => {
  if (code === "FORBIDDEN") {
    return "Only an owner or admin can add a facility.";
  }
  if (code === "VALIDATION_ERROR") {
    return "Check the figures and the grid region, then try again.";
  }
  if (code === "CONFLICT") {
    return "That request was already submitted. Reload the facilities list.";
  }
  return "We could not add this facility. Please try again.";
};

export function FacilityForm() {
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  const [failure, setFailure] = useState<string | null>(null);
  const [idempotencyKey] = useState(() => crypto.randomUUID());

  const form = useForm<FacilityValues>({
    resolver: zodResolver(facilitySchema),
    defaultValues: {
      name: "",
      facilityReference: "",
      address: "",
      countryCode: "TW",
      gridRegion: "TW",
      type: "component_fabrication",
      declaredAnnualProductionCapacity: "",
      declaredAnnualEnergyConsumptionKwh: "",
    },
  });

  const submit = (values: FacilityValues) => {
    setFailure(null);
    startTransition(async () => {
      try {
        const result = await submitFacility(
          {
            name: values.name,
            address: values.address,
            countryCode: values.countryCode,
            gridRegion: values.gridRegion,
            type: values.type,
            facilityReference: values.facilityReference ?? "",
            declaredAnnualProductionCapacity:
              values.declaredAnnualProductionCapacity,
            declaredAnnualEnergyConsumptionKwh:
              values.declaredAnnualEnergyConsumptionKwh,
          },
          idempotencyKey,
        );

        if (!result.ok) {
          setFailure(failureMessage(result.code));
          return;
        }

        router.push(`/facilities/${result.facility.id}`);
      } catch (error) {
        setFailure(
          `The request could not be sent: ${error instanceof Error ? error.message : "unknown error"}`,
        );
      }
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
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Facility name</FormLabel>
              <FormControl>
                <Input placeholder="Hsinchu Fab TW-01" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="facilityReference"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Facility reference (optional)</FormLabel>
              <FormControl>
                <Input placeholder="TW-HSC-01" {...field} />
              </FormControl>
              <FormDescription>
                The site identifier on your organization&apos;s registration
                record. Providing it lets us match this facility against the
                registry, which removes the discount applied to your credit
                ceiling.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="address"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Address</FormLabel>
              <FormControl>
                <Input
                  placeholder="No. 8 Li-Hsin Road, Hsinchu Science Park"
                  {...field}
                />
              </FormControl>
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
          name="gridRegion"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Grid region</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {gridRegions.map((region) => (
                    <SelectItem key={region} value={region}>
                      {gridRegionLabels[region]}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormDescription>
                Determines the grid emission factor used when this facility
                claims renewable energy, so it must be the grid the site
                actually draws from.
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
              <FormLabel>Facility type</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {facilityTypes.map((type) => (
                    <SelectItem key={type} value={type}>
                      {facilityTypeLabels[type]}
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
          name="declaredAnnualProductionCapacity"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Declared annual production capacity</FormLabel>
              <FormControl>
                <Input inputMode="decimal" placeholder="18000000" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="declaredAnnualEnergyConsumptionKwh"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Declared annual energy consumption, kWh</FormLabel>
              <FormControl>
                <Input inputMode="decimal" placeholder="31000000" {...field} />
              </FormControl>
              <FormDescription>
                These figures cap what this facility can ever claim. An
                unmatched facility is limited to half of what its declared scale
                would support, so overstating them does not help you.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit" size="lg" disabled={pending}>
          Add facility
        </Button>
      </form>
    </Form>
  );
}
