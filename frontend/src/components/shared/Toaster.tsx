"use client";

import {
  Toast,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from "@/components/ui/toast";
import { useToastQueueStore } from "@/stores/toast-queue";

export function Toaster() {
  const toasts = useToastQueueStore((state) => state.toasts);
  const dismissToast = useToastQueueStore((state) => state.dismissToast);

  return (
    <ToastProvider swipeDirection="right">
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          tone={toast.tone}
          onOpenChange={(open) => {
            if (!open) dismissToast(toast.id);
          }}
        >
          <div className="grid gap-1">
            <ToastTitle>{toast.title}</ToastTitle>
            {toast.description ? (
              <ToastDescription>{toast.description}</ToastDescription>
            ) : null}
          </div>
          <ToastClose />
        </Toast>
      ))}
      <ToastViewport />
    </ToastProvider>
  );
}
