import "server-only";
import { serverConfig } from "@/lib/config/server";
import { provenanceBandForScore } from "@/lib/status";
import type {
  AnchorStatus,
  Checkpoint,
  CheckpointType,
  ProductCategory,
  PublicBatchView,
  ShippingMethod,
} from "@/lib/types";

type ApiPublicCheckpoint = {
  type: CheckpointType;
  location_label: string;
  country_code: string;
  shipping_method: ShippingMethod | null;
  occurred_at: string;
  anchor_status: AnchorStatus;
  anchor_epoch: number | null;
  anchor_transaction_hash: string | null;
  superseded: boolean;
};

type ApiPublicBatch = {
  public_reference: string;
  product_category: ProductCategory;
  component_type: string;
  originating_facility_name: string;
  originating_facility_country: string;
  produced_at: string;
  provenance_score: {
    total: number;
    components: {
      label: string;
      earned: number;
      available: number;
      explanation: string;
    }[];
  };
  checkpoints: ApiPublicCheckpoint[];
  last_updated_at: string;
};

const toCheckpoint = (
  checkpoint: ApiPublicCheckpoint,
  index: number,
): Checkpoint => ({
  id: `public-${index}`,
  batchId: "",
  type: checkpoint.type,
  location: {
    label: checkpoint.location_label,
    countryCode:
      checkpoint.country_code as Checkpoint["location"]["countryCode"],
    coordinates: null,
  },
  shippingMethod: checkpoint.shipping_method,
  occurredAt: checkpoint.occurred_at,
  reportedAt: checkpoint.occurred_at,
  reportedByOrganizationName: "",
  anchor: {
    status: checkpoint.anchor_status,
    epoch: checkpoint.anchor_epoch,
    transactionHash:
      checkpoint.anchor_transaction_hash as Checkpoint["anchor"]["transactionHash"],
    inclusionProofAvailable: false,
  },
  supersedesCheckpointId: null,
  supersededByCheckpointId: checkpoint.superseded ? "superseded" : null,
  correctionReason: null,
});

export const fetchPublicBatch = async (
  publicReference: string,
): Promise<PublicBatchView | null> => {
  const response = await fetch(
    new URL(`/v1/track/${publicReference}`, serverConfig.apiGatewayUrl),
    { headers: { Accept: "application/json" }, next: { revalidate: 30 } },
  );

  if (response.status === 404) return null;

  if (!response.ok) {
    throw new Error(`Public batch lookup failed with ${response.status}`);
  }

  const { data } = (await response.json()) as { data: ApiPublicBatch };

  return {
    publicReference: data.public_reference,
    productCategory: data.product_category,
    componentType: data.component_type,
    producedAt: data.produced_at,
    originatingFacilityName: data.originating_facility_name,
    originatingFacilityCountry:
      data.originating_facility_country as PublicBatchView["originatingFacilityCountry"],
    provenanceScore: {
      total: data.provenance_score.total,
      band: provenanceBandForScore(data.provenance_score.total),
      components: data.provenance_score.components,
    },
    checkpoints: data.checkpoints.map(toCheckpoint),
    approvedClaimSummaries: [],
    lastUpdatedAt: data.last_updated_at,
  };
};
