import { z } from "zod";

import type { ViewerActiveMode } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

type SwitchViewerActiveModeOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

/**
 * viewer の active mode を切り替える。
 */
export async function switchViewerActiveMode(
  activeMode: ViewerActiveMode,
  options: SwitchViewerActiveModeOptions = {},
): Promise<void> {
  await requestJson({
    ...(options.baseUrl ? { baseUrl: options.baseUrl } : {}),
    ...(options.fetcher ? { fetcher: options.fetcher } : {}),
    init: {
      body: JSON.stringify({
        activeMode,
      }),
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      method: "PUT",
    },
    path: "/api/viewer/active-mode",
    schema: z.undefined(),
  });
}
