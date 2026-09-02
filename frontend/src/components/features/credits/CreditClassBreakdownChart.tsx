"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { activityTypeLabels } from "@/lib/labels";
import type { CreditClassBalance } from "@/lib/types";

const numberFormat = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 2,
});

export function CreditClassBreakdownChart({
  balances,
}: {
  balances: CreditClassBalance[];
}) {
  const data = balances.map((balance) => ({
    name: `${activityTypeLabels[balance.creditClass.activityType]} ${balance.creditClass.vintageYear}`,
    Available: Number(balance.available),
    Escrowed: Number(balance.escrowed),
    Retired: Number(balance.retired),
  }));

  return (
    <Card>
      <CardHeader>
        <CardTitle>Holdings by credit class</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} margin={{ left: 8, right: 8, top: 8 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis
                dataKey="name"
                tick={{ fontSize: 12 }}
                interval={0}
                height={56}
              />
              <YAxis
                tick={{ fontSize: 12 }}
                tickFormatter={(value: number) => numberFormat.format(value)}
              />
              <Tooltip
                formatter={(value) =>
                  `${numberFormat.format(Number(value ?? 0))} tCO2e`
                }
              />
              <Legend />
              <Bar
                dataKey="Available"
                stackId="holding"
                fill="var(--color-success-700)"
              />
              <Bar
                dataKey="Escrowed"
                stackId="holding"
                fill="var(--color-warning-700)"
              />
              <Bar
                dataKey="Retired"
                stackId="holding"
                fill="var(--color-neutral-600)"
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
        <p className="mt-3 text-caption text-pretty text-neutral-600">
          Chart figures are rounded for display only. Every balance, sale, and
          retirement is carried in exact decimal to six places.
        </p>
      </CardContent>
    </Card>
  );
}
