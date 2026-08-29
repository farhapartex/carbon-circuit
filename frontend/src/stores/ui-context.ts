import { create } from "zustand";
import { persist } from "zustand/middleware";

export type TableDensity = "comfortable" | "compact";

type UiContextState = {
  activeOrganizationId: string | null;
  sidebarCollapsed: boolean;
  tableDensity: TableDensity;
  setActiveOrganization: (organizationId: string) => void;
  toggleSidebar: () => void;
  setTableDensity: (density: TableDensity) => void;
};

export const useUiContextStore = create<UiContextState>()(
  persist(
    (set) => ({
      activeOrganizationId: null,
      sidebarCollapsed: false,
      tableDensity: "comfortable",
      setActiveOrganization: (organizationId) =>
        set({ activeOrganizationId: organizationId }),
      toggleSidebar: () =>
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      setTableDensity: (density) => set({ tableDensity: density }),
    }),
    { name: "carboncircuit.ui-context" },
  ),
);
