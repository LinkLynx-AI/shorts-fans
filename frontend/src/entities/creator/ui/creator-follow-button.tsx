import { cn } from "@/shared/lib";
import { Button, type ButtonProps } from "@/shared/ui";

type CreatorFollowButtonLabels = {
  follow: string;
  followPending: string;
  following: string;
  unfollowPending: string;
};

export type CreatorFollowButtonProps = Omit<ButtonProps, "children" | "variant"> & {
  fullWidth?: boolean | undefined;
  isFollowing: boolean;
  isPending?: boolean | undefined;
  labels?: Partial<CreatorFollowButtonLabels> | undefined;
};

const defaultCreatorFollowButtonLabels = {
  follow: "Follow",
  followPending: "Following...",
  following: "Following",
  unfollowPending: "Unfollowing...",
} as const satisfies CreatorFollowButtonLabels;

function resolveCreatorFollowButtonLabel(
  isFollowing: boolean,
  isPending: boolean,
  labels: CreatorFollowButtonLabels,
): string {
  if (!isPending) {
    return isFollowing ? labels.following : labels.follow;
  }

  return isFollowing ? labels.unfollowPending : labels.followPending;
}

/**
 * creator の follow relation 状態に応じた CTA button を表示する。
 */
export function CreatorFollowButton({
  className,
  disabled,
  fullWidth = false,
  isFollowing,
  isPending = false,
  labels,
  type = "button",
  ...props
}: CreatorFollowButtonProps) {
  const resolvedLabels = {
    ...defaultCreatorFollowButtonLabels,
    ...labels,
  };

  return (
    <Button
      {...props}
      aria-busy={isPending || undefined}
      aria-pressed={isFollowing}
      className={cn(
        "text-[15px] font-semibold",
        fullWidth && "w-full",
        isFollowing && "shadow-none",
        className,
      )}
      disabled={disabled || isPending}
      type={type}
      variant={isFollowing ? "secondary" : "default"}
    >
      {resolveCreatorFollowButtonLabel(isFollowing, isPending, resolvedLabels)}
    </Button>
  );
}
