import "server-only";
import { gatewayPost } from "@/lib/api/gateway";
import type { VerificationStatus } from "@/lib/status";
import type {
  OrganizationRole,
  OrganizationState,
  OrganizationType,
} from "@/lib/types/organization";

export type RegistryRejection =
  "entity_dissolved" | "sanctions_flag" | "name_mismatch";

export type OrganizationDraft = {
  name: string;
  type: OrganizationType;
  countryOfIncorporation: string;
  businessRegistrationNumber: string;
  productCategories: string[];
};

type ApiCreated = {
  organization: {
    id: string;
    name: string;
    type: OrganizationType;
    state: OrganizationState;
    verification_status: VerificationStatus;
    role: OrganizationRole;
  };
  outcome: {
    status: VerificationStatus;
    rejection: RegistryRejection | null;
    registry_match_found: boolean;
    name_similarity: string | null;
  };
};

export type CreatedOrganization = {
  id: string;
  name: string;
  state: OrganizationState;
  verificationStatus: VerificationStatus;
  rejection: RegistryRejection | null;
  registryMatchFound: boolean;
  nameSimilarity: string | null;
};

export const createOrganization = async (
  token: string,
  draft: OrganizationDraft,
  idempotencyKey: string,
): Promise<CreatedOrganization> => {
  const created = await gatewayPost<ApiCreated>(
    "/v1/organizations",
    token,
    {
      name: draft.name,
      type: draft.type,
      country_of_incorporation: draft.countryOfIncorporation,
      business_registration_number: draft.businessRegistrationNumber,
      product_categories: draft.productCategories,
    },
    idempotencyKey,
  );

  return {
    id: created.organization.id,
    name: created.organization.name,
    state: created.organization.state,
    verificationStatus: created.outcome.status,
    rejection: created.outcome.rejection,
    registryMatchFound: created.outcome.registry_match_found,
    nameSimilarity: created.outcome.name_similarity,
  };
};
