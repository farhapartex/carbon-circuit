import type { FieldValues, Path, UseFormSetError } from "react-hook-form";

export type ApiFieldError = {
  field: string;
  code: string;
};

const FIELD_ERROR_MESSAGES: Record<string, string> = {
  REQUIRED: "This field is required.",
  BELOW_MINIMUM: "This value is below the allowed minimum.",
  ABOVE_MAXIMUM: "This value is above the allowed maximum.",
  EXCEEDS_CEILING: "This exceeds the computed ceiling for this claim.",
  INVALID_FORMAT: "This value is not in the expected format.",
  ALREADY_EXISTS: "This value is already in use.",
  UNSUPPORTED_VALUE: "This value is not accepted for this field.",
  TOO_SHORT: "This value is shorter than the required length.",
  TOO_LONG: "This value is longer than the allowed length.",
};

export const messageForFieldError = (code: string) =>
  FIELD_ERROR_MESSAGES[code] ?? "This value was rejected.";

export const applyApiFieldErrors = <TFieldValues extends FieldValues>(
  setError: UseFormSetError<TFieldValues>,
  details: ApiFieldError[],
) => {
  for (const detail of details) {
    setError(detail.field as Path<TFieldValues>, {
      type: "server",
      message: messageForFieldError(detail.code),
    });
  }
};
