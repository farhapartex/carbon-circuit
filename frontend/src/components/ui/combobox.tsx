"use client";

import { Check, ChevronDownIcon } from "lucide-react";
import * as React from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

export type ComboboxOption = {
  value: string;
  label: string;
};

type ComboboxProps = {
  options: ComboboxOption[];
  value: string;
  onValueChange: (value: string) => void;
  placeholder?: string | undefined;
  searchPlaceholder?: string | undefined;
  emptyMessage?: string | undefined;
  id?: string | undefined;
  className?: string | undefined;
};

const matches = (option: ComboboxOption, query: string) =>
  option.label.toLowerCase().includes(query) ||
  option.value.toLowerCase().includes(query);

export function Combobox({
  options,
  value,
  onValueChange,
  placeholder = "Select an option",
  searchPlaceholder = "Search",
  emptyMessage = "Nothing matches that.",
  id,
  className,
}: ComboboxProps) {
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");

  const selected = options.find((option) => option.value === value);
  const trimmed = query.trim().toLowerCase();
  const visible = trimmed
    ? options.filter((option) => matches(option, trimmed))
    : options;

  const choose = (option: ComboboxOption) => {
    onValueChange(option.value);
    setQuery("");
    setOpen(false);
  };

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) setQuery("");
      }}
    >
      <PopoverTrigger
        id={id}
        role="combobox"
        aria-expanded={open}
        className={cn(
          "flex h-8 w-full items-center justify-between gap-1.5 rounded-lg border border-input bg-transparent py-2 pr-2 pl-2.5 text-sm whitespace-nowrap transition-colors select-none disabled:cursor-not-allowed disabled:opacity-50",
          className,
        )}
      >
        <span className={cn("truncate", !selected && "text-muted-foreground")}>
          {selected?.label ?? placeholder}
        </span>
        <ChevronDownIcon
          className="pointer-events-none size-4 shrink-0 text-muted-foreground"
          aria-hidden
        />
      </PopoverTrigger>

      <PopoverContent
        align="start"
        className="w-(--radix-popover-trigger-width) p-0"
      >
        <div className="border-b border-neutral-200 p-2">
          <input
            autoFocus
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={searchPlaceholder}
            aria-label={searchPlaceholder}
            className="w-full rounded-md border border-input px-2 py-1 text-sm outline-none"
          />
        </div>

        <ul className="max-h-64 overflow-y-auto p-1" role="listbox">
          {visible.length === 0 ? (
            <li className="px-2 py-3 text-center text-caption text-muted-foreground">
              {emptyMessage}
            </li>
          ) : (
            visible.map((option) => (
              <li key={option.value}>
                <button
                  type="button"
                  data-slot="combobox-item"
                  role="option"
                  aria-selected={option.value === value}
                  onClick={() => choose(option)}
                  className="flex w-full items-center justify-between gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
                >
                  <span className="truncate">{option.label}</span>
                  {option.value === value ? (
                    <Check className="size-4 shrink-0" aria-hidden />
                  ) : null}
                </button>
              </li>
            ))
          )}
        </ul>
      </PopoverContent>
    </Popover>
  );
}
