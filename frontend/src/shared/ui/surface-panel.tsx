import type { ComponentPropsWithoutRef } from "react";

import { cn } from "@/shared/lib";

type SurfacePanelProps = ComponentPropsWithoutRef<"div">;

/**
 * glass surface を持つ共通 panel を表示する。
 */
export function SurfacePanel({ className, ...props }: SurfacePanelProps) {
  return (
    <div
      className={cn(
        "rounded-[28px] border border-border bg-white shadow-[0_18px_40px_rgba(15,23,42,0.08)]",
        className,
      )}
      {...props}
    />
  );
}
