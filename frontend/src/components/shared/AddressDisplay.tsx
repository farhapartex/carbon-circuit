import { CopyButton } from "@/components/shared/CopyButton";
import type { EthereumAddress } from "@/lib/types";
import { cn } from "@/lib/utils";

const LEADING_CHARACTERS = 6;
const TRAILING_CHARACTERS = 4;

export const truncateAddress = (address: string) =>
  address.length <= LEADING_CHARACTERS + TRAILING_CHARACTERS
    ? address
    : `${address.slice(0, LEADING_CHARACTERS)}…${address.slice(-TRAILING_CHARACTERS)}`;

type AddressDisplayProps = {
  address: EthereumAddress;
  knownAs?: string;
  className?: string;
  showCopy?: boolean;
};

export function AddressDisplay({
  address,
  knownAs,
  className,
  showCopy = true,
}: AddressDisplayProps) {
  return (
    <span
      data-slot="address-display"
      className={cn("inline-flex items-center gap-1.5", className)}
    >
      {knownAs ? <span className="font-medium">{knownAs}</span> : null}
      <span
        className="font-mono text-caption text-muted-foreground tabular-nums"
        title={address}
      >
        {truncateAddress(address)}
      </span>
      {showCopy ? <CopyButton value={address} label="address" /> : null}
    </span>
  );
}
