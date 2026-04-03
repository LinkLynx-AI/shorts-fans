import type { ReactNode } from "react";

import { FanBottomNavigation } from "@/features/fan-navigation";

export default function FanLayout({ children }: { children: ReactNode }) {
  return (
    <div className="px-0 md:px-6 md:py-6">
      <div className="relative mx-auto min-h-svh w-full max-w-[408px] overflow-hidden bg-white text-foreground shadow-[var(--device-shadow)] md:min-h-[min(880px,calc(100svh-3rem))] md:rounded-[36px]">
        <div className="relative min-h-svh overflow-hidden pb-[76px] md:min-h-[min(880px,calc(100svh-3rem))]">{children}</div>
        <div className="absolute inset-x-0 bottom-0 z-30">
          <FanBottomNavigation />
        </div>
      </div>
    </div>
  );
}
