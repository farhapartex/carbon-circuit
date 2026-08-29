"use client";

import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { Button } from "@/components/ui/button";
import { useToastQueueStore } from "@/stores/toast-queue";

export function CancelSubscriptionDialog({ planName }: { planName: string }) {
  const [open, setOpen] = useState(false);
  const pushToast = useToastQueueStore((state) => state.pushToast);

  return (
    <>
      <Button variant="ghost" size="sm" onClick={() => setOpen(true)}>
        Cancel subscription
      </Button>

      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title={`Cancel ${planName}?`}
        description="Your subscription stays active until the end of the current billing period, then the organization becomes read-only."
        consequence="Your credits are never seized and any active listings remain honourable. You can still log in, view everything, and export your data. You will not be able to create batches, log checkpoints, submit claims, or create listings."
        confirmLabel="Cancel subscription"
        destructive
        requiredPhrase="CANCEL"
        onConfirm={() => {
          pushToast({
            tone: "warning",
            title: "Subscription cancelled",
            description:
              "Access continues until the end of this billing period.",
          });
          setOpen(false);
        }}
      />
    </>
  );
}
