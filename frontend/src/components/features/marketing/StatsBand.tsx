export type MarketingStat = {
  value: string;
  label: string;
};

type StatsBandProps = {
  stats: MarketingStat[];
};

export function StatsBand({ stats }: StatsBandProps) {
  return (
    <section className="border-y border-neutral-200 bg-white">
      <dl className="mx-auto grid max-w-6xl gap-8 px-6 py-12 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <div key={stat.label} className="space-y-1">
            <dt className="text-caption text-neutral-600">{stat.label}</dt>
            <dd className="text-section-heading tabular-nums">{stat.value}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
