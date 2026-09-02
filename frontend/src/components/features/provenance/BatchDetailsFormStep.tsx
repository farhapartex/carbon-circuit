"use client";

import { Plus, X } from "lucide-react";
import type { Control } from "react-hook-form";
import { useFieldArray } from "react-hook-form";
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Facility } from "@/lib/types";
import type { BatchDraftValues } from "@/components/features/provenance/batchDraft";

type BatchDetailsFormStepProps = {
  control: Control<BatchDraftValues>;
  facilities: Facility[];
};

export function BatchDetailsFormStep({
  control,
  facilities,
}: BatchDetailsFormStepProps) {
  const parents = useFieldArray({ control, name: "parentReferences" });

  return (
    <div className="space-y-6">
      <FormField
        control={control}
        name="originatingFacilityId"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Originating facility</FormLabel>
            <Select onValueChange={field.onChange} defaultValue={field.value}>
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
            <FormDescription>
              The site that produced this batch. Facilities must be registered
              before a batch can cite them.
            </FormDescription>
            <FormMessage />
          </FormItem>
        )}
      />

      <FormField
        control={control}
        name="componentType"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Component type</FormLabel>
            <FormControl>
              <Input placeholder="300mm silicon wafer, 5nm node" {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <div className="grid gap-6 sm:grid-cols-2">
        <FormField
          control={control}
          name="lotNumber"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Lot number (optional)</FormLabel>
              <FormControl>
                <Input placeholder="WL-884" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={control}
          name="externalId"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Your reference (optional)</FormLabel>
              <FormControl>
                <Input placeholder="ERP-WL-884" {...field} />
              </FormControl>
              <FormDescription>
                The identifier this batch has in your own system.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <div className="grid gap-6 sm:grid-cols-2">
        <FormField
          control={control}
          name="quantity"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Quantity</FormLabel>
              <FormControl>
                <Input inputMode="decimal" placeholder="5000" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={control}
          name="unit"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Unit</FormLabel>
              <FormControl>
                <Input placeholder="wafers" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <FormField
        control={control}
        name="producedAt"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Production date</FormLabel>
            <FormControl>
              <Input type="date" {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <fieldset className="space-y-3">
        <legend className="font-medium">Component batches (optional)</legend>
        <p className="text-caption text-pretty text-neutral-600">
          If this batch was made from components you bought, enter the public
          reference each supplier gave you. A reference that matches a
          registered batch resolves and raises this batch&apos;s Provenance
          Score; one that does not still records the declaration.
        </p>

        {parents.fields.map((entry, index) => (
          <FormField
            key={entry.id}
            control={control}
            name={`parentReferences.${index}.value`}
            render={({ field }) => (
              <FormItem>
                <div className="flex items-start gap-2">
                  <FormControl>
                    <Input
                      placeholder="mR3xKp8Zc1Vt5nQwJ7dBaU"
                      className="font-mono"
                      {...field}
                    />
                  </FormControl>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => parents.remove(index)}
                    aria-label={`Remove component batch ${index + 1}`}
                  >
                    <X className="size-4" aria-hidden />
                  </Button>
                </div>
                <FormMessage />
              </FormItem>
            )}
          />
        ))}

        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => parents.append({ value: "" })}
        >
          <Plus className="size-4" aria-hidden />
          Add a component batch
        </Button>
      </fieldset>
    </div>
  );
}
