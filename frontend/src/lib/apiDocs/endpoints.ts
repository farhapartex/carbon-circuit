import type { DocumentedGroup } from "@/lib/apiDocs/types";

const batchResponseExample = `{
  "data": {
    "batch": {
      "id": "01a063f4-fa3f-73c8-880d-c8cac3a1efae",
      "originating_facility_id": "01a0619c-3d21-7b40-9f8e-2c7f5a1b4d02",
      "originating_facility_name": "Hsinchu Fab TW-01",
      "public_reference": "1Nevv4JVPZUNIQSRo6cWhC",
      "product_category": "electronics",
      "component_type": "300mm silicon wafer, 5nm node",
      "lot_number": "WL-884",
      "quantity": "5000.000000",
      "unit": "wafers",
      "produced_at": "2026-08-28T00:00:00Z",
      "external_id": "ERP-WL-884",
      "checkpoint_count": 0,
      "provenance_score": {
        "total": 0,
        "components": [
          {
            "label": "Checkpoint completeness",
            "earned": 0,
            "available": 40,
            "explanation": "No checkpoints recorded yet."
          }
        ]
      },
      "created_at": "2026-09-07T06:12:04Z"
    },
    "parents": []
  }
}`;

const checkpointResponseExample = `{
  "data": {
    "checkpoint": {
      "id": "01a06411-9ddd-759c-84ee-9767632d9de3",
      "batch_id": "01a063f4-fa3f-73c8-880d-c8cac3a1efae",
      "type": "departed_origin",
      "location_label": "Taoyuan International Airport",
      "country_code": "TW",
      "latitude": null,
      "longitude": null,
      "shipping_method": "air_freight_long_haul",
      "occurred_at": "2026-09-02T15:12:00Z",
      "reported_at": "2026-09-02T21:13:26Z",
      "reported_by_organization_name": "Formosa Precision Semiconductor Co., Ltd.",
      "anchor_status": "unanchored",
      "anchor_epoch": null,
      "anchor_transaction_hash": null,
      "inclusion_proof_available": false,
      "supersedes_checkpoint_id": null,
      "superseded_by_checkpoint_id": null,
      "correction_reason": null
    },
    "provenance_score": { "total": 38, "components": [] }
  }
}`;

