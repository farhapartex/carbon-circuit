import { z } from "zod";
import { gridRegionLabels, recycledMaterialLabels } from "@/lib/labels";
import type { ActivityType, GridRegion, RecycledMaterial } from "@/lib/types";

const ACTIVITY_TYPES: [ActivityType, ...ActivityType[]] = [
  "renewable_energy",
  "reduced_emission_logistics",
  "responsible_sourcing",
];

const gridRegions = Object.keys(gridRegionLabels) as [
  GridRegion,
  ...GridRegion[],
];

const materials = Object.keys(recycledMaterialLabels) as [
  RecycledMaterial,
  ...RecycledMaterial[],
];

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

export const claimDraftSchema = z
  .object({
    facilityId: z.string().min(1, "Select the facility this claim is for."),
    activityType: z.enum(ACTIVITY_TYPES),
    vintageYear: z
      .string()
      .regex(/^\d{4}$/, "Enter a four digit year.")
      .refine((value) => Number(value) >= 2020, {
        message: "Vintage years before 2020 are not accepted.",
      })
      .refine((value) => Number(value) <= new Date().getFullYear(), {
        message: "A vintage year cannot be in the future.",
      }),
    periodStart: z.string().min(1, "When did the claim period start?"),
    periodEnd: z.string().min(1, "When did the claim period end?"),
    verifiedKwh: z.string().optional(),
    gridRegion: z.enum(gridRegions).optional(),
    tonneKilometres: z.string().optional(),
    actualFactorKgPerTonneKm: z.string().optional(),
    material: z.enum(materials).optional(),
    verifiedQuantity: z.string().optional(),
    quantityUnit: z.enum(["tonne", "kilogram"]).optional(),
    requestedAmount: positiveDecimal("Requested amount"),
    exclusivityAttested: z.boolean().refine((value) => value, {
      message: "The exclusivity attestation is required to submit a claim.",
    }),
  })
  .refine((values) => values.periodEnd >= values.periodStart, {
    path: ["periodEnd"],
    message: "The claim period cannot end before it starts.",
  })
  .refine(
    (values) =>
      values.activityType !== "renewable_energy" ||
      (Boolean(values.verifiedKwh) && Boolean(values.gridRegion)),
    {
      path: ["verifiedKwh"],
      message: "Renewable energy claims need verified kWh and a grid region.",
    },
  )
  .refine(
    (values) =>
      values.activityType !== "reduced_emission_logistics" ||
      (Boolean(values.tonneKilometres) &&
        Boolean(values.actualFactorKgPerTonneKm)),
    {
      path: ["tonneKilometres"],
      message:
        "Logistics claims need tonne-kilometres and the actual emissions factor.",
    },
  )
  .refine(
    (values) =>
      values.activityType !== "responsible_sourcing" ||
      (Boolean(values.material) && Boolean(values.verifiedQuantity)),
    {
      path: ["verifiedQuantity"],
      message: "Sourcing claims need a material and a verified quantity.",
    },
  );

export type ClaimDraftValues = z.infer<typeof claimDraftSchema>;

export const CLAIM_STEPS = [
  "activity",
  "figures",
  "evidence",
  "attestation",
  "review",
] as const;

export type ClaimStep = (typeof CLAIM_STEPS)[number];

export const claimStepLabels: Record<ClaimStep, string> = {
  activity: "Activity",
  figures: "Figures",
  evidence: "Evidence",
  attestation: "Attestation",
  review: "Review",
};

export const claimStepFields: Record<ClaimStep, (keyof ClaimDraftValues)[]> = {
  activity: [
    "facilityId",
    "activityType",
    "vintageYear",
    "periodStart",
    "periodEnd",
  ],
  figures: [
    "verifiedKwh",
    "gridRegion",
    "tonneKilometres",
    "actualFactorKgPerTonneKm",
    "material",
    "verifiedQuantity",
    "quantityUnit",
    "requestedAmount",
  ],
  evidence: [],
  attestation: ["exclusivityAttested"],
  review: [],
};
