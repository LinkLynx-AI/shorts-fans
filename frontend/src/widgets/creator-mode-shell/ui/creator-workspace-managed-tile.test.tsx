import {
  render,
  screen,
} from "@testing-library/react";

import type {
  ApprovedCreatorWorkspaceManagedItem,
  ApprovedCreatorWorkspacePoster,
} from "../model/approved-creator-workspace";
import { CreatorWorkspaceManagedTile } from "./creator-workspace-managed-tile";

function buildManagedItem(): ApprovedCreatorWorkspaceManagedItem {
  return {
    detail: "差し戻し内容を確認してください",
    metric: "",
    shortId: "short_rejected",
    status: "公開不可",
    tone: "removed",
  };
}

function buildPoster(): ApprovedCreatorWorkspacePoster {
  return {
    shortId: "short_rejected",
    tile: {
      bottom: "#1f2937",
      mid: "#64748b",
      top: "#dbeafe",
    },
  };
}

describe("CreatorWorkspaceManagedTile", () => {
  it("exposes non-approved review status in the accessible label", () => {
    render(
      <CreatorWorkspaceManagedTile
        item={buildManagedItem()}
        onOpenDetail={() => {}}
        poster={buildPoster()}
        tab="shorts"
      />,
    );

    expect(screen.getByRole("button", { name: "ショート詳細を開く short_rejected 公開不可" })).toBeInTheDocument();
  });
});
