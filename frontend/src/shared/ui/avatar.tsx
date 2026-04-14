import * as AvatarPrimitive from "@radix-ui/react-avatar";
import type { ComponentPropsWithoutRef } from "react";

import { cn } from "@/shared/lib";

type AvatarProps = ComponentPropsWithoutRef<typeof AvatarPrimitive.Root>;
type AvatarImageProps = ComponentPropsWithoutRef<typeof AvatarPrimitive.Image>;
type AvatarFallbackProps = ComponentPropsWithoutRef<typeof AvatarPrimitive.Fallback>;

/**
 * creator や profile 表示用の avatar 容器を表示する。
 */
export function Avatar({ className, ...props }: AvatarProps) {
  return (
    <AvatarPrimitive.Root
      className={cn(
        "relative flex size-12 shrink-0 overflow-hidden rounded-full border border-border bg-[linear-gradient(180deg,#dbeaff_0%,#abcdf1_100%)] text-[#173252] shadow-[0_10px_24px_rgba(80,159,224,0.16)]",
        className,
      )}
      {...props}
    />
  );
}

/**
 * avatar 画像を表示する。
 */
export function AvatarImage({ className, ...props }: AvatarImageProps) {
  return <AvatarPrimitive.Image className={cn("aspect-square size-full object-cover", className)} {...props} />;
}

/**
 * avatar 画像がない場合の代替表示を行う。
 */
export function AvatarFallback({ className, ...props }: AvatarFallbackProps) {
  return (
    <AvatarPrimitive.Fallback
      className={cn(
        "flex size-full items-center justify-center bg-[linear-gradient(180deg,#d6e9ff_0%,#9fc7ec_100%)] font-display text-sm font-semibold uppercase tracking-[0.08em]",
        className,
      )}
      {...props}
    />
  );
}
