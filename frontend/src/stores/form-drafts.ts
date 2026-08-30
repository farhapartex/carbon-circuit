import { create } from "zustand";
import {
  createJSONStorage,
  persist,
  type StateStorage,
} from "zustand/middleware";

export type DraftKind = "batch" | "claim" | "organization";

export type FormDraft = {
  step: number;
  values: Record<string, string | number | boolean | null>;
  evidenceIds: string[];
  updatedAt: string;
};

type FormDraftState = {
  drafts: Partial<Record<DraftKind, FormDraft>>;
  saveDraft: (kind: DraftKind, draft: Omit<FormDraft, "updatedAt">) => void;
  clearDraft: (kind: DraftKind) => void;
};

const unavailableStorage: StateStorage = {
  getItem: () => null,
  setItem: () => undefined,
  removeItem: () => undefined,
};

const draftStorage = createJSONStorage(() =>
  typeof window === "undefined" ? unavailableStorage : window.sessionStorage,
);

export const useFormDraftStore = create<FormDraftState>()(
  persist(
    (set) => ({
      drafts: {},
      saveDraft: (kind, draft) =>
        set((state) => ({
          drafts: {
            ...state.drafts,
            [kind]: { ...draft, updatedAt: new Date().toISOString() },
          },
        })),
      clearDraft: (kind) =>
        set((state) => {
          const { [kind]: discarded, ...remaining } = state.drafts;
          void discarded;
          return { drafts: remaining };
        }),
    }),
    { name: "carboncircuit.form-drafts", storage: draftStorage },
  ),
);
