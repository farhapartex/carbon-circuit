"use client";

import { activityTypeLabels } from "@/lib/labels";
import type { ActivityType } from "@/lib/types";
import { cn } from "@/lib/utils";

const ACTIVITY_DETAIL: Record<ActivityType, string> = {
  renewable_energy:
    "Verified renewable kWh, multiplied by your facility's regional grid factor",
  reduced_emission_logistics:
    "Tonne-kilometres shipped below the baseline factor for that route",
  responsible_sourcing:
    "Verified recycled or responsibly-sourced material, by emissions avoided",
};

type ActivityTypeSelectorProps = {
  value: ActivityType;
  onValueChange: (activity: ActivityType) => void;
};

export function ActivityTypeSelector({
  value,
  onValueChange,
}: ActivityTypeSelectorProps) {
  return (
    <fieldset className="space-y-3">
      <legend className="font-medium">Activity type</legend>
      <div className="space-y-3">
        {(Object.keys(activityTypeLabels) as ActivityType[]).map((activity) => (
          <label
            key={activity}
            className={cn(
              "flex cursor-pointer items-start gap-3 rounded-lg border px-4 py-3",
              value === activity
                ? "border-primary-600 bg-primary-50"
                : "hover:border-neutral-300 border-neutral-200",
            )}
          >
            <input
              type="radio"
              name="activityType"
              value={activity}
              checked={value === activity}
              onChange={() => onValueChange(activity)}
              className="mt-1"
            />
            <span>
              <span className="block font-medium">
                {activityTypeLabels[activity]}
              </span>
              <span className="block text-caption text-pretty text-neutral-600">
                {ACTIVITY_DETAIL[activity]}
              </span>
            </span>
          </label>
        ))}
      </div>
      <p className="text-caption text-pretty text-neutral-600">
        The activity type fixes which formula computes your ceiling, and which
        reference table that formula draws its multiplier from.
      </p>
    </fieldset>
  );
}
