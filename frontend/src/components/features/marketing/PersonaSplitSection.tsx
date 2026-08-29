import { ArrowRight, Leaf, Route as RouteIcon } from "lucide-react";
import Link from "next/link";
import type { Route } from "next";

type Persona = {
  icon: typeof RouteIcon;
  eyebrow: string;
  title: string;
  description: string;
  bullets: string[];
  href: Route;
  linkLabel: string;
};

const personas: Persona[] = [
  {
    icon: RouteIcon,
    eyebrow: "For manufacturers, assemblers, and logistics partners",
    title: "Track my supply chain",
    description:
      "Record every step a batch takes, from raw material to finished good, and give anyone holding the product a way to check the journey rather than take your word for it.",
    bullets: [
      "Batch and checkpoint history that cannot be quietly rewritten",
      "A Provenance Score that explains itself component by component",
      "A public page behind every QR code, with no login required",
    ],
    href: "/solutions/traceability",
    linkLabel: "See how traceability works",
  },
  {
    icon: Leaf,
    eyebrow: "For sustainability teams and credit buyers",
    title: "Buy verified carbon credits",
    description:
      "Every credit names the facility that earned it, the year, and the practice behind it. That attribution survives every transfer and is still attached the moment it is retired.",
    bullets: [
      "AI-assisted review with a human verifier making every decision",
      "Credit Class attribution you can quote in an ESG report",
      "A public retirement log anyone can audit",
    ],
    href: "/solutions/carbon-credits",
    linkLabel: "See how credits work",
  },
];

export function PersonaSplitSection() {
  return (
    <section className="mx-auto max-w-6xl px-6 py-20">
      <div className="grid gap-6 lg:grid-cols-2">
        {personas.map((persona) => (
          <article
            key={persona.title}
            className="flex flex-col gap-5 rounded-lg border border-neutral-200 bg-white p-8 shadow-sm"
          >
            <span className="flex size-10 items-center justify-center rounded-lg bg-primary-50">
              <persona.icon className="size-5 text-primary-600" aria-hidden />
            </span>
            <div className="space-y-2">
              <p className="text-caption text-neutral-600">{persona.eyebrow}</p>
              <h2 className="text-section-heading">{persona.title}</h2>
              <p className="text-body text-pretty text-neutral-600">
                {persona.description}
              </p>
            </div>
            <ul className="space-y-2">
              {persona.bullets.map((bullet) => (
                <li
                  key={bullet}
                  className="flex gap-2 text-caption text-neutral-600"
                >
                  <span
                    className="mt-1.5 size-1.5 shrink-0 rounded-full bg-primary-600"
                    aria-hidden
                  />
                  {bullet}
                </li>
              ))}
            </ul>
            <Link
              href={persona.href}
              className="mt-auto inline-flex items-center gap-1.5 rounded-sm font-medium text-primary-700 underline underline-offset-4"
            >
              {persona.linkLabel}
              <ArrowRight className="size-4" aria-hidden />
            </Link>
          </article>
        ))}
      </div>
    </section>
  );
}
