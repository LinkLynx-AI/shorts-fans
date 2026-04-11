import {
  getCreatorProfileHeader,
  getCreatorProfileShortGrid,
  type CreatorProfileHeader,
  type CreatorProfileShortGridItem,
} from "@/entities/creator";
import { ApiError } from "@/shared/api";

export type CreatorProfileShellShortItem = CreatorProfileShortGridItem & {
  routeShortId: string;
};

export type CreatorProfileShellState =
  | {
      creator: CreatorProfileHeader["creator"];
      kind: "empty";
      shorts: readonly [];
      stats: CreatorProfileHeader["stats"];
      viewer: CreatorProfileHeader["viewer"];
    }
  | {
      creator: CreatorProfileHeader["creator"];
      kind: "ready";
      shorts: readonly CreatorProfileShellShortItem[];
      stats: CreatorProfileHeader["stats"];
      viewer: CreatorProfileHeader["viewer"];
    };

type LoadCreatorProfileShellStateOptions = {
  sessionToken?: string | undefined;
};

function isNotFoundApiError(error: unknown): error is ApiError {
  return error instanceof ApiError && error.code === "http" && error.status === 404;
}

/**
 * creator profile shell 用の contract-backed state を取得する。
 */
export async function loadCreatorProfileShellState(
  creatorId: string,
  options: LoadCreatorProfileShellStateOptions = {},
): Promise<CreatorProfileShellState | undefined> {
  try {
    const [profile, shortGrid] = await Promise.all([
      getCreatorProfileHeader({
        creatorId,
        sessionToken: options.sessionToken,
      }),
      getCreatorProfileShortGrid({
        creatorId,
      }),
    ]);

    if (shortGrid.items.length === 0) {
      return {
        creator: profile.creator,
        kind: "empty",
        shorts: [],
        stats: profile.stats,
        viewer: profile.viewer,
      };
    }

    return {
      creator: profile.creator,
      kind: "ready",
      shorts: shortGrid.items.map((item) => ({
        ...item,
        routeShortId: item.id,
      })),
      stats: profile.stats,
      viewer: profile.viewer,
    };
  } catch (error) {
    if (isNotFoundApiError(error)) {
      return undefined;
    }

    throw error;
  }
}
