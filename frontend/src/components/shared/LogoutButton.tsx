import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";

export const LOGOUT_PATH = "/auth/logout";

type LogoutButtonProps = {
  label?: string | undefined;
  variant?: "ghost" | "outline" | undefined;
};

export function LogoutButton({
  label = "Sign out",
  variant = "ghost",
}: LogoutButtonProps) {
  return (
    <Button asChild variant={variant} size="sm">
      <a href={LOGOUT_PATH}>
        <LogOut className="size-4" aria-hidden />
        {label}
      </a>
    </Button>
  );
}
