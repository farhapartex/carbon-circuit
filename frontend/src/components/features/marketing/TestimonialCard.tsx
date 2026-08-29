import { Avatar, AvatarFallback } from "@/components/ui/avatar";

type TestimonialCardProps = {
  quote: string;
  authorName: string;
  authorRole: string;
  authorCompany: string;
};

const initialsOf = (name: string) =>
  name
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join("");

export function TestimonialCard({
  quote,
  authorName,
  authorRole,
  authorCompany,
}: TestimonialCardProps) {
  return (
    <figure className="space-y-4 rounded-lg border border-neutral-200 bg-white p-6 shadow-sm">
      <blockquote className="text-body text-pretty text-neutral-900">
        {quote}
      </blockquote>
      <figcaption className="flex items-center gap-3">
        <Avatar className="size-8">
          <AvatarFallback className="text-caption">
            {initialsOf(authorName)}
          </AvatarFallback>
        </Avatar>
        <span className="text-caption text-neutral-600">
          <span className="font-medium text-neutral-900">{authorName}</span>
          {" · "}
          {authorRole}, {authorCompany}
        </span>
      </figcaption>
    </figure>
  );
}
