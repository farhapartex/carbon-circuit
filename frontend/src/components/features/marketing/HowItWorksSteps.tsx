export type WorkStep = {
  title: string;
  description: string;
};

type HowItWorksStepsProps = {
  heading: string;
  steps: WorkStep[];
};

export function HowItWorksSteps({ heading, steps }: HowItWorksStepsProps) {
  return (
    <section className="border-y border-neutral-200 bg-neutral-100">
      <div className="mx-auto max-w-6xl px-6 py-20">
        <h2 className="text-section-heading">{heading}</h2>
        <ol className="mt-10 grid gap-8 md:grid-cols-2 lg:grid-cols-4">
          {steps.map((step, index) => (
            <li key={step.title} className="space-y-3">
              <span className="flex size-8 items-center justify-center rounded-full bg-primary-700 text-caption font-medium text-white tabular-nums">
                {index + 1}
              </span>
              <h3 className="font-medium">{step.title}</h3>
              <p className="text-caption text-pretty text-neutral-600">
                {step.description}
              </p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
