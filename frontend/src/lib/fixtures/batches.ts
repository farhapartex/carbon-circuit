import { provenanceBandForScore } from "@/lib/status";
import type {
  Batch,
  Checkpoint,
  ProvenanceScore,
  ProvenanceScoreComponent,
  PublicBatchView,
} from "@/lib/types";
import {
  facilityMatched,
  verifiedManufacturer,
} from "@/lib/fixtures/organizations";

const scoreFrom = (components: ProvenanceScoreComponent[]): ProvenanceScore => {
  const total = components.reduce((sum, part) => sum + part.earned, 0);
  return { total, band: provenanceBandForScore(total), components };
};

const completeJourneyScore = scoreFrom([
  {
    label: "Checkpoint completeness",
    earned: 40,
    available: 40,
    explanation: "All 5 expected checkpoint types recorded.",
  },
  {
    label: "On-chain anchoring",
    earned: 20,
    available: 20,
    explanation: "All 5 checkpoints included in a confirmed epoch anchor.",
  },
  {
    label: "Chain depth resolution",
    earned: 15,
    available: 15,
    explanation: "Both declared parent batches resolve to registered batches.",
  },
  {
    label: "Reporting timeliness",
    earned: 15,
    available: 15,
    explanation: "Median reporting lag of 3 hours, well under 24 hours.",
  },
  {
    label: "Facility sustainability record",
    earned: 10,
    available: 10,
    explanation: "Originating facility holds an approved 2026 claim.",
  },
]);

const partialJourneyScore = scoreFrom([
  {
    label: "Checkpoint completeness",
    earned: 24,
    available: 40,
    explanation: "3 of 5 expected checkpoint types recorded.",
  },
  {
    label: "On-chain anchoring",
    earned: 13,
    available: 20,
    explanation: "2 of 3 checkpoints included in a confirmed epoch anchor.",
  },
  {
    label: "Chain depth resolution",
    earned: 15,
    available: 15,
    explanation: "This batch declares no parent batches.",
  },
  {
    label: "Reporting timeliness",
    earned: 6,
    available: 15,
    explanation: "Median reporting lag of 4 days.",
  },
  {
    label: "Facility sustainability record",
    earned: 0,
    available: 10,
    explanation: "Originating facility has no approved sustainability claim.",
  },
]);

const emptyJourneyScore = scoreFrom([
  {
    label: "Checkpoint completeness",
    earned: 0,
    available: 40,
    explanation: "No checkpoints recorded yet.",
  },
  {
    label: "On-chain anchoring",
    earned: 0,
    available: 20,
    explanation: "Nothing to anchor yet.",
  },
  {
    label: "Chain depth resolution",
    earned: 0,
    available: 15,
    explanation: "One declared parent batch has not resolved.",
  },
  {
    label: "Reporting timeliness",
    earned: 0,
    available: 15,
    explanation: "No checkpoints recorded yet.",
  },
  {
    label: "Facility sustainability record",
    earned: 0,
    available: 10,
    explanation: "Originating facility has no approved sustainability claim.",
  },
]);

