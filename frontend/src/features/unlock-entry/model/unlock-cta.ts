export type UnlockCtaStateType =
  | "continue_main"
  | "owner_preview"
  | "setup_required"
  | "unavailable"
  | "unlock_available";

export type UnlockCtaState = {
  mainDurationSeconds: number | null;
  priceJpy: number | null;
  resumePositionSeconds: number | null;
  state: UnlockCtaStateType;
};

function formatSecondsAsTimestamp(seconds: number): string {
  const normalized = Math.max(0, Math.floor(seconds));
  const minutes = Math.floor(normalized / 60);
  const remainingSeconds = normalized % 60;

  return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
}

function formatCompactMinutes(seconds: number): string {
  const roundedMinutes = Math.max(1, Math.round(seconds / 60));
  return `${roundedMinutes}分`;
}

function formatJpy(priceJpy: number): string {
  return `¥${priceJpy.toLocaleString("ja-JP")}`;
}

/**
 * CTA state から表示ラベルを組み立てる。
 */
export function getUnlockCtaLabel(cta: UnlockCtaState): string {
  switch (cta.state) {
    case "continue_main":
      return "Continue main";
    case "owner_preview":
      return "Owner preview";
    case "unavailable":
      return "Unavailable";
    case "setup_required":
    case "unlock_available":
      return "Unlock";
  }
}

/**
 * CTA state から右側の補助メタ表示を組み立てる。
 */
export function getUnlockCtaMeta(cta: UnlockCtaState): string | null {
  switch (cta.state) {
    case "continue_main":
      return cta.resumePositionSeconds === null ? null : formatSecondsAsTimestamp(cta.resumePositionSeconds);
    case "setup_required":
    case "unlock_available":
      if (cta.priceJpy === null || cta.mainDurationSeconds === null) {
        return null;
      }

      return `${formatJpy(cta.priceJpy)} | ${formatCompactMinutes(cta.mainDurationSeconds)}`;
    case "owner_preview":
    case "unavailable":
      return null;
  }
}
