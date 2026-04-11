"use client";

import type {
  ApprovedCreatorWorkspaceManagedItem,
  ApprovedCreatorWorkspaceManagedTab,
  ApprovedCreatorWorkspacePoster,
} from "../model/approved-creator-workspace";
import {
  createPosterStyle,
  getManagedTileFrameClassName,
  getManagedTileOverlayClassName,
  getManagedTileStatusClassName,
} from "../lib/creator-mode-shell-ui";
import type { CreatorWorkspaceDetailSelection } from "./creator-mode-shell.types";

export function CreatorWorkspaceManagedTile({
  item,
  onOpenDetail,
  poster,
  tab,
}: {
  item: ApprovedCreatorWorkspaceManagedItem;
  onOpenDetail: (selection: CreatorWorkspaceDetailSelection) => void;
  poster: ApprovedCreatorWorkspacePoster;
  tab: ApprovedCreatorWorkspaceManagedTab;
}) {
  return (
    <button
      aria-label={poster.title}
      className="relative overflow-hidden rounded-[4px] text-left transition"
      onClick={() => {
        onOpenDetail({ kind: "mock", shortId: item.shortId, tab });
      }}
      type="button"
    >
      <span
        aria-hidden="true"
        className={`block aspect-[3/4] bg-[linear-gradient(180deg,var(--creator-workspace-tile-top),var(--creator-workspace-tile-mid)_42%,var(--creator-workspace-tile-bottom)_100%)] transition ${getManagedTileFrameClassName(item.tone)}`}
        style={createPosterStyle(poster)}
      />
      <span className={`absolute inset-0 grid place-items-center p-2 ${getManagedTileOverlayClassName(item.tone)}`}>
        {item.tone !== "approved" ? (
          <span
            className={`inline-flex min-h-8 items-center justify-center rounded-full px-4 text-[11px] font-bold uppercase tracking-[0.16em] backdrop-blur-[8px] ${getManagedTileStatusClassName(item.tone)}`}
          >
            {item.status}
          </span>
        ) : null}
      </span>
    </button>
  );
}
