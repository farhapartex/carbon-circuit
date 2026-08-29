"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { UserPlus } from "lucide-react";
import { useState } from "react";
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

export function InviteMemberDialog() {
  const [open, setOpen] = useState(false);
  const pushToast = useToastQueueStore((state) => state.pushToast);

  const form = useForm<InviteValues>({
    resolver: zodResolver(inviteSchema),
    defaultValues: { email: "", role: "member" },
  });

  const submit = (values: InviteValues) => {
    pushToast({
      tone: "success",
      title: "Invitation sent",
      description: `${values.email} has seven days to accept.`,
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
                      {ROLE_OPTIONS.map((option) => (
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
        </Form>

        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button type="submit" form="invite-member">
            Send invitation
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
