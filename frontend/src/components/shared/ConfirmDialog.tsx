"use client";

import { useState, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type ConfirmDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmLabel: string;
  onConfirm: () => void;
  destructive?: boolean | undefined;
  consequence?: ReactNode | undefined;
  requiredPhrase?: string | undefined;
};

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  onConfirm,
  destructive = false,
  consequence,
  requiredPhrase,
}: ConfirmDialogProps) {
  const [typedPhrase, setTypedPhrase] = useState("");
  const phraseSatisfied = !requiredPhrase || typedPhrase === requiredPhrase;

  const close = (next: boolean) => {
    if (!next) setTypedPhrase("");
    onOpenChange(next);
  };

  return (
    <Dialog open={open} onOpenChange={close}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>

        {consequence ? (
          <div className="rounded-lg bg-warning-50 px-4 py-3 text-caption text-warning-700">
            {consequence}
          </div>
        ) : null}

        {requiredPhrase ? (
          <div className="grid gap-2">
            <Label htmlFor="confirm-phrase">
              Type <span className="font-mono">{requiredPhrase}</span> to
              confirm
            </Label>
            <Input
              id="confirm-phrase"
              value={typedPhrase}
              onChange={(event) => setTypedPhrase(event.target.value)}
              autoComplete="off"
            />
          </div>
        ) : null}

        <DialogFooter>
          <Button variant="outline" onClick={() => close(false)}>
            Cancel
          </Button>
          <Button
            variant={destructive ? "destructive" : "default"}
            disabled={!phraseSatisfied}
            onClick={() => {
              onConfirm();
              close(false);
            }}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