export const apiDocGroups: DocumentedGroup[] = [
  {
    id: "batches",
    title: "Batches",
    detail:
      "A batch records a produced quantity of a component. Its product category is fixed for the batch's whole lifetime, and its public reference is generated on creation.",
    endpoints: [
      {
        id: "create-batch",
        method: "POST",
        path: "/v1/batches",
        summary: "Register a batch",
        detail:
          "Creates a batch against a facility your organization has already registered. Facilities are never auto-created; an unknown facility is refused.",
        idempotencyKey: true,
        body: [
          {
            name: "originating_facility_id",
            type: "uuid",
            requirement: "required",
            detail: "A facility belonging to your organization.",
          },
          {
            name: "product_category",
            type: "electronics | agriculture | pharma | textiles",
            requirement: "required",
            detail: "Permanent for this batch. It cannot be changed later.",
          },
          {
            name: "component_type",
            type: "string, up to 160",
            requirement: "required",
            detail: "What the batch contains.",
          },
          {
            name: "quantity",
            type: "decimal string, up to 6 places",
            requirement: "required",
            detail:
              "Sent as a string so no precision is lost in transit. Must be greater than zero.",
          },
          {
            name: "unit",
            type: "string, up to 32",
            requirement: "required",
            detail: "The unit the quantity is counted in, such as wafers.",
          },
          {
            name: "produced_at",
            type: "RFC3339 timestamp",
            requirement: "required",
            detail: "Cannot be in the future.",
          },
          {
            name: "external_id",
            type: "string, up to 128",
            requirement: "conditional",
            detail:
              "Required when authenticating with an API key. This is the record's identifier in your own system, and it is what makes a replay safe.",
          },
          {
            name: "lot_number",
            type: "string, up to 64",
            requirement: "optional",
            detail: "Your lot or batch number, if you use one.",
          },
          {
            name: "parent_references",
            type: "array of 22-character references",
            requirement: "optional",
            detail:
              "Public references of component batches this one was made from. A reference that matches no registered batch is still recorded, and counts against the chain depth score.",
          },
        ],
        requestExample: `{
  "originating_facility_id": "01a0619c-3d21-7b40-9f8e-2c7f5a1b4d02",
  "product_category": "electronics",
  "component_type": "300mm silicon wafer, 5nm node",
  "lot_number": "WL-884",
  "quantity": "5000",
  "unit": "wafers",
  "produced_at": "2026-08-28T00:00:00Z",
  "external_id": "ERP-WL-884",
  "parent_references": []
}`,
        successStatus: 201,
        responseExample: batchResponseExample,
        failures: [
          {
            status: 422,
            code: "VALIDATION_ERROR",
            when: "A field is missing or malformed. With an API key, a missing external_id reports code REQUIRED_FOR_API_SUBMISSION against that field.",
          },
          {
            status: 409,
            code: "CONFLICT",
            when: "This external_id was already used by your organization.",
          },
          {
            status: 409,
            code: "CONFLICT",
            when: "The originating facility is not registered to your organization.",
          },
          {
            status: 403,
            code: "FORBIDDEN",
            when: "Your organization is read only, or its type cannot produce batches.",
          },
          {
            status: 422,
            code: "IDEMPOTENCY_KEY_REQUIRED",
            when: "The Idempotency-Key header was absent.",
          },
        ],
      },
      {
        id: "list-batches",
        method: "GET",
        path: "/v1/batches",
        summary: "List your batches",
        detail:
          "Returns batches your organization owns, newest first, using an opaque cursor. Batches belonging to other organizations are never returned.",
        idempotencyKey: false,
        query: [
          {
            name: "after",
            type: "opaque cursor",
            requirement: "optional",
            detail: "The cursor returned by the previous page.",
          },
          {
            name: "per_page",
            type: "integer, up to 100",
            requirement: "optional",
            detail: "Defaults to 25.",
          },
        ],
        successStatus: 200,
        responseExample: `{
  "data": {
    "batches": [ { "id": "01a063f4-...", "component_type": "300mm silicon wafer, 5nm node" } ],
    "cursor": "01a063f4-fa3f-73c8-880d-c8cac3a1efae",
    "has_more": false
  }
}`,
        failures: [
          {
            status: 401,
            code: "UNAUTHENTICATED",
            when: "The credential was missing, malformed, or revoked.",
          },
        ],
      },
      {
        id: "get-batch",
        method: "GET",
        path: "/v1/batches/{batchId}",
        summary: "Read one batch",
        detail:
          "Returns a batch you own, together with the component batches it declares. A batch you do not own returns 404 rather than 403, so the endpoint never confirms that someone else's batch exists.",
        idempotencyKey: false,
        successStatus: 200,
        responseExample: batchResponseExample,
        failures: [
          {
            status: 404,
            code: "RESOURCE_NOT_FOUND",
            when: "No batch with that id belongs to your organization.",
          },
        ],
      },
    ],
  },
  {
    id: "checkpoints",
    title: "Checkpoints",
    detail:
      "A checkpoint records where a batch was and when. Checkpoints are append-only: a mistake is corrected by filing a superseding entry, never by editing or deleting the original.",
    endpoints: [
      {
        id: "log-checkpoint",
        method: "POST",
        path: "/v1/batches/{batchId}/checkpoints",
        summary: "Log a checkpoint",
        detail:
          "Only the organization that owns the batch may log checkpoints against it. Logging one recomputes the batch's Provenance Score, which is returned alongside the checkpoint.",
        idempotencyKey: true,
        body: [
          {
            name: "type",
            type: "production_complete | departed_origin | customs_export | customs_import | arrived_destination",
            requirement: "required",
            detail: "The five expected types for the electronics category.",
          },
          {
            name: "location_label",
            type: "string, up to 120",
            requirement: "required",
            detail:
              "Where the event happened, as a shipping document would name it.",
          },
          {
            name: "country_code",
            type: "ISO 3166-1 alpha-2",
            requirement: "required",
            detail: "Two letters, such as TW.",
          },
          {
            name: "occurred_at",
            type: "RFC3339 timestamp",
            requirement: "required",
            detail:
              "The event time, not the reporting time. It cannot precede the batch's production date or sit in the future.",
          },
          {
            name: "shipping_method",
            type: "string",
            requirement: "conditional",
            detail:
              "Required for every type except production_complete, because it carries the emissions factor for the leg.",
          },
          {
            name: "external_id",
            type: "string, up to 128",
            requirement: "conditional",
            detail: "Required when authenticating with an API key.",
          },
          {
            name: "corrects_checkpoint_id",
            type: "uuid",
            requirement: "optional",
            detail:
              "Files this checkpoint as superseding an earlier one. Allowed within seven days of the original, and only once per original.",
          },
          {
            name: "correction_reason",
            type: "string, up to 500",
            requirement: "conditional",
            detail: "Required whenever corrects_checkpoint_id is given.",
          },
          {
            name: "latitude, longitude",
            type: "decimal strings",
            requirement: "optional",
            detail: "Give both or neither.",
          },
        ],
        requestExample: `{
  "type": "departed_origin",
  "location_label": "Taoyuan International Airport",
  "country_code": "TW",
  "shipping_method": "air_freight_long_haul",
  "occurred_at": "2026-09-02T15:12:00Z",
  "external_id": "WMS-88213"
}`,
        successStatus: 201,
        responseExample: checkpointResponseExample,
        failures: [
          {
            status: 403,
            code: "FORBIDDEN",
            when: "The batch belongs to another organization.",
          },
          {
            status: 422,
            code: "VALIDATION_ERROR",
            when: "A movement was logged without a shipping method, the event time precedes production, or it sits in the future.",
          },
          {
            status: 409,
            code: "CONFLICT",
            when: "The checkpoint being corrected was already corrected, or the seven day window has closed.",
          },
        ],
      },
      {
        id: "list-checkpoints",
        method: "GET",
        path: "/v1/batches/{batchId}/checkpoints",
        summary: "Read a batch's history",
        detail:
          "Returns the full checkpoint history in event order, including superseded entries so the correction trail stays visible.",
        idempotencyKey: false,
        query: [
          {
            name: "after",
            type: "opaque cursor",
            requirement: "optional",
            detail: "The cursor returned by the previous page.",
          },
          {
            name: "per_page",
            type: "integer, up to 100",
            requirement: "optional",
            detail: "Defaults to 25.",
          },
        ],
        successStatus: 200,
        responseExample: `{
  "data": {
    "checkpoints": [ { "id": "01a06411-...", "type": "departed_origin" } ],
    "cursor": "01a06411-9ddd-759c-84ee-9767632d9de3",
    "has_more": false
  }
}`,
        failures: [
          {
            status: 404,
            code: "RESOURCE_NOT_FOUND",
            when: "No batch with that id is visible to your organization.",
          },
        ],
      },
    ],
  },
];
