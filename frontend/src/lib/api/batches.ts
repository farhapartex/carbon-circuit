import "server-only";
import { gatewayGet, gatewayPost } from "@/lib/api/gateway";
import type {
  AnchorStatus,
  Checkpoint,
  CheckpointType,
  ProductCategory,
  ProvenanceScore,
  ProvenanceScoreComponent,
  ShippingMethod,
} from "@/lib/types";
import { provenanceBandForScore } from "@/lib/status";

type ApiScoreComponent = {
  label: string;
  earned: number;
  available: number;
  explanation: string;
};

type ApiScore = {
  total: number;
  components: ApiScoreComponent[];
};

type ApiBatch = {
  id: string;
  originating_facility_id: string;
  originating_facility_name: string;
  public_reference: string;
  product_category: ProductCategory;
  component_type: string;
  lot_number: string | null;
  quantity: string;
  unit: string;
  produced_at: string;
  external_id: string | null;
  checkpoint_count: number;
  provenance_score: ApiScore;
  created_at: string;
};

type ApiParent = {
  declared_reference: string;
  resolved: boolean;
  id: string | null;
  component_type: string | null;
  product_category: ProductCategory | null;
  originating_facility_name: string | null;
};

type ApiCheckpoint = {
  id: string;
  batch_id: string;
  type: CheckpointType;
  location_label: string;
  country_code: string;
  latitude: string | null;
  longitude: string | null;
  shipping_method: ShippingMethod | null;
  occurred_at: string;
  reported_at: string;
  reported_by_organization_name: string;
  anchor_status: AnchorStatus;
  anchor_epoch: number | null;
  anchor_transaction_hash: string | null;
  inclusion_proof_available: boolean;
  supersedes_checkpoint_id: string | null;
  superseded_by_checkpoint_id: string | null;
  correction_reason: string | null;
};

export type BatchRecord = {
  id: string;
  originatingFacilityId: string;
  originatingFacilityName: string;
  publicReference: string;
  productCategory: ProductCategory;
  componentType: string;
  lotNumber: string | null;
  quantity: string;
  unit: string;
  producedAt: string;
  externalId: string | null;
  checkpointCount: number;
  provenanceScore: ProvenanceScore;
  createdAt: string;
};

export type ParentBatchRecord = {
  declaredReference: string;
  resolved: boolean;
  id: string | null;
  componentType: string | null;
  productCategory: ProductCategory | null;
  originatingFacilityName: string | null;
};

export type CheckpointRecord = Checkpoint;

export type BatchPage = {
  items: BatchRecord[];
  cursor: string | null;
  hasMore: boolean;
};

export type BatchDetail = {
  batch: BatchRecord;
  parents: ParentBatchRecord[];
};

export type ComponentBatchDetail = {
  batch: BatchRecord;
  checkpoints: CheckpointRecord[];
};

export type BatchDraft = {
  originatingFacilityId: string;
  productCategory: ProductCategory;
  componentType: string;
  lotNumber: string;
  quantity: string;
  unit: string;
  producedAt: string;
  externalId: string;
  parentReferences: string[];
};

export type CheckpointDraft = {
  type: CheckpointType;
  locationLabel: string;
  countryCode: string;
  shippingMethod: string;
  occurredAt: string;
};

const toScoreComponent = (
  component: ApiScoreComponent,
): ProvenanceScoreComponent => ({
  label: component.label,
  earned: component.earned,
  available: component.available,
  explanation: component.explanation,
});

const toScore = (scored: ApiScore): ProvenanceScore => ({
  total: scored.total,
  band: provenanceBandForScore(scored.total),
  components: scored.components.map(toScoreComponent),
});

const toBatch = (batch: ApiBatch): BatchRecord => ({
  id: batch.id,
  originatingFacilityId: batch.originating_facility_id,
  originatingFacilityName: batch.originating_facility_name,
  publicReference: batch.public_reference,
  productCategory: batch.product_category,
  componentType: batch.component_type,
  lotNumber: batch.lot_number,
  quantity: batch.quantity,
  unit: batch.unit,
  producedAt: batch.produced_at,
  externalId: batch.external_id,
  checkpointCount: batch.checkpoint_count,
  provenanceScore: toScore(batch.provenance_score),
  createdAt: batch.created_at,
});

const toParent = (parent: ApiParent): ParentBatchRecord => ({
  declaredReference: parent.declared_reference,
  resolved: parent.resolved,
  id: parent.id,
  componentType: parent.component_type,
  productCategory: parent.product_category,
  originatingFacilityName: parent.originating_facility_name,
});

