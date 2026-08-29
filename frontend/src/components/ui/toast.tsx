"use client";

import { cva, type VariantProps } from "class-variance-authority";
import { X } from "lucide-react";
import { Toast as ToastPrimitive } from "radix-ui";
import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

const ToastProvider = ToastPrimitive.Provider;

function ToastViewport({
  className,
  ...props
}: ComponentProps<typeof ToastPrimitive.Viewport>) {
  return (
    <ToastPrimitive.Viewport
      data-slot="toast-viewport"
      className={cn(
        "fixed right-0 bottom-0 z-100 flex max-h-screen w-full flex-col-reverse gap-2 p-4 sm:top-0 sm:right-0 sm:bottom-auto sm:flex-col md:max-w-sm",
        className,
      )}
      {...props}
    />
  );
}

const toastVariants = cva(
  "group/toast pointer-events-auto relative flex w-full items-start gap-3 rounded-lg border p-4 shadow-sm transition-all data-[state=closed]:animate-out data-[state=closed]:fade-out-80 data-[state=open]:animate-in data-[state=open]:slide-in-from-top-full data-[swipe=end]:animate-out sm:data-[state=open]:slide-in-from-bottom-full",
  {
    variants: {
      tone: {
        neutral: "border-neutral-200 bg-white text-neutral-900",
        success: "border-success-600/30 bg-success-50 text-success-700",
        warning: "border-warning-600/30 bg-warning-50 text-warning-700",
        danger: "border-danger-600/30 bg-danger-50 text-danger-700",
        info: "border-info-600/30 bg-info-50 text-info-700",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

function Toast({
  className,
  tone,
  ...props
}: ComponentProps<typeof ToastPrimitive.Root> &
  VariantProps<typeof toastVariants>) {
  return (
    <ToastPrimitive.Root
      data-slot="toast"
      className={cn(toastVariants({ tone }), className)}
      {...props}
    />
  );
}

function ToastTitle({
  className,
  ...props
}: ComponentProps<typeof ToastPrimitive.Title>) {
  return (
    <ToastPrimitive.Title
      data-slot="toast-title"
      className={cn("text-sm font-medium", className)}
      {...props}
    />
  );
}

function ToastDescription({
  className,
  ...props
}: ComponentProps<typeof ToastPrimitive.Description>) {
  return (
    <ToastPrimitive.Description
      data-slot="toast-description"
      className={cn("text-sm opacity-90", className)}
      {...props}
    />
  );
}

function ToastClose({
  className,
  ...props
}: ComponentProps<typeof ToastPrimitive.Close>) {
  return (
    <ToastPrimitive.Close
      data-slot="toast-close"
      aria-label="Dismiss notification"
      className={cn(
        "ml-auto shrink-0 rounded-sm opacity-70 transition-opacity hover:opacity-100",
        className,
      )}
      {...props}
    >
      <X className="size-4" />
    </ToastPrimitive.Close>
  );
}

export {
  Toast,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
  toastVariants,
};
