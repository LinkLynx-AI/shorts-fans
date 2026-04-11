import {
  getCreatorWorkspacePreviewMains,
  getCreatorWorkspacePreviewShorts,
  type CreatorWorkspacePreviewMainItem,
  type CreatorWorkspacePreviewMainListPage,
  type CreatorWorkspacePreviewShortItem,
  type CreatorWorkspacePreviewShortListPage,
} from "../api/get-creator-workspace-preview-collections";
import type { CreatorWorkspacePreviewCollections } from "./creator-workspace-preview-collections";

type CreatorWorkspacePreviewPage<TItem> = {
  items: readonly TItem[];
  page: {
    hasNext: boolean;
    nextCursor: string | null;
  };
  requestId: string;
};

/**
 * creator workspace preview list の全ページを取得する。
 */
async function loadAllCreatorWorkspacePreviewPages<TItem>(
  loadPage: (cursor?: string) => Promise<CreatorWorkspacePreviewPage<TItem>>,
): Promise<CreatorWorkspacePreviewPage<TItem>> {
  const items: TItem[] = [];
  let cursor: string | undefined;
  let firstRequestID: string | null = null;

  for (;;) {
    const page = await loadPage(cursor);

    if (firstRequestID === null) {
      firstRequestID = page.requestId;
    }

    items.push(...page.items);

    if (!page.page.hasNext || page.page.nextCursor === null) {
      return {
        items,
        page: page.page,
        requestId: firstRequestID,
      };
    }

    cursor = page.page.nextCursor;
  }
}

async function loadAllCreatorWorkspacePreviewShortPages(signal?: AbortSignal): Promise<CreatorWorkspacePreviewShortListPage> {
  return loadAllCreatorWorkspacePreviewPages<CreatorWorkspacePreviewShortItem>((cursor) => (
    getCreatorWorkspacePreviewShorts({
      credentials: "include",
      ...(cursor ? { cursor } : {}),
      ...(signal ? { signal } : {}),
    })
  ));
}

async function loadAllCreatorWorkspacePreviewMainPages(signal?: AbortSignal): Promise<CreatorWorkspacePreviewMainListPage> {
  return loadAllCreatorWorkspacePreviewPages<CreatorWorkspacePreviewMainItem>((cursor) => (
    getCreatorWorkspacePreviewMains({
      credentials: "include",
      ...(cursor ? { cursor } : {}),
      ...(signal ? { signal } : {}),
    })
  ));
}

/**
 * creator workspace 下側一覧に必要な `shorts` / `main` の全ページを取得する。
 */
export async function loadCreatorWorkspacePreviewCollections(
  signal?: AbortSignal,
): Promise<CreatorWorkspacePreviewCollections> {
  const [shorts, mains] = await Promise.all([
    loadAllCreatorWorkspacePreviewShortPages(signal),
    loadAllCreatorWorkspacePreviewMainPages(signal),
  ]);

  return {
    mains,
    shorts,
  };
}