export const completeJourneyCheckpoints: Checkpoint[] = [
  {
    id: "chk_1",
    batchId: "bat_wafer_lot_884",
    type: "production_complete",
    location: {
      label: "Hsinchu Fab TW-01",
      countryCode: "TW",
      coordinates: { latitude: 24.7784, longitude: 121.0033 },
    },
    shippingMethod: null,
    occurredAt: "2026-07-02T02:10:00Z",
    reportedAt: "2026-07-02T03:44:00Z",
    reportedByOrganizationName: verifiedManufacturer.name,
    anchor: {
      status: "confirmed",
      epoch: 41822,
      transactionHash:
        "0xb1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_2",
    batchId: "bat_wafer_lot_884",
    type: "departed_origin",
    location: {
      label: "Taoyuan International Airport",
      countryCode: "TW",
      coordinates: { latitude: 25.0777, longitude: 121.2328 },
    },
    shippingMethod: "air_freight_long_haul",
    occurredAt: "2026-07-03T11:25:00Z",
    reportedAt: "2026-07-03T12:02:00Z",
    reportedByOrganizationName: "Meridian Freight Solutions",
    anchor: {
      status: "confirmed",
      epoch: 41945,
      transactionHash:
        "0xc2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_3_superseded",
    batchId: "bat_wafer_lot_884",
    type: "customs_export",
    location: {
      label: "Taoyuan Customs",
      countryCode: "TW",
      coordinates: null,
    },
    shippingMethod: null,
    occurredAt: "2026-07-03T13:00:00Z",
    reportedAt: "2026-07-03T13:20:00Z",
    reportedByOrganizationName: "Meridian Freight Solutions",
    anchor: {
      status: "confirmed",
      epoch: 41946,
      transactionHash:
        "0xd3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: "chk_3_correction",
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_3_correction",
    batchId: "bat_wafer_lot_884",
    type: "customs_export",
    location: {
      label: "Taoyuan Customs, Terminal 2",
      countryCode: "TW",
      coordinates: null,
    },
    shippingMethod: null,
    occurredAt: "2026-07-03T14:35:00Z",
    reportedAt: "2026-07-05T09:10:00Z",
    reportedByOrganizationName: "Meridian Freight Solutions",
    anchor: {
      status: "confirmed",
      epoch: 42198,
      transactionHash:
        "0xe4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: "chk_3_superseded",
    correctionReason:
      "Original entry recorded the wrong terminal and an export time 95 minutes early.",
  },
  {
    id: "chk_4",
    batchId: "bat_wafer_lot_884",
    type: "customs_import",
    location: {
      label: "Penang International Airport",
      countryCode: "MY",
      coordinates: { latitude: 5.2971, longitude: 100.2769 },
    },
    shippingMethod: null,
    occurredAt: "2026-07-04T06:40:00Z",
    reportedAt: "2026-07-04T07:15:00Z",
    reportedByOrganizationName: "Northbridge Assembly Works",
    anchor: {
      status: "confirmed",
      epoch: 42055,
      transactionHash:
        "0xf5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_5",
    batchId: "bat_wafer_lot_884",
    type: "arrived_destination",
    location: {
      label: "Penang Line MY-01",
      countryCode: "MY",
      coordinates: { latitude: 5.3234, longitude: 100.2802 },
    },
    shippingMethod: "road_hgv",
    occurredAt: "2026-07-04T10:05:00Z",
    reportedAt: "2026-07-04T10:31:00Z",
    reportedByOrganizationName: "Northbridge Assembly Works",
    anchor: {
      status: "confirmed",
      epoch: 42077,
      transactionHash:
        "0xa6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
];

const partialJourneyCheckpoints: Checkpoint[] = [
  {
    id: "chk_p1",
    batchId: "bat_connector_lot_12",
    type: "production_complete",
    location: {
      label: "Penang Line MY-01",
      countryCode: "MY",
      coordinates: null,
    },
    shippingMethod: null,
    occurredAt: "2026-08-10T04:00:00Z",
    reportedAt: "2026-08-14T09:12:00Z",
    reportedByOrganizationName: "Northbridge Assembly Works",
    anchor: {
      status: "confirmed",
      epoch: 47710,
      transactionHash:
        "0xb7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_p2",
    batchId: "bat_connector_lot_12",
    type: "departed_origin",
    location: {
      label: "Port of Penang",
      countryCode: "MY",
      coordinates: null,
    },
    shippingMethod: "sea_freight_container",
    occurredAt: "2026-08-16T08:30:00Z",
    reportedAt: "2026-08-20T14:02:00Z",
    reportedByOrganizationName: "Meridian Freight Solutions",
    anchor: {
      status: "confirmed",
      epoch: 48566,
      transactionHash:
        "0xc8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_p3",
    batchId: "bat_connector_lot_12",
    type: "customs_export",
    location: { label: "Penang Customs", countryCode: "MY", coordinates: null },
    shippingMethod: null,
    occurredAt: "2026-08-27T09:00:00Z",
    reportedAt: "2026-08-28T18:40:00Z",
    reportedByOrganizationName: "Meridian Freight Solutions",
    anchor: {
      status: "provisional",
      epoch: null,
      transactionHash: null,
      inclusionProofAvailable: false,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
];

export const completeJourneyBatch: Batch = {
  id: "bat_wafer_lot_884",
  publicReference: "kQ7bTn4Wx2Ld9fPvR6sYhE",
  organizationId: verifiedManufacturer.id,
  originatingFacilityId: facilityMatched.id,
  originatingFacilityName: facilityMatched.name,
  productCategory: "electronics",
  componentType: "300mm silicon wafer, 5nm node",
  lotNumber: "WL-884",
  quantity: 5000,
  unit: "wafers",
  producedAt: "2026-07-02T02:10:00Z",
  provenanceScore: completeJourneyScore,
  checkpointCount: completeJourneyCheckpoints.length,
  parentBatches: [
    {
      id: "bat_polysilicon_221",
      publicReference: "mR3xKp8Zc1Vt5nQwJ7dBaU",
      componentType: "Electronic-grade polysilicon",
      productCategory: "electronics",
      originatingFacilityName: "Kuantan Polysilicon Works",
      originatingFacilityCountry: "MY",
      resolved: true,
    },
    {
      id: "bat_photoresist_57",
      publicReference: "vN9jHs2Yq6Fw0zLxG4mCtP",
      componentType: "EUV photoresist, batch 57",
      productCategory: "electronics",
      originatingFacilityName: "Kawasaki Speciality Chemicals",
      originatingFacilityCountry: "JP",
      resolved: true,
    },
  ],
  externalId: "ERP-WL-884",
  createdAt: "2026-07-02T03:44:00Z",
};

export const partialJourneyBatch: Batch = {
  id: "bat_connector_lot_12",
  publicReference: "tY5cWm1Nb8Rj3kFxQ2pDzA",
  organizationId: "org_unverified_assembler",
  originatingFacilityId: "fac_my_01",
  originatingFacilityName: "Penang Line MY-01",
  productCategory: "electronics",
  componentType: "Board-to-board connector, 40-pin",
  lotNumber: "CN-12",
  quantity: 120000,
  unit: "units",
  producedAt: "2026-08-10T04:00:00Z",
  provenanceScore: partialJourneyScore,
  checkpointCount: partialJourneyCheckpoints.length,
  parentBatches: [],
  externalId: null,
  createdAt: "2026-08-14T09:12:00Z",
};

export const unstartedBatch: Batch = {
  id: "bat_cell_lot_003",
  publicReference: "hJ6dQz9Pk4Xr7vBnM3tGwS",
  organizationId: verifiedManufacturer.id,
  originatingFacilityId: facilityMatched.id,
  originatingFacilityName: facilityMatched.name,
  productCategory: "electronics",
  componentType: "Lithium-ion cell, 21700 format",
  lotNumber: "LC-003",
  quantity: 24000,
  unit: "cells",
  producedAt: "2026-08-29T01:00:00Z",
  provenanceScore: emptyJourneyScore,
  checkpointCount: 0,
  parentBatches: [
    {
      id: "bat_unknown_cathode",
      publicReference: "",
      componentType: "Cathode active material",
      productCategory: "electronics",
      originatingFacilityName: "Unregistered supplier",
      originatingFacilityCountry: "CN",
      resolved: false,
    },
  ],
  externalId: "ERP-LC-003",
  createdAt: "2026-08-29T01:30:00Z",
};

export const batches: Batch[] = [
  completeJourneyBatch,
  partialJourneyBatch,
  unstartedBatch,
];

export const checkpointsByBatchId: Record<string, Checkpoint[]> = {
  [completeJourneyBatch.id]: completeJourneyCheckpoints,
  [partialJourneyBatch.id]: partialJourneyCheckpoints,
  [unstartedBatch.id]: [],
};

export const publicBatchViews: Record<string, PublicBatchView> = {
  [completeJourneyBatch.publicReference]: {
    publicReference: completeJourneyBatch.publicReference,
    productCategory: completeJourneyBatch.productCategory,
    componentType: completeJourneyBatch.componentType,
    producedAt: completeJourneyBatch.producedAt,
    originatingFacilityName: facilityMatched.name,
    originatingFacilityCountry: "TW",
    provenanceScore: completeJourneyScore,
    checkpoints: completeJourneyCheckpoints,
    approvedClaimSummaries: [
      {
        activityTypeLabel: "Renewable energy",
        vintageYear: 2026,
        approvedAt: "2026-08-12T13:20:00Z",
      },
    ],
    lastUpdatedAt: "2026-08-29T08:00:00Z",
  },
  [unstartedBatch.publicReference]: {
    publicReference: unstartedBatch.publicReference,
    productCategory: unstartedBatch.productCategory,
    componentType: unstartedBatch.componentType,
    producedAt: unstartedBatch.producedAt,
    originatingFacilityName: facilityMatched.name,
    originatingFacilityCountry: "TW",
    provenanceScore: emptyJourneyScore,
    checkpoints: [],
    approvedClaimSummaries: [],
    lastUpdatedAt: "2026-08-29T01:30:00Z",
  },
};

const polysiliconScore = scoreFrom([
  {
    label: "Checkpoint completeness",
    earned: 40,
    available: 40,
    explanation: "All 5 expected checkpoint types recorded.",
  },
  {
    label: "On-chain anchoring",
    earned: 20,
    available: 20,
    explanation: "All 5 checkpoints included in a confirmed epoch anchor.",
  },
  {
    label: "Chain depth resolution",
    earned: 15,
    available: 15,
    explanation: "This batch declares no parent batches.",
  },
  {
    label: "Reporting timeliness",
    earned: 12,
    available: 15,
    explanation: "Median reporting lag of 9 hours.",
  },
  {
    label: "Facility sustainability record",
    earned: 10,
    available: 10,
    explanation: "Originating facility holds an approved 2026 claim.",
  },
]);

const photoresistScore = scoreFrom([
  {
    label: "Checkpoint completeness",
    earned: 24,
    available: 40,
    explanation: "3 of 5 expected checkpoint types recorded.",
  },
  {
    label: "On-chain anchoring",
    earned: 20,
    available: 20,
    explanation: "All 3 checkpoints included in a confirmed epoch anchor.",
  },
  {
    label: "Chain depth resolution",
    earned: 15,
    available: 15,
    explanation: "This batch declares no parent batches.",
  },
  {
    label: "Reporting timeliness",
    earned: 15,
    available: 15,
    explanation: "Median reporting lag of 2 hours, well under 24 hours.",
  },
  {
    label: "Facility sustainability record",
    earned: 0,
    available: 10,
    explanation: "Originating facility has no approved sustainability claim.",
  },
]);

const polysiliconCheckpoints: Checkpoint[] = [
  {
    id: "chk_poly_1",
    batchId: "bat_polysilicon_221",
    type: "production_complete",
    location: {
      label: "Kuantan Polysilicon Works",
      countryCode: "MY",
      coordinates: { latitude: 3.8077, longitude: 103.326 },
    },
    shippingMethod: null,
    occurredAt: "2026-05-18T06:30:00Z",
    reportedAt: "2026-05-18T15:10:00Z",
    reportedByOrganizationName: "Kuantan Polysilicon Works",
    anchor: {
      status: "confirmed",
      epoch: 38104,
      transactionHash:
        "0xa1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_poly_2",
    batchId: "bat_polysilicon_221",
    type: "departed_origin",
    location: {
      label: "Port of Kuantan",
      countryCode: "MY",
      coordinates: { latitude: 3.9741, longitude: 103.4292 },
    },
    shippingMethod: "sea_freight_bulk",
    occurredAt: "2026-05-21T09:00:00Z",
    reportedAt: "2026-05-21T18:40:00Z",
    reportedByOrganizationName: "Straits Bulk Logistics",
    anchor: {
      status: "confirmed",
      epoch: 38311,
      transactionHash:
        "0xb2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_poly_3",
    batchId: "bat_polysilicon_221",
    type: "customs_export",
    location: {
      label: "Kuantan Customs",
      countryCode: "MY",
      coordinates: null,
    },
    shippingMethod: null,
    occurredAt: "2026-05-21T14:20:00Z",
    reportedAt: "2026-05-22T01:05:00Z",
    reportedByOrganizationName: "Straits Bulk Logistics",
    anchor: {
      status: "confirmed",
      epoch: 38344,
      transactionHash:
        "0xc3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_poly_4",
    batchId: "bat_polysilicon_221",
    type: "customs_import",
    location: {
      label: "Keelung Customs",
      countryCode: "TW",
      coordinates: null,
    },
    shippingMethod: null,
    occurredAt: "2026-06-02T08:15:00Z",
    reportedAt: "2026-06-02T16:30:00Z",
    reportedByOrganizationName: "Straits Bulk Logistics",
    anchor: {
      status: "confirmed",
      epoch: 39102,
      transactionHash:
        "0xd4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_poly_5",
    batchId: "bat_polysilicon_221",
    type: "arrived_destination",
    location: {
      label: "Hsinchu Fab TW-01",
      countryCode: "TW",
      coordinates: { latitude: 24.7784, longitude: 121.0033 },
    },
    shippingMethod: "road_hgv",
    occurredAt: "2026-06-03T05:40:00Z",
    reportedAt: "2026-06-03T09:55:00Z",
    reportedByOrganizationName: verifiedManufacturer.name,
    anchor: {
      status: "confirmed",
      epoch: 39188,
      transactionHash:
        "0xe5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
];

const photoresistCheckpoints: Checkpoint[] = [
  {
    id: "chk_resist_1",
    batchId: "bat_photoresist_57",
    type: "production_complete",
    location: {
      label: "Kawasaki Speciality Chemicals",
      countryCode: "JP",
      coordinates: { latitude: 35.5308, longitude: 139.7029 },
    },
    shippingMethod: null,
    occurredAt: "2026-06-05T23:10:00Z",
    reportedAt: "2026-06-06T01:20:00Z",
    reportedByOrganizationName: "Kawasaki Speciality Chemicals",
    anchor: {
      status: "confirmed",
      epoch: 39420,
      transactionHash:
        "0xf60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_resist_2",
    batchId: "bat_photoresist_57",
    type: "departed_origin",
    location: {
      label: "Tokyo Haneda Airport",
      countryCode: "JP",
      coordinates: { latitude: 35.5494, longitude: 139.7798 },
    },
    shippingMethod: "air_freight_short_haul",
    occurredAt: "2026-06-08T13:45:00Z",
    reportedAt: "2026-06-08T15:30:00Z",
    reportedByOrganizationName: "Nihon Express Cargo",
    anchor: {
      status: "confirmed",
      epoch: 39655,
      transactionHash:
        "0x0718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f6",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
  {
    id: "chk_resist_3",
    batchId: "bat_photoresist_57",
    type: "customs_export",
    location: { label: "Haneda Customs", countryCode: "JP", coordinates: null },
    shippingMethod: null,
    occurredAt: "2026-06-08T18:05:00Z",
    reportedAt: "2026-06-08T20:15:00Z",
    reportedByOrganizationName: "Nihon Express Cargo",
    anchor: {
      status: "confirmed",
      epoch: 39681,
      transactionHash:
        "0x18293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f607",
      inclusionProofAvailable: true,
    },
    supersededByCheckpointId: null,
    supersedesCheckpointId: null,
    correctionReason: null,
  },
];

export const componentBatchViews: Record<string, PublicBatchView> = {
  bat_polysilicon_221: {
    publicReference: "mR3xKp8Zc1Vt5nQwJ7dBaU",
    productCategory: "electronics",
    componentType: "Electronic-grade polysilicon",
    producedAt: "2026-05-18T06:30:00Z",
    originatingFacilityName: "Kuantan Polysilicon Works",
    originatingFacilityCountry: "MY",
    provenanceScore: polysiliconScore,
    checkpoints: polysiliconCheckpoints,
    approvedClaimSummaries: [
      {
        activityTypeLabel: "Renewable energy",
        vintageYear: 2026,
        approvedAt: "2026-04-30T08:00:00Z",
      },
    ],
    lastUpdatedAt: "2026-06-03T09:55:00Z",
  },
  bat_photoresist_57: {
    publicReference: "vN9jHs2Yq6Fw0zLxG4mCtP",
    productCategory: "electronics",
    componentType: "EUV photoresist, batch 57",
    producedAt: "2026-06-05T23:10:00Z",
    originatingFacilityName: "Kawasaki Speciality Chemicals",
    originatingFacilityCountry: "JP",
    provenanceScore: photoresistScore,
    checkpoints: photoresistCheckpoints,
    approvedClaimSummaries: [],
    lastUpdatedAt: "2026-06-08T20:15:00Z",
  },
};
