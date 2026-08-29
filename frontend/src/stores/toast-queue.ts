import { create } from "zustand";

export type ToastTone = "neutral" | "success" | "warning" | "danger" | "info";

export type Toast = {
  id: string;
  tone: ToastTone;
  title: string;
  description?: string;
};

type ToastQueueState = {
  toasts: Toast[];
  pushToast: (toast: Omit<Toast, "id">) => string;
  dismissToast: (id: string) => void;
  clearToasts: () => void;
};

export const useToastQueueStore = create<ToastQueueState>()((set) => ({
  toasts: [],
  pushToast: (toast) => {
    const id = crypto.randomUUID();
    set((state) => ({ toasts: [...state.toasts, { ...toast, id }] }));
    return id;
  },
  dismissToast: (id) =>
    set((state) => ({ toasts: state.toasts.filter((it) => it.id !== id) })),
  clearToasts: () => set({ toasts: [] }),
}));
