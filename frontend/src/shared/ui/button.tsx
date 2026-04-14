import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import type { ButtonHTMLAttributes } from "react";

import { cn } from "@/shared/lib";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-full border border-transparent text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/60 disabled:pointer-events-none disabled:opacity-55",
  {
    defaultVariants: {
      size: "default",
      variant: "default",
    },
    variants: {
      size: {
        default: "h-11 px-5 text-[15px]",
        icon: "size-10 p-0",
        lg: "h-12 px-6 text-base",
        sm: "h-9 px-4 text-sm",
      },
      variant: {
        chrome:
          "border-border bg-white text-foreground shadow-[0_8px_20px_rgba(15,23,42,0.06)] hover:bg-surface-subtle",
        default:
          "bg-accent text-white shadow-[0_12px_28px_rgba(80,159,224,0.24)] hover:bg-accent-strong",
        ghost: "bg-transparent text-foreground hover:bg-accent-soft",
        secondary:
          "border-border bg-white text-foreground shadow-[0_10px_24px_rgba(15,23,42,0.06)] hover:bg-surface-subtle",
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
