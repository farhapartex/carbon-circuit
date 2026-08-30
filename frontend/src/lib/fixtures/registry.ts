import type {
  BusinessRegistryRecord,
  RegistryVerificationOutcome,
} from "@/lib/types";

export const NAME_SIMILARITY_THRESHOLD = 0.85;

export const businessRegistry: BusinessRegistryRecord[] = [
  {
    countryCode: "TW",
    registrationNumber: "TW-28419377",
    legalName: "Formosa Precision Semiconductor Co., Ltd.",
    registeredAddress: "No. 8 Li-Hsin Road, Hsinchu Science Park, Hsinchu",
    incorporationDate: "2009-03-17T00:00:00Z",
    entityStatus: "active",
    industryCodes: ["2611", "2620"],
    sanctioned: false,
  },
  {
    countryCode: "SG",
    registrationNumber: "SG-201744820K",
    legalName: "Meridian Freight Solutions Pte. Ltd.",
    registeredAddress: "12 Tanjong Penjuru Crescent, Singapore",
    incorporationDate: "2017-11-02T00:00:00Z",
    entityStatus: "active",
    industryCodes: ["5210", "5229"],
    sanctioned: false,
  },
  {
    countryCode: "UK",
    registrationNumber: "UK-08812340",
    legalName: "Halden Insurance Group plc",
    registeredAddress: "40 Bishopsgate, London",
    incorporationDate: "2013-12-09T00:00:00Z",
    entityStatus: "active",
    industryCodes: ["6512"],
    sanctioned: false,
  },
  {
    countryCode: "DE",
    registrationNumber: "DE-HRB-114502",
    legalName: "Ostwerk Bauteile GmbH",
    registeredAddress: "Industriestrasse 44, Dresden",
    incorporationDate: "2004-06-21T00:00:00Z",
    entityStatus: "dissolved",
    industryCodes: ["2612"],
    sanctioned: false,
  },
  {
    countryCode: "VN",
    registrationNumber: "VN-0311998744",
    legalName: "Thanh Long Components JSC",
    registeredAddress: "Lot A2, Tan Thuan Export Processing Zone, Ho Chi Minh",
    incorporationDate: "2012-08-30T00:00:00Z",
    entityStatus: "active",
    industryCodes: ["2610"],
    sanctioned: true,
  },
];

const normalise = (value: string) =>
  value
    .toLowerCase()
    .replace(/[.,]/g, " ")
    .replace(
      /\b(co|ltd|limited|plc|gmbh|inc|corp|corporation|pte|jsc|company|group)\b/g,
      " ",
    )
    .replace(/\s+/g, " ")
    .trim();

const bigramsOf = (value: string) => {
  const cleaned = normalise(value).replace(/\s/g, "");
  const pairs = new Set<string>();
  for (let index = 0; index < cleaned.length - 1; index += 1) {
    pairs.add(cleaned.slice(index, index + 2));
  }
  return pairs;
};

export const nameSimilarity = (left: string, right: string) => {
  const first = bigramsOf(left);
  const second = bigramsOf(right);
  if (first.size === 0 || second.size === 0) return 0;
  let shared = 0;
  for (const pair of first) {
    if (second.has(pair)) shared += 1;
  }
  return (2 * shared) / (first.size + second.size);
};

export const verifyRegistration = async (
  countryCode: string,
  registrationNumber: string,
  declaredName: string,
): Promise<RegistryVerificationOutcome> => {
  const matchedRecord =
    businessRegistry.find(
      (record) =>
        record.countryCode === countryCode &&
        record.registrationNumber.toLowerCase() ===
          registrationNumber.trim().toLowerCase(),
    ) ?? null;

  if (!matchedRecord) {
    return {
      status: "unverified",
      matchedRecord: null,
      nameSimilarity: null,
      rejectionReason: null,
    };
  }

  const similarity = nameSimilarity(declaredName, matchedRecord.legalName);

  if (matchedRecord.entityStatus === "dissolved") {
    return {
      status: "rejected",
      matchedRecord,
      nameSimilarity: similarity,
      rejectionReason: "entity_dissolved",
    };
  }

  if (matchedRecord.sanctioned) {
    return {
      status: "rejected",
      matchedRecord,
      nameSimilarity: similarity,
      rejectionReason: "sanctions_flag",
    };
  }

  if (similarity < NAME_SIMILARITY_THRESHOLD) {
    return {
      status: "rejected",
      matchedRecord,
      nameSimilarity: similarity,
      rejectionReason: "name_mismatch",
    };
  }

  return {
    status: "verified",
    matchedRecord,
    nameSimilarity: similarity,
    rejectionReason: null,
  };
};
