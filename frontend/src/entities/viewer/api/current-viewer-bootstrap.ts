import { z } from "zod";

import { requestJson } from "@/shared/api";

import type { CurrentViewer } from "../model/current-viewer";
import {
  viewerActiveModes,
  viewerSessionCookieName,
} from "../model/current-viewer";

const currentViewerSchema = z.object({
  activeMode: z.enum(viewerActiveModes),
  canAccessCreatorMode: z.boolean(),
  id: z.string().min(1),
});

export const currentViewerBootstrapSchema = z.object({
  data: z.object({
    currentViewer: currentViewerSchema.nullable(),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

type GetCurrentViewerBootstrapOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  sessionToken?: string;
};

/**
 * app bootstrap 用の current viewer state を取得する。
 */
export async function getCurrentViewerBootstrap({
  baseUrl,
  credentials,
  fetcher,
  sessionToken,
}: GetCurrentViewerBootstrapOptions = {}): Promise<CurrentViewer | null> {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `${viewerSessionCookieName}=${sessionToken}`);
  }

  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      ...(credentials ? { credentials } : {}),
      headers,
    },
    path: "/api/viewer/bootstrap",
    schema: currentViewerBootstrapSchema,
  });

  return response.data.currentViewer;
}
