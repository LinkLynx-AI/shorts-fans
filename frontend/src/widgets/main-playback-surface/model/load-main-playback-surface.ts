import { ApiError } from "@/shared/api";

import { fetchMainPlayback } from "../api/request-main-playback";
import {
  buildMainPlaybackSurface,
  type MainPlaybackSurface,
} from "./main-playback-surface";

export type LoadMainPlaybackSurfaceResult =
  | {
      kind: "auth_required";
    }
  | {
      kind: "locked";
    }
  | {
      kind: "not_found";
    }
  | {
      kind: "ready";
      surface: MainPlaybackSurface;
    };

type LoadMainPlaybackSurfaceOptions = {
  fromShortId: string;
  grant: string;
  sessionToken?: string | undefined;
};

/**
 * main playback API 応答を page 用の UI surface に変換する。
 */
export async function loadMainPlaybackSurface(
  mainId: string,
  options: LoadMainPlaybackSurfaceOptions,
): Promise<LoadMainPlaybackSurfaceResult> {
  try {
    const payload = await fetchMainPlayback({
      fromShortId: options.fromShortId,
      grant: options.grant,
      mainId,
      sessionToken: options.sessionToken,
    });
    const surface = buildMainPlaybackSurface(payload, options.fromShortId);

    if (!surface) {
      return {
        kind: "not_found",
      };
    }

    return {
      kind: "ready",
      surface,
    };
  } catch (error) {
    if (error instanceof ApiError && error.code === "http") {
      if (error.status === 401) {
        return {
          kind: "auth_required",
        };
      }

      if (error.status === 403) {
        return {
          kind: "locked",
        };
      }

      if (error.status === 404) {
        return {
          kind: "not_found",
        };
      }
    }

    throw error;
  }
}
