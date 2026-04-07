import {
  getCreatorById,
  getCreatorProfileHeader,
  getCreatorProfileShortGrid,
  type CreatorProfileHeader,
  type CreatorProfileShortGridItem,
} from "@/entities/creator";
import { getShortById, listShorts } from "@/entities/short";
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

const creatorProfileShortRouteAliasById: Readonly<Record<string, string>> = {
  short_mina_hotel_mirror: "mirror",
  short_mina_rooftop: "rooftop",
};

function isNotFoundApiError(error: unknown): error is ApiError {
  return error instanceof ApiError && error.code === "http" && error.status === 404;
}

function resolveCreatorProfileShortRouteId(item: CreatorProfileShortGridItem): string {
  const aliasedShortId = creatorProfileShortRouteAliasById[item.id];

  if (aliasedShortId) {
    return aliasedShortId;
  }

  if (getShortById(item.id)) {
    return item.id;
  }

  const resolvedCreatorID = getCreatorById(item.creatorId)?.id;

  if (!resolvedCreatorID) {
    return item.id;
  }

  const matchingShorts = listShorts().filter((short) =>
    getCreatorById(short.creatorId)?.id === resolvedCreatorID &&
    short.previewDurationSeconds === item.previewDurationSeconds,
  );

  const [matchingShort] = matchingShorts;

  return matchingShorts.length === 1 && matchingShort ? matchingShort.id : item.id;
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
        routeShortId: resolveCreatorProfileShortRouteId(item),
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
