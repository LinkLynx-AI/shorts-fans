import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import type { ButtonHTMLAttributes } from "react";

import { cn } from "@/shared/lib";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-full text-sm font-semibold transition focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/70 disabled:pointer-events-none disabled:opacity-55",
  {
    defaultVariants: {
      size: "default",
      variant: "default",
    },
    variants: {
      size: {
        default: "h-11 px-5",
        icon: "size-10 p-0",
        lg: "h-12 px-6 text-base",
        sm: "h-9 px-4 text-[11px] uppercase tracking-[0.14em]",
      },
      variant: {
        chrome:
          "border border-white/72 bg-white/76 text-accent-strong shadow-[0_10px_24px_rgba(36,94,132,0.1)] backdrop-blur-md hover:bg-white/90",
        default:
          "bg-[linear-gradient(135deg,var(--accent)_0%,var(--accent-strong)_100%)] text-white shadow-[0_18px_44px_rgba(16,130,200,0.28)] hover:brightness-105",
        ghost: "bg-transparent text-foreground hover:bg-white/12",
        secondary:
          "bg-white/82 text-foreground shadow-[inset_0_0_0_1px_rgba(167,220,249,0.52),0_12px_28px_rgba(36,94,132,0.12)] backdrop-blur-md hover:bg-white",
      },
    },
  },
);

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

/**
 * shadcn/ui 互換の基本ボタンを表示する。
 */
export function Button({
  asChild = false,
  className,
  size,
  variant,
  ...props
}: ButtonProps) {
  const Comp = asChild ? Slot : "button";

  return <Comp className={cn(buttonVariants({ className, size, variant }))} {...props} />;
}
