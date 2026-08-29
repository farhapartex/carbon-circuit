import { SettingsNav } from "@/components/features/settings/SettingsNav";
import { PageHeader } from "@/components/shared/PageHeader";

export default function SettingsLayout({ children }: LayoutProps<"/settings">) {
  return (
    <div className="space-y-8">
      <PageHeader
        title="Settings"
        description="Everything about you, your organization, and how you pay for it."
      />
      <div className="flex flex-col gap-8 lg:flex-row">
        <SettingsNav />
        <div className="min-w-0 flex-1 space-y-6">{children}</div>
      </div>
    </div>
  );
}
