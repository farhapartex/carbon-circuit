import { Bot, TriangleAlert } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { AIReviewRecord } from "@/lib/api/verifier";

const assessmentLabels: Record<string, string> = {
  not_assessed: "Not assessed",
  corroborated: "Corroborated by the evidence",
  corroborated_with_discrepancy: "Corroborated, with a discrepancy",
  uncorroborated: "Not corroborated by the evidence",
  contradicted: "Contradicted by the evidence",
};

export function AIAssessmentPanel({
  review,
}: {
  review: AIReviewRecord | null;
}) {
  const unavailable = review === null || review.assessment === "not_assessed";

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bot className="size-4 shrink-0 text-neutral-600" aria-hidden />
          AI assessment
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {unavailable ? (
          <div className="rounded-md border border-warning-600 bg-warning-50 px-4 py-3">
            <p className="flex items-center gap-2 font-medium text-warning-700">
              <TriangleAlert className="size-4 shrink-0" aria-hidden />
              No assessment is available
            </p>
            <p className="mt-1 text-caption text-pretty text-warning-700">
              {review?.narrative ??
                "This claim was not assessed before it reached you."}{" "}
              Read the evidence yourself and weigh it against the declared
              figures. The decision was always yours to make; you are simply
              making it without a head start.
            </p>
            {review ? (
              <p className="mt-2 text-caption text-warning-700">
                Recorded reviewer: <code>{review.assessedBy}</code>
              </p>
            ) : null}
          </div>
        ) : (
          <>
            <p className="font-medium">
              {assessmentLabels[review.assessment] ?? review.assessment}
            </p>
            {review.confidence ? (
              <p className="text-caption text-neutral-600">
                Confidence {review.confidence}
              </p>
            ) : null}
            <p className="text-caption text-pretty text-neutral-600">
              {review.narrative}
            </p>
            {review.flags.length > 0 ? (
              <ul className="space-y-1">
                {review.flags.map((flag) => (
                  <li key={flag} className="text-caption text-warning-700">
                    {flag}
                  </li>
                ))}
              </ul>
            ) : null}
          </>
        )}

        <p className="border-t border-neutral-200 pt-3 text-caption text-pretty text-neutral-600">
          The assessment never decides the outcome and never computes the credit
          amount. The amount comes from the formula, capped by the ceiling.
        </p>
      </CardContent>
    </Card>
  );
}