const toCoordinates = (checkpoint: ApiCheckpoint) =>
  checkpoint.latitude === null || checkpoint.longitude === null
    ? null
    : {
        latitude: Number(checkpoint.latitude),
        longitude: Number(checkpoint.longitude),
      };

const toCheckpoint = (checkpoint: ApiCheckpoint): Checkpoint => ({
  id: checkpoint.id,
  batchId: checkpoint.batch_id,
  type: checkpoint.type,
  location: {
    label: checkpoint.location_label,
    countryCode:
      checkpoint.country_code as Checkpoint["location"]["countryCode"],
    coordinates: toCoordinates(checkpoint),
  },
  shippingMethod: checkpoint.shipping_method,
  occurredAt: checkpoint.occurred_at,
  reportedAt: checkpoint.reported_at,
  reportedByOrganizationName: checkpoint.reported_by_organization_name,
  anchor: {
    status: checkpoint.anchor_status,
    epoch: checkpoint.anchor_epoch,
    transactionHash:
      checkpoint.anchor_transaction_hash as Checkpoint["anchor"]["transactionHash"],
    inclusionProofAvailable: checkpoint.inclusion_proof_available,
  },
  supersedesCheckpointId: checkpoint.supersedes_checkpoint_id,
  supersededByCheckpointId: checkpoint.superseded_by_checkpoint_id,
  correctionReason: checkpoint.correction_reason,
});

export const fetchBatches = async (
  token: string,
  after?: string,
): Promise<BatchPage> => {
  const query = after ? `?after=${encodeURIComponent(after)}` : "";
  const page = await gatewayGet<{
    batches: ApiBatch[];
    cursor: string | null;
    has_more: boolean;
  }>(`/v1/batches${query}`, token);

  return {
    items: page.batches.map(toBatch),
    cursor: page.cursor,
    hasMore: page.has_more,
  };
};

export const fetchBatch = async (
  token: string,
  batchId: string,
): Promise<BatchDetail> => {
  const detail = await gatewayGet<{ batch: ApiBatch; parents: ApiParent[] }>(
    `/v1/batches/${batchId}`,
    token,
  );

  return {
    batch: toBatch(detail.batch),
    parents: detail.parents.map(toParent),
  };
};

export const fetchCheckpoints = async (
  token: string,
  batchId: string,
): Promise<CheckpointRecord[]> => {
  const page = await gatewayGet<{ checkpoints: ApiCheckpoint[] }>(
    `/v1/batches/${batchId}/checkpoints`,
    token,
  );

  return page.checkpoints.map(toCheckpoint);
};

export const fetchComponentBatch = async (
  token: string,
  batchId: string,
  componentBatchId: string,
): Promise<ComponentBatchDetail> => {
  const detail = await gatewayGet<{
    batch: ApiBatch;
    checkpoints: ApiCheckpoint[];
  }>(`/v1/batches/${batchId}/components/${componentBatchId}`, token);

  return {
    batch: toBatch(detail.batch),
    checkpoints: detail.checkpoints.map(toCheckpoint),
  };
};

export const createBatch = async (
  token: string,
  draft: BatchDraft,
  idempotencyKey: string,
): Promise<BatchDetail> => {
  const created = await gatewayPost<{ batch: ApiBatch; parents: ApiParent[] }>(
    "/v1/batches",
    token,
    {
      originating_facility_id: draft.originatingFacilityId,
      product_category: draft.productCategory,
      component_type: draft.componentType,
      lot_number: draft.lotNumber,
      quantity: draft.quantity,
      unit: draft.unit,
      produced_at: draft.producedAt,
      external_id: draft.externalId,
      parent_references: draft.parentReferences,
    },
    idempotencyKey,
  );

  return {
    batch: toBatch(created.batch),
    parents: created.parents.map(toParent),
  };
};

export const logCheckpoint = async (
  token: string,
  batchId: string,
  draft: CheckpointDraft,
  idempotencyKey: string,
): Promise<CheckpointRecord> => {
  const logged = await gatewayPost<{ checkpoint: ApiCheckpoint }>(
    `/v1/batches/${batchId}/checkpoints`,
    token,
    {
      type: draft.type,
      location_label: draft.locationLabel,
      country_code: draft.countryCode,
      shipping_method: draft.shippingMethod,
      occurred_at: draft.occurredAt,
    },
    idempotencyKey,
  );

  return toCheckpoint(logged.checkpoint);
};
