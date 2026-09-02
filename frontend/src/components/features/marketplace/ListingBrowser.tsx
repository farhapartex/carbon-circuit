"use client";

import { Store } from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";
import { ListingCard } from "@/components/features/marketplace/ListingCard";
import { EmptyState } from "@/components/shared/EmptyState";
import { FilterBar } from "@/components/shared/FilterBar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { countryName } from "@/lib/countries";
import { compareUsdcAmounts, usdcAmount } from "@/lib/decimal";
import { activityTypeLabels } from "@/lib/labels";
import { trustTierPresentation, type TrustTier } from "@/lib/status";
import type { MarketplaceListing } from "@/lib/types";

const ANY = "any";

const uniqueSorted = (values: string[]) =>
  [...new Set(values)].sort((a, b) => a.localeCompare(b));

const isPlainPrice = (value: string) => /^\d+(\.\d{1,6})?$/.test(value);

export function ListingBrowser({
  listings,
}: {
  listings: MarketplaceListing[];
}) {
  const [activity, setActivity] = useState(ANY);
  const [vintage, setVintage] = useState(ANY);
  const [region, setRegion] = useState(ANY);
  const [facility, setFacility] = useState(ANY);
  const [tier, setTier] = useState(ANY);
  const [maxPrice, setMaxPrice] = useState("");

  const vintages = uniqueSorted(
    listings.map((listing) => String(listing.creditClass.vintageYear)),
  );
  const regions = uniqueSorted(
    listings.map((listing) => listing.creditClass.facilityCountry),
  );
  const facilities = uniqueSorted(
    listings.map((listing) => listing.creditClass.facilityName),
  );

  const priceCeiling = isPlainPrice(maxPrice) ? usdcAmount(maxPrice) : null;

  const filtered = useMemo(
    () =>
      listings.filter((listing) => {
        if (activity !== ANY && listing.creditClass.activityType !== activity) {
          return false;
        }
        if (
          vintage !== ANY &&
          String(listing.creditClass.vintageYear) !== vintage
        ) {
          return false;
        }
        if (region !== ANY && listing.creditClass.facilityCountry !== region) {
          return false;
        }
        if (facility !== ANY && listing.creditClass.facilityName !== facility) {
          return false;
        }
        if (tier !== ANY && listing.sellerTrustTier !== tier) {
          return false;
        }
        if (
          priceCeiling !== null &&
          compareUsdcAmounts(listing.pricePerTonne, priceCeiling) > 0
        ) {
          return false;
        }
        return true;
      }),
    [listings, activity, vintage, region, facility, tier, priceCeiling],
  );

  const activeCount =
    [activity, vintage, region, facility, tier].filter((value) => value !== ANY)
      .length + (maxPrice === "" ? 0 : 1);

  const clear = () => {
    setActivity(ANY);
    setVintage(ANY);
    setRegion(ANY);
    setFacility(ANY);
    setTier(ANY);
    setMaxPrice("");
  };

  return (
    <div className="space-y-4">
      <FilterBar activeCount={activeCount} onClear={clear}>
        <Select value={activity} onValueChange={setActivity}>
          <SelectTrigger className="w-52">
            <SelectValue placeholder="Any activity" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ANY}>Any activity</SelectItem>
            {Object.entries(activityTypeLabels).map(([value, label]) => (
              <SelectItem key={value} value={value}>
                {label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={vintage} onValueChange={setVintage}>
          <SelectTrigger className="w-36">
            <SelectValue placeholder="Any vintage" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ANY}>Any vintage</SelectItem>
            {vintages.map((year) => (
              <SelectItem key={year} value={year}>
                {year}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={region} onValueChange={setRegion}>
          <SelectTrigger className="w-44">
            <SelectValue placeholder="Any region" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ANY}>Any region</SelectItem>
            {regions.map((code) => (
              <SelectItem key={code} value={code}>
                {countryName(code)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={facility} onValueChange={setFacility}>
          <SelectTrigger className="w-52">
            <SelectValue placeholder="Any facility" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ANY}>Any facility</SelectItem>
            {facilities.map((name) => (
              <SelectItem key={name} value={name}>
                {name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={tier} onValueChange={setTier}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="Any trust tier" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ANY}>Any trust tier</SelectItem>
            {(Object.keys(trustTierPresentation) as TrustTier[]).map((name) => (
              <SelectItem key={name} value={name}>
                {trustTierPresentation[name].label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Input
          inputMode="decimal"
          placeholder="Max USDC / tCO2e"
          value={maxPrice}
          onChange={(event) => setMaxPrice(event.target.value)}
          className="w-44"
          aria-label="Maximum price per tCO2e"
        />
      </FilterBar>

      {filtered.length === 0 ? (
        <EmptyState
          icon={Store}
          title={
            activeCount > 0
              ? "No listings match those filters"
              : "No listings on the market"
          }
          description="Sellers list credits from a single credit class at a price per tonne. Listed credits sit in escrow until they sell or the listing expires."
          action={
            <Button asChild variant="outline">
              <Link href="/credits">View your credits</Link>
            </Button>
          }
        />
      ) : (
        <div className="grid gap-4">
          {filtered.map((listing) => (
            <ListingCard key={listing.id} listing={listing} />
          ))}
        </div>
      )}
    </div>
  );
}
