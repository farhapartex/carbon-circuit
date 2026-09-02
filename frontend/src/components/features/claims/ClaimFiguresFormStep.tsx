"use client";

import type { Control } from "react-hook-form";
import { CalculatedCeilingPreview } from "@/components/features/claims/CalculatedCeilingPreview";
import type { ClaimDraftValues } from "@/components/features/claims/claimDraft";
import {
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
import { gridRegionLabels, recycledMaterialLabels } from "@/lib/labels";
import type { ActivityType, GridRegion, RecycledMaterial } from "@/lib/types";

const gridRegions = Object.keys(gridRegionLabels) as GridRegion[];
const materials = Object.keys(recycledMaterialLabels) as RecycledMaterial[];

type ClaimFiguresFormStepProps = {
  control: Control<ClaimDraftValues>;
  activityType: ActivityType;
};

export function ClaimFiguresFormStep({
  control,
  activityType,
}: ClaimFiguresFormStepProps) {
  return (
    <div className="space-y-6">
      {activityType === "renewable_energy" ? (
        <>
          <FormField
            control={control}
            name="verifiedKwh"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Verified renewable energy, kWh</FormLabel>
                <FormControl>
                  <Input
                    inputMode="decimal"
                    placeholder="12400000"
                    {...field}
                    value={field.value ?? ""}
                  />
                </FormControl>
                <FormDescription>
                  Only energy you can evidence as renewable from your supply.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={control}
            name="gridRegion"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Grid region</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value ?? ""}
                >
                  <FormControl>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select the grid region" />
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
                  Sets the emission factor applied, so it must match the grid
                  the facility actually draws from.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </>
      ) : null}

      {activityType === "reduced_emission_logistics" ? (
        <>
          <FormField
            control={control}
            name="tonneKilometres"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Tonne-kilometres shipped</FormLabel>
                <FormControl>
                  <Input
                    inputMode="decimal"
                    placeholder="480000"
                    {...field}
                    value={field.value ?? ""}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={control}
            name="actualFactorKgPerTonneKm"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Actual factor, kgCO2e per tonne-km</FormLabel>
                <FormControl>
                  <Input
                    inputMode="decimal"
                    placeholder="0.012"
                    {...field}
                    value={field.value ?? ""}
                  />
                </FormControl>
                <FormDescription>
                  Derived from carrier fuel records. A factor at or above the
                  baseline for the route yields no credits and is rejected at
                  submission.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </>
      ) : null}

      {activityType === "responsible_sourcing" ? (
        <>
          <FormField
            control={control}
            name="material"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Material</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value ?? ""}
                >
                  <FormControl>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select the material" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {materials.map((material) => (
                      <SelectItem key={material} value={material}>
                        {recycledMaterialLabels[material]}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />

          <div className="grid gap-6 sm:grid-cols-2">
            <FormField
              control={control}
              name="verifiedQuantity"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Verified quantity</FormLabel>
                  <FormControl>
                    <Input
                      inputMode="decimal"
                      placeholder="820"
                      {...field}
                      value={field.value ?? ""}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={control}
              name="quantityUnit"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Unit</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value ?? "tonne"}
                  >
                    <FormControl>
                      <SelectTrigger className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="tonne">Tonnes</SelectItem>
                      <SelectItem value="kilogram">Kilograms</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </>
      ) : null}

      <FormField
        control={control}
        name="requestedAmount"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Requested amount, tCO2e</FormLabel>
            <FormControl>
              <Input inputMode="decimal" placeholder="6000.000000" {...field} />
            </FormControl>
            <FormDescription>
              What you are asking to be issued. It can never exceed the computed
              ceiling, whatever you enter here.
            </FormDescription>
            <FormMessage />
          </FormItem>
        )}
      />

      <CalculatedCeilingPreview />
    </div>
  );
}
