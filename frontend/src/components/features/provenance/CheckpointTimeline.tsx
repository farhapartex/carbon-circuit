import { Link2, MapPin } from "lucide-react";
import { TimestampDisplay } from "@/components/shared/TimestampDisplay";
import { Timeline, type TimelineEntry } from "@/components/shared/Timeline";
import type { Checkpoint, CheckpointType } from "@/lib/types";

const CHECKPOINT_LABELS: Record<CheckpointType, string> = {
  production_complete: "Production complete",
  departed_origin: "Departed origin",
  customs_export: "Cleared export customs",
  customs_import: "Cleared import customs",
  arrived_destination: "Arrived at destination",
};

const anchorLabel = (checkpoint: Checkpoint) => {
  if (checkpoint.anchor.status === "confirmed") {
    return `Anchored in epoch ${checkpoint.anchor.epoch}`;
  }
  if (checkpoint.anchor.status === "provisional") {
    return "Anchoring in progress";
  }
  return "Not yet anchored";
};

type CheckpointTimelineProps = {
  checkpoints: Checkpoint[];
  showReporter?: boolean | undefined;
};

export function CheckpointTimeline({
  checkpoints,
  showReporter = false,
}: CheckpointTimelineProps) {
  const entries: TimelineEntry[] = checkpoints.map((checkpoint) => ({
    id: checkpoint.id,
    title: CHECKPOINT_LABELS[checkpoint.type],
    superseded: Boolean(checkpoint.supersededByCheckpointId),
    variant:
      checkpoint.anchor.status === "confirmed"
        ? "success"
        : checkpoint.anchor.status === "provisional"
          ? "warning"
          : "neutral",
    meta: (
      <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <span className="inline-flex items-center gap-1">
          <MapPin className="size-3" aria-hidden />
          {checkpoint.location.label}, {checkpoint.location.countryCode}
        </span>
        <TimestampDisplay value={checkpoint.occurredAt} />
        {showReporter ? (
          <span>Reported by {checkpoint.reportedByOrganizationName}</span>
        ) : null}
      </span>
    ),
    body: (
      <div className="space-y-1">
        <p className="inline-flex items-center gap-1 text-caption text-muted-foreground">
          <Link2 className="size-3" aria-hidden />
          {anchorLabel(checkpoint)}
        </p>
        {checkpoint.correctionReason ? (
          <p className="rounded-lg bg-info-50 px-3 py-2 text-caption text-info-700">
            Correction: {checkpoint.correctionReason}
          </p>
        ) : null}
        {checkpoint.supersededByCheckpointId ? (
          <p className="text-caption text-muted-foreground">
            Superseded by a later correction. The original stays on the record.
          </p>
        ) : null}
      </div>
    ),
  }));

  return <Timeline entries={entries} />;
}
