import { CapacityCard } from "@/components/features/facilities/CapacityCard";
import { FacilityProfileCard } from "@/components/features/facilities/FacilityProfileCard";
import { MetricCard } from "@/components/shared/MetricCard";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { FacilityRecord } from "@/lib/api/facilities";

const PENDING_TABS = [
  {
    value: "batches",
    label: "Batches",
    notice:
      "Batches originating at this facility will be listed here once the provenance service is serving them.",
  },
  {
    value: "claims",
    label: "Claims",
    notice:
      "Sustainability claims filed for this facility will be listed here once the sustainability service exists.",
  },
  {
    value: "credits",
    label: "Credits",
    notice:
      "Credits issued from this facility will be listed here once the credit ledger exists.",
  },
];

export function FacilityTabs({ facility }: { facility: FacilityRecord }) {
  return (
    <Tabs defaultValue="overview">
      <TabsList>
        <TabsTrigger value="overview">Overview</TabsTrigger>
        {PENDING_TABS.map((tab) => (
          <TabsTrigger key={tab.value} value={tab.value}>
            {tab.label}
          </TabsTrigger>
        ))}
      </TabsList>

      <TabsContent value="overview" className="space-y-6">
        <div className="grid gap-4 sm:grid-cols-3">
          <MetricCard
            label="Ceiling discount"
            value={facility.ceilingDiscountFactor}
            hint="Applied to every credit ceiling computed for this facility."
          />
          <MetricCard
            label="Registry reference"
            value={facility.facilityReference ?? "None supplied"}
          />
          <MetricCard label="Trust tier" value={facility.trustTier} />
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <FacilityProfileCard facility={facility} />
          <CapacityCard facility={facility} />
        </div>
      </TabsContent>

      {PENDING_TABS.map((tab) => (
        <TabsContent key={tab.value} value={tab.value}>
          <p className="rounded-lg border border-dashed border-neutral-200 bg-white px-6 py-10 text-center text-caption text-pretty text-neutral-600">
            {tab.notice}
          </p>
        </TabsContent>
      ))}
    </Tabs>
  );
}
