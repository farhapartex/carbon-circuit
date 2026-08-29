import { Cpu, Pill, Shirt, Wheat } from "lucide-react";

const industries = [
  {
    icon: Cpu,
    name: "Electronics",
    status: "Launch vertical",
    available: true,
  },
  {
    icon: Wheat,
    name: "Agriculture",
    status: "Designed for",
    available: false,
  },
  { icon: Pill, name: "Pharma", status: "Designed for", available: false },
  { icon: Shirt, name: "Textiles", status: "Designed for", available: false },
];

export function IndustryBand() {
  return (
    <section className="border-y border-neutral-200 bg-white">
      <div className="mx-auto max-w-6xl px-6 py-12">
        <p className="text-caption text-neutral-600">
          Product Category is a first-class part of the data model, not a label.
          Each one defines its own batch attributes, expected checkpoint
          sequence, and available claim types.
        </p>
        <ul className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {industries.map((industry) => (
            <li
              key={industry.name}
              className="flex items-center gap-3 rounded-lg border border-neutral-200 p-4"
            >
              <industry.icon
                className={
                  industry.available
                    ? "size-5 text-primary-600"
                    : "size-5 text-neutral-400"
                }
                aria-hidden
              />
              <span>
                <span className="block font-medium">{industry.name}</span>
                <span className="block text-caption text-neutral-600">
                  {industry.status}
                </span>
              </span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
