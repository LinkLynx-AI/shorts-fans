"use client";

import { useFanAuthDialogControls } from "@/features/fan-auth";
import { Button } from "@/shared/ui";

/**
 * following feed の auth-required fallback から shared auth modal を開く。
 */
export function FeedAuthRequiredCtaButton() {
  const { openFanAuthDialog } = useFanAuthDialogControls();

  return (
    <Button
      className="h-11 border border-white/18 bg-white text-foreground shadow-[0_16px_32px_rgba(255,255,255,0.18)] hover:bg-white/94"
      onClick={() => {
        openFanAuthDialog();
      }}
      size="sm"
      type="button"
    >
      ログインして続ける
    </Button>
  );
}
