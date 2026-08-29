import type { ProvenanceBand } from "@/lib/status";
import type {
  CountryCode,
  Id,
  IsoTimestamp,
  Location,
  ProductCategory,
  PublicBatchReference,
  ShippingMethod,
  TransactionHash,
} from "@/lib/types/common";

export type CheckpointType =
  | "production_complete"
  | "departed_origin"
  | "customs_export"
  | "customs_import"
  | "arrived_destination";

export type AnchorStatus = "unanchored" | "provisional" | "confirmed";

export type CheckpointAnchor = {
  status: AnchorStatus;
  epoch: number | null;
  transactionHash: TransactionHash | null;
  inclusionProofAvailable: boolean;
};

export type Checkpoint = {
  id: Id;
  batchId: Id;
  type: CheckpointType;
  location: Location;
  shippingMethod: ShippingMethod | null;
  occurredAt: IsoTimestamp;
  reportedAt: IsoTimestamp;
  reportedByOrganizationName: string;
  anchor: CheckpointAnchor;
  supersededByCheckpointId: Id | null;
  supersedesCheckpointId: Id | null;
  correctionReason: string | null;
};

export type ProvenanceScoreComponent = {
  label: string;
  earned: number;
  available: number;
  explanation: string;
};

export type ProvenanceScore = {
  total: number;
  band: ProvenanceBand;
  components: ProvenanceScoreComponent[];
};

export type ParentBatchReference = {
  id: Id;
  publicReference: PublicBatchReference;
  componentType: string;
  productCategory: ProductCategory;
  originatingFacilityName: string;
  originatingFacilityCountry: CountryCode;
  resolved: boolean;
};

export type Batch = {
  id: Id;
  publicReference: PublicBatchReference;
  organizationId: Id;
  originatingFacilityId: Id;
  originatingFacilityName: string;
  productCategory: ProductCategory;
  componentType: string;
  lotNumber: string | null;
  quantity: number;
  unit: string;
  producedAt: IsoTimestamp;
  provenanceScore: ProvenanceScore;
  checkpointCount: number;
  parentBatches: ParentBatchReference[];
  externalId: string | null;
  createdAt: IsoTimestamp;
};

export type PublicBatchView = {
  publicReference: PublicBatchReference;
  productCategory: ProductCategory;
  componentType: string;
  producedAt: IsoTimestamp;
  originatingFacilityName: string;
  originatingFacilityCountry: CountryCode;
  provenanceScore: ProvenanceScore;
  checkpoints: Checkpoint[];
  approvedClaimSummaries: PublicClaimSummary[];
  lastUpdatedAt: IsoTimestamp;
};

export type PublicClaimSummary = {
  activityTypeLabel: string;
  vintageYear: number;
  approvedAt: IsoTimestamp;
};

export const EXPECTED_CHECKPOINT_SEQUENCE: Record<
  ProductCategory,
  CheckpointType[]
> = {
  electronics: [
    "production_complete",
    "departed_origin",
    "customs_export",
    "customs_import",
    "arrived_destination",
  ],
  agriculture: [],
  pharma: [],
  textiles: [],
};

export const MAXIMUM_PARENT_CHAIN_DEPTH = 10;
export const CHECKPOINT_CORRECTION_WINDOW_DAYS = 7;
