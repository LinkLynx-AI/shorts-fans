import type { CreatorWorkspacePreviewDetailState } from "../model/use-creator-workspace-preview-detail";

type CreatorWorkspaceDetailSummarySelection = {
  kind: "mock" | "preview-main" | "preview-short";
};

export function resolveCreatorWorkspaceDetailSummary(
  detailSelection: CreatorWorkspaceDetailSummarySelection,
  summary: string,
  previewDetailState: CreatorWorkspacePreviewDetailState,
): string | null {
  if (detailSelection.kind === "mock") {
    const normalizedSummary = summary.trim();

    return normalizedSummary.length > 0 ? normalizedSummary : null;
  }

  if (detailSelection.kind === "preview-main") {
    return null;
  }

  if (previewDetailState.kind !== "ready" || previewDetailState.detail.kind !== "preview-short") {
    return null;
  }

  const normalizedCaption = previewDetailState.detail.short.caption.trim();

  return normalizedCaption.length > 0 ? normalizedCaption : null;
}
