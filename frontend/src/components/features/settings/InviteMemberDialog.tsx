"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { UserPlus } from "lucide-react";
import { useState, useTransition } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CopyButton } from "@/components/shared/CopyButton";
import { submitInvitation } from "@/lib/actions/members";
import { useToastQueueStore } from "@/stores/toast-queue";

const inviteSchema = z.object({
  email: z.email("Enter a valid email address."),
  role: z.enum(["owner", "admin", "member"]),
});

type InviteValues = z.infer<typeof inviteSchema>;

const ROLE_OPTIONS: {
  value: InviteValues["role"];
  label: string;
  detail: string;
}[] = [
  {
    value: "member",
    label: "Member",
    detail: "Can create batches, log checkpoints, and submit claims",
  },
  {
    value: "admin",
    label: "Administrator",
    detail: "Everything a member can do, plus managing people and billing",
  },
  {
    value: "owner",
    label: "Owner",
    detail: "Full control, including changing the Treasury Address",
  },
];

type InviteMemberDialogProps = {
  canGrantOwner: boolean;
};

export function InviteMemberDialog({ canGrantOwner }: InviteMemberDialogProps) {
  const [open, setOpen] = useState(false);
  const [pending, startTransition] = useTransition();
  const [inviteLink, setInviteLink] = useState<string | null>(null);
  const pushToast = useToastQueueStore((state) => state.pushToast);

  const grantable = canGrantOwner
    ? ROLE_OPTIONS
    : ROLE_OPTIONS.filter((option) => option.value !== "owner");

  const form = useForm<InviteValues>({
    resolver: zodResolver(inviteSchema),
    defaultValues: { email: "", role: "member" },
  });

  const submit = (values: InviteValues) => {
    startTransition(async () => {
      const result = await submitInvitation(
        values.email,
        values.role,
        crypto.randomUUID(),
      );

      if (!result.ok) {
        pushToast({
          tone: "danger",
          title: "Invitation not sent",
          description: inviteFailure(result.code),
        });
        return;
      }

      setInviteLink(`${window.location.origin}/invite/${result.issued.token}`);
      form.reset();
      pushToast({
        tone: "success",
        title: "Invitation created",
        description: `Send the link to ${values.email}. It expires in seven days.`,
      });
    });
    form.reset();
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <UserPlus className="size-4" aria-hidden />
          Invite member
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite a member</DialogTitle>
          <DialogDescription>
            They will receive an email with a single-use link that expires in
            seven days.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form
            id="invite-member"
            onSubmit={form.handleSubmit(submit)}
            className="space-y-4"
          >
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email address</FormLabel>
                  <FormControl>
                    <Input
                      type="email"
                      autoComplete="off"
                      placeholder="name@company.com"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    The invitation binds to whoever signs in with a verified
                    address matching this one.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="role"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Role</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {grantable.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    {
                      ROLE_OPTIONS.find(
                        (option) => option.value === field.value,
                      )?.detail
                    }
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>

          {inviteLink ? (
            <div className="space-y-2 rounded-md border border-primary-600/30 bg-primary-50 px-4 py-3">
              <p className="font-700 text-helper text-primary-700">
                Send this link to the person you invited
              </p>
              <p className="text-helper text-primary-700">
                It is shown once and cannot be retrieved later.
              </p>
              <div className="flex items-center gap-2">
                <code className="min-w-0 flex-1 truncate rounded bg-white px-2 py-1 text-helper">
                  {inviteLink}
                </code>
                <CopyButton value={inviteLink} label="Copy invite link" />
              </div>
            </div>
          ) : null}
        </Form>

        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button type="submit" form="invite-member" disabled={pending}>
            {pending ? "Creating invitation…" : "Create invitation"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function inviteFailure(code: string): string {
  switch (code) {
    case "FORBIDDEN":
      return "Your role does not permit inviting that role.";
    case "CONFLICT":
      return "An invitation is already pending for that address.";
    case "VALIDATION_ERROR":
      return "Check the email address and role.";
    default:
      return "Something went wrong on our side. Try again shortly.";
  }
}
