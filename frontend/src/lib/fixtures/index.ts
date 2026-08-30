import * as batchFixtures from "@/lib/fixtures/batches";
import * as billingFixtures from "@/lib/fixtures/billing";
import * as claimFixtures from "@/lib/fixtures/claims";
import * as marketplaceFixtures from "@/lib/fixtures/marketplace";
import * as notificationFixtures from "@/lib/fixtures/notifications";
import * as organizationFixtures from "@/lib/fixtures/organizations";
import * as registryFixtures from "@/lib/fixtures/registry";
import type {
  AIReviewAvailability,
  Batch,
  Checkpoint,
  CreditPortfolio,
  CursorPaginated,
  Facility,
  Invoice,
  MfaSettings,
  MarketplaceListing,
  Notification,
  Organization,
  OrganizationInvitation,
  OrganizationUser,
  Paginated,
  PlanUsage,
  PublicBatchView,
  Retirement,
  SustainabilityClaim,
  Subscription,
  UserProfile,
  VerifierQueueEntry,
} from "@/lib/types";

const DEFAULT_PER_PAGE = 25;

const paginate = <T>(items: T[], page = 1, perPage = DEFAULT_PER_PAGE) =>
  ({
    items: items.slice((page - 1) * perPage, page * perPage),
    meta: {
      page,
      perPage,
      totalItems: items.length,
      totalPages: Math.max(1, Math.ceil(items.length / perPage)),
    },
  }) satisfies Paginated<T>;

const cursorPaginate = <T>(items: T[]) =>
  ({
    items,
    meta: { nextCursor: null, hasMore: false },
  }) satisfies CursorPaginated<T>;

export const getCurrentOrganization = async (): Promise<Organization> =>
  organizationFixtures.verifiedManufacturer;

export const getUnverifiedOrganization = async (): Promise<Organization> =>
  organizationFixtures.unverifiedAssembler;

export const getRestrictedOrganization = async (): Promise<Organization> =>
  organizationFixtures.restrictedLogisticsPartner;

export const getBuyerOrganization = async (): Promise<Organization> =>
  organizationFixtures.creditBuyer;

export const listFacilities = async (page = 1): Promise<Paginated<Facility>> =>
  paginate(organizationFixtures.facilities, page);

export const getFacility = async (
  facilityId: string,
): Promise<Facility | null> =>
  organizationFixtures.facilities.find(
    (facility) => facility.id === facilityId,
  ) ?? null;

export const listOrganizationUsers = async (): Promise<OrganizationUser[]> =>
  organizationFixtures.organizationUsers;

export const getSignedInUser = async (): Promise<UserProfile> =>
  organizationFixtures.signedInUserProfile;

export const getMfaSettings = async (): Promise<MfaSettings> =>
  organizationFixtures.mfaSettings;

export const listInvitations = async (): Promise<OrganizationInvitation[]> =>
  organizationFixtures.invitations;

export const listApiKeys = async () => organizationFixtures.apiKeys;

export const getPendingTreasuryChange = async () =>
  organizationFixtures.pendingTreasuryChange;

export const listActiveSessions = async () =>
  organizationFixtures.activeSessions;

export const listBatches = async (page = 1): Promise<Paginated<Batch>> =>
  paginate(batchFixtures.batches, page);

export const listEmptyBatches = async (): Promise<Paginated<Batch>> =>
  paginate([] as Batch[]);

export const getBatch = async (batchId: string): Promise<Batch | null> =>
  batchFixtures.batches.find((batch) => batch.id === batchId) ?? null;

export const listCheckpoints = async (
  batchId: string,
): Promise<CursorPaginated<Checkpoint>> =>
  cursorPaginate(batchFixtures.checkpointsByBatchId[batchId] ?? []);

export const getPublicBatchView = async (
  publicReference: string,
): Promise<PublicBatchView | null> =>
  batchFixtures.publicBatchViews[publicReference] ?? null;

export const listClaims = async (
  page = 1,
): Promise<Paginated<SustainabilityClaim>> =>
  paginate(claimFixtures.claims, page);

export const getClaim = async (
  claimId: string,
): Promise<SustainabilityClaim | null> =>
  claimFixtures.claims.find((claim) => claim.id === claimId) ?? null;

export const getAIReview = async (
  claimId: string,
): Promise<AIReviewAvailability> =>
  claimFixtures.aiReviewByClaimId[claimId] ?? { state: "pending" };

export const listVerifierQueue = async (): Promise<VerifierQueueEntry[]> =>
  claimFixtures.verifierQueue;

export const getCreditPortfolio = async (): Promise<CreditPortfolio> =>
  marketplaceFixtures.creditPortfolio;

export const listListings = async (
  page = 1,
): Promise<Paginated<MarketplaceListing>> =>
  paginate(marketplaceFixtures.listings, page);

export const getListing = async (
  listingId: string,
): Promise<MarketplaceListing | null> =>
  marketplaceFixtures.listings.find((listing) => listing.id === listingId) ??
  null;

export const listTrades = async () => marketplaceFixtures.trades;

export const listRetirements = async (): Promise<CursorPaginated<Retirement>> =>
  cursorPaginate(marketplaceFixtures.retirements);

export const getSubscription = async (): Promise<Subscription> =>
  billingFixtures.currentSubscription;

export const getPlanUsage = async (): Promise<PlanUsage> =>
  billingFixtures.currentUsage;

export const getPaymentMethod = async () => billingFixtures.paymentMethod;

export const listInvoices = async (): Promise<Invoice[]> =>
  billingFixtures.invoices;

export const listNotifications = async (): Promise<
  CursorPaginated<Notification>
> => cursorPaginate(notificationFixtures.notifications);

export const listFraudFlags = async () => notificationFixtures.fraudFlags;

export class FixtureRequestError extends Error {
  constructor(
    readonly code: string,
    readonly requestId: string,
  ) {
    super(`Fixture request failed with ${code}`);
    this.name = "FixtureRequestError";
  }
}

export const failingRequest = async (
  code = "DEPENDENCY_UNAVAILABLE",
): Promise<never> => {
  throw new FixtureRequestError(code, "01JQFIXTURE0000000000000000");
};

export const verifyRegistration = registryFixtures.verifyRegistration;
export const NAME_SIMILARITY_THRESHOLD =
  registryFixtures.NAME_SIMILARITY_THRESHOLD;
