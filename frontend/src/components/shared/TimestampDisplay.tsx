import type { IsoTimestamp } from "@/lib/types";
import { cn } from "@/lib/utils";

const utcFormat = new Intl.DateTimeFormat("en-GB", {
  timeZone: "UTC",
  day: "numeric",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});

const utcDateFormat = new Intl.DateTimeFormat("en-GB", {
  timeZone: "UTC",
  day: "numeric",
  month: "short",
  year: "numeric",
});

type TimestampDisplayProps = {
  value: IsoTimestamp;
  dateOnly?: boolean | undefined;
  className?: string | undefined;
};

export function TimestampDisplay({
  value,
  dateOnly = false,
  className,
}: TimestampDisplayProps) {
  const parsed = new Date(value);
  const formatter = dateOnly ? utcDateFormat : utcFormat;

  return (
    <time
      dateTime={value}
      title={`${utcFormat.format(parsed)} UTC`}
      className={cn("tabular-nums", className)}
    >
      {formatter.format(parsed)}
      {dateOnly ? null : " UTC"}
    </time>
  );
}
