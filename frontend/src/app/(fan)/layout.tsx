import type { ReactNode } from "react";

import { FanBottomNavigation } from "@/features/fan-navigation";

export default function FanLayout({ children }: { children: ReactNode }) {
  return (
    <div className="relative mx-auto min-h-svh w-full max-w-[408px] overflow-hidden bg-white text-foreground">
      <div className="relative min-h-svh overflow-hidden pb-[76px]">{children}</div>
      <div className="absolute inset-x-0 bottom-0 z-30">
        <FanBottomNavigation />
      </div>
    </div>
  );
}
