import type {
  CreatorWorkspacePreviewMainItem,
  CreatorWorkspacePreviewShortItem,
} from "../api/get-creator-workspace-preview-collections";
import type {
  CreatorWorkspacePreviewMainDetail,
  CreatorWorkspacePreviewShortDetail,
} from "../api/get-creator-workspace-preview-detail";
import type {
  ApprovedCreatorWorkspaceDetailMetric,
  ApprovedCreatorWorkspaceDetailSetting,
  ApprovedCreatorWorkspaceManagedItemTone,
  ApprovedCreatorWorkspaceManagedTab,
  ApprovedCreatorWorkspacePoster,
} from "../model/approved-creator-workspace";
import type { CreatorWorkspacePreviewCollectionsState } from "../model/creator-workspace-preview-collections";

export type CreatorWorkspaceDetailSelection = {
  kind: "mock";
  shortId: string;
  tab: ApprovedCreatorWorkspaceManagedTab;
};

export type CreatorWorkspacePreviewDetailSelection =
  | {
      id: string;
      kind: "preview-main";
      tab: "main";
    }
  | {
      id: string;
      kind: "preview-short";
      tab: "shorts";
    };

export type CreatorWorkspacePreviewDetailData =
  | {
      detail: CreatorWorkspacePreviewMainDetail;
      kind: "preview-main";
    }
  | {
      detail: CreatorWorkspacePreviewShortDetail;
      kind: "preview-short";
    };

export type CreatorWorkspaceLinkedPreviewItems =
  readonly (CreatorWorkspacePreviewMainItem | CreatorWorkspacePreviewShortItem)[];

export type CreatorWorkspaceDetailViewSelection =
  | CreatorWorkspaceDetailSelection
  | CreatorWorkspacePreviewDetailSelection;

export type CreatorWorkspaceResolvedDetailState = {
  durationLabel: string;
  kindLabel: string;
  linkedMainShortId: string | null;
  linkedShortIds: readonly string[];
  metrics: readonly ApprovedCreatorWorkspaceDetailMetric[];
  settings: readonly ApprovedCreatorWorkspaceDetailSetting[];
  statusLabel: string | null;
  statusTone: ApprovedCreatorWorkspaceManagedItemTone | null;
  summary: string;
};

export type CreatorWorkspaceDetailPoster =
  | {
      kind: "mock";
      poster: ApprovedCreatorWorkspacePoster;
    }
  | {
      kind: "preview";
      posterUrl: string;
    };

export type CreatorWorkspaceReadyPreviewCollections =
  Extract<CreatorWorkspacePreviewCollectionsState, { kind: "ready" }>["collections"];
