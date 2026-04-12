import { z } from "zod";

import { creatorSummarySchema } from "@/entities/creator";
import { publicShortSummarySchema } from "@/entities/short";
import { requestJson } from "@/shared/api";

import type { FanLibraryItem } from "../model/fan-profile";
import {
  buildFanProfileCursorPath,
  buildFanProfileRequestHeaders,
  fanProfileCursorMetaSchema,
  type FanProfileCursorPage,
} from "./shared";

const mainAccessStateSchema = z.object({
  mainId: z.string().min(1),
  reason: z.enum(["owner_preview", "session_unlocked"]),
  status: z.enum(["owner", "unlocked"]),
});

const fanProfileLibraryItemSchema = z.object({
  access: mainAccessStateSchema,
  creator: creatorSummarySchema,
  entryShort: publicShortSummarySchema,
  main: z.object({
    durationSeconds: z.number().int().positive(),
    id: z.string().min(1),
  }),
});

const fanProfileLibraryResponseSchema = z.object({
  data: z.object({
    items: z.array(fanProfileLibraryItemSchema),
  }),
  error: z.null(),
  meta: fanProfileCursorMetaSchema,
});

export type FanProfileLibraryPage = {
  items: readonly FanLibraryItem[];
  page: FanProfileCursorPage;
  requestId: string;
};

type FetchFanProfileLibraryPageOptions = {
  baseUrl?: string | undefined;
  cursor?: string | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
  signal?: AbortSignal | undefined;
};

/**
 * fan profile library API から main 一覧 1 page を取得する。
 */
export async function fetchFanProfileLibraryPage({
  baseUrl,
  cursor,
  fetcher,
  sessionToken,
  signal,
}: FetchFanProfileLibraryPageOptions = {}): Promise<FanProfileLibraryPage> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      headers: buildFanProfileRequestHeaders(sessionToken),
      ...(signal ? { signal } : {}),
    },
    path: buildFanProfileCursorPath("/api/fan/profile/library", cursor),
    schema: fanProfileLibraryResponseSchema,
  });

  return {
    items: response.data.items,
    page: response.meta.page,
    requestId: response.meta.requestId,
  };
}
