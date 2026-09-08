export type FieldRequirement = "required" | "optional" | "conditional";

export type DocumentedField = {
  name: string;
  type: string;
  requirement: FieldRequirement;
  detail: string;
};

export type DocumentedFailure = {
  status: number;
  code: string;
  when: string;
};

export type DocumentedEndpoint = {
  id: string;
  method: "GET" | "POST" | "PATCH" | "DELETE";
  path: string;
  summary: string;
  detail: string;
  idempotencyKey: boolean;
  query?: DocumentedField[];
  body?: DocumentedField[];
  requestExample?: string;
  successStatus: number;
  responseExample: string;
  failures: DocumentedFailure[];
};

export type DocumentedGroup = {
  id: string;
  title: string;
  detail: string;
  endpoints: DocumentedEndpoint[];
};
