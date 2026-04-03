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
        "rounded-[26px] border border-white/72 bg-white/74 shadow-[0_18px_40px_rgba(36,94,132,0.12)] backdrop-blur-xl",
        className,
      )}
      {...props}
    />
  );
}
