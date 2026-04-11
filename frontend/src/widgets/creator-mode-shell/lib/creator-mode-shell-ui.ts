import type { CSSProperties } from "react";

import type {
  CreatorWorkspacePreviewMainItem,
  CreatorWorkspacePreviewShortItem,
} from "../api/get-creator-workspace-preview-collections";
import type {
  ApprovedCreatorWorkspaceDetailSetting,
  ApprovedCreatorWorkspaceManagedItemTone,
  ApprovedCreatorWorkspacePoster,
} from "../model/approved-creator-workspace";

export function formatCount(value: number): string {
  return value.toLocaleString("ja-JP");
}

export function formatJpy(value: number): string {
  return `¥${value.toLocaleString("ja-JP")}`;
}

export function formatDurationLabel(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, "0")}:${seconds.toString().padStart(2, "0")}`;
  }

  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

export function buildPreviewShortAriaLabel(item: CreatorWorkspacePreviewShortItem, index: number): string {
  return `ショート詳細を開く ${index + 1}件目 ${formatDurationLabel(item.previewDurationSeconds)}`;
}

export function buildPreviewMainAriaLabel(item: CreatorWorkspacePreviewMainItem, index: number): string {
  return `本編詳細を開く ${index + 1}件目 ${formatJpy(item.priceJpy)} ${formatDurationLabel(item.durationSeconds)}`;
}

export function buildRevisionRequestedDetail({
  mainCount,
  shortCount,
}: {
  mainCount: number;
  shortCount: number;
}): string {
  const scopes = [
    shortCount > 0 ? `short ${formatCount(shortCount)}件` : null,
    mainCount > 0 ? `main ${formatCount(mainCount)}件` : null,
  ].filter((scope) => scope !== null);

  if (scopes.length === 0) {
    return "修正依頼内容を確認してください";
  }

  return `${scopes.join(" / ")}を確認してください`;
}

export function createPosterStyle(poster: ApprovedCreatorWorkspacePoster): CSSProperties {
  return {
    "--creator-workspace-tile-bottom": poster.tile.bottom,
    "--creator-workspace-tile-mid": poster.tile.mid,
    "--creator-workspace-tile-top": poster.tile.top,
  } as CSSProperties;
}

export function createVideoPosterStyle(posterUrl: string): CSSProperties {
  return {
    backgroundImage: `url("${posterUrl}")`,
    backgroundPosition: "center",
    backgroundSize: "cover",
  };
}

export function formatUnlockMetric(value: number): string {
  return `${formatCount(value)} unlocks`;
}

export function getManagedTileFrameClassName(tone: ApprovedCreatorWorkspaceManagedItemTone): string {
  if (tone === "approved") {
    return "";
  }

  return "brightness-[0.72] saturate-[0.82]";
}

export function getManagedTileOverlayClassName(tone: ApprovedCreatorWorkspaceManagedItemTone): string {
  if (tone === "approved") {
    return "bg-[linear-gradient(180deg,rgba(6,21,33,0.04)_0%,rgba(6,21,33,0.02)_38%,rgba(6,21,33,0.48)_100%)]";
  }

  return "bg-[linear-gradient(180deg,rgba(6,21,33,0.42)_0%,rgba(6,21,33,0.18)_34%,rgba(6,21,33,0.72)_100%)]";
}

export function getManagedTileStatusClassName(tone: ApprovedCreatorWorkspaceManagedItemTone): string {
  switch (tone) {
    case "hidden":
      return "bg-[rgba(7,19,29,0.18)] text-[#f6fbff]";
    case "paused":
      return "bg-[rgba(16,130,200,0.18)] text-[#eff8ff]";
    case "pending":
      return "bg-[rgba(16,130,200,0.18)] text-[#eff8ff]";
    case "removed":
      return "bg-[rgba(217,77,77,0.2)] text-[#fff7f7]";
    case "revision":
      return "bg-[rgba(244,152,45,0.18)] text-[#fff7ea]";
    case "approved":
      return "bg-[rgba(52,168,83,0.12)] text-[#effff2]";
  }
}

export function buildPreviewShortDetailSettings(
  item: CreatorWorkspacePreviewShortItem,
): readonly ApprovedCreatorWorkspaceDetailSetting[] {
  return [
    { label: "長さ", value: formatDurationLabel(item.previewDurationSeconds) },
  ];
}

export function buildPreviewMainDetailSettings(
  item: CreatorWorkspacePreviewMainItem,
): readonly ApprovedCreatorWorkspaceDetailSetting[] {
  return [
    { label: "価格", value: formatJpy(item.priceJpy) },
    { label: "長さ", value: formatDurationLabel(item.durationSeconds) },
  ];
}
