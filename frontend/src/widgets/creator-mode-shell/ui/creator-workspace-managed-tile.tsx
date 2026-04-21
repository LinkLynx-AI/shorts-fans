"use client";

import { X } from "lucide-react";

import type {
  ApprovedCreatorWorkspaceManagedItem,
  ApprovedCreatorWorkspaceManagedTab,
  ApprovedCreatorWorkspacePoster,
} from "../model/approved-creator-workspace";
import { createPosterStyle } from "../lib/creator-mode-shell-ui";
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
  const isRejected = item.tone === "removed";
  const statusLabel = item.tone === "approved" ? "" : ` ${item.status}`;

  return (
    <button
      aria-label={`${tab === "main" ? "本編" : "ショート"}詳細を開く ${poster.shortId}${statusLabel}`}
      className="group min-w-0 rounded-[18px] text-left transition hover:opacity-95 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-[#1082c8]/20"
      onClick={() => {
        onOpenDetail({ kind: "mock", shortId: item.shortId, tab });
      }}
      type="button"
    >
      <span className="relative block overflow-hidden rounded-[18px] border border-[rgba(7,19,29,0.06)] bg-[#eef4f7] shadow-[0_8px_22px_rgba(36,92,129,0.08)]">
        <span
          aria-hidden="true"
          className={`block aspect-[3/4] bg-[linear-gradient(180deg,var(--creator-workspace-tile-top),var(--creator-workspace-tile-mid)_42%,var(--creator-workspace-tile-bottom)_100%)] transition duration-500 group-hover:scale-[1.025] ${
            isRejected ? "brightness-[0.62] saturate-[0.74]" : ""
          }`}
          style={createPosterStyle(poster)}
        />
        {isRejected ? (
          <span
            aria-hidden="true"
            className="absolute inset-0 grid place-items-center bg-black/10"
            data-testid="creator-workspace-rejected-marker"
          >
            <span className="inline-flex size-12 items-center justify-center rounded-full bg-black/40 text-white shadow-[0_10px_24px_rgba(0,0,0,0.24)] backdrop-blur-[8px]">
              <X className="size-7" strokeWidth={2.3} />
            </span>
          </span>
        ) : null}
      </span>
    </button>
  );
}
