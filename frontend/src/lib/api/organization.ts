import "server-only";
import { gatewayGet } from "@/lib/api/gateway";
import type { RegistryRejection } from "@/lib/api/organizations";
import type { VerificationStatus } from "@/lib/status";
import type { EthereumAddress } from "@/lib/types/common";
import type {
  OrganizationRole,
  OrganizationState,
  OrganizationType,
} from "@/lib/types/organization";

type ApiOrganizationDetail = {
  id: string;
  name: string;
  type: OrganizationType;
  state: OrganizationState;
  verification_status: VerificationStatus;
  country_of_incorporation: string;
  business_registration_number: string;
  product_categories: string[];
  treasury_designated: boolean;
  treasury_address: string | null;
  role: OrganizationRole;
  created_at: string;
  outcome: {
    status: VerificationStatus;
    rejection: RegistryRejection | null;
    registry_match_found: boolean;
    name_similarity: string | null;
    registered_legal_name: string | null;
  };
};

export type OrganizationDetail = {
  id: string;
  name: string;
  type: OrganizationType;
  state: OrganizationState;
  verificationStatus: VerificationStatus;
  countryOfIncorporation: string;
  businessRegistrationNumber: string;
  productCategories: string[];
  treasuryDesignated: boolean;
  treasuryAddress: EthereumAddress | null;
  role: OrganizationRole;
  createdAt: string;
  outcome: {
    status: VerificationStatus;
    rejection: RegistryRejection | null;
    registryMatchFound: boolean;
    nameSimilarity: number | null;
    registeredLegalName: string | null;
  };
};

export const fetchCurrentOrganization = async (
  token: string,
): Promise<OrganizationDetail> => {
  const detail = await gatewayGet<ApiOrganizationDetail>(
    "/v1/organizations/current",
    token,
  );

  return {
    id: detail.id,
    name: detail.name,
    type: detail.type,
    state: detail.state,
    verificationStatus: detail.verification_status,
    countryOfIncorporation: detail.country_of_incorporation,
    businessRegistrationNumber: detail.business_registration_number,
    productCategories: detail.product_categories,
    treasuryDesignated: detail.treasury_designated,
    treasuryAddress: detail.treasury_address as EthereumAddress | null,
    role: detail.role,
    createdAt: detail.created_at,
    outcome: {
      status: detail.outcome.status,
      rejection: detail.outcome.rejection,
      registryMatchFound: detail.outcome.registry_match_found,
      nameSimilarity:
        detail.outcome.name_similarity === null
          ? null
          : Number(detail.outcome.name_similarity),
      registeredLegalName: detail.outcome.registered_legal_name,
    },
  };
};
