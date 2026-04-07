import { z } from "zod";

import { viewerSessionCookieName } from "@/entities/viewer";
import { requestJson } from "@/shared/api";

import type { CreatorSummary } from "../model/creator";
import {
  creatorProfileHeaderSchema,
  creatorProfileStatsSchema,
} from "./contracts";

const creatorProfileHeaderResponseSchema = z.object({
  data: z.object({
    profile: creatorProfileHeaderSchema,
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

type CreatorProfileHeaderStats = z.output<typeof creatorProfileStatsSchema>;

export type CreatorProfileHeader = {
  creator: CreatorSummary;
  stats: CreatorProfileHeaderStats;
  viewer: {
    isFollowing: boolean;
  };
};

type GetCreatorProfileHeaderOptions = {
  baseUrl?: string | undefined;
  creatorId: string;
  credentials?: RequestCredentials | undefined;
  fetcher?: typeof fetch | undefined;
  sessionToken?: string | undefined;
};

function buildCreatorProfileHeaderPath(creatorId: string): `/${string}` {
  return `/api/fan/creators/${encodeURIComponent(creatorId)}`;
}

/**
 * creator profile header を contract-backed API から取得する。
 */
export async function getCreatorProfileHeader({
  baseUrl,
  creatorId,
  credentials,
  fetcher,
  sessionToken,
}: GetCreatorProfileHeaderOptions): Promise<CreatorProfileHeader> {
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
    path: buildCreatorProfileHeaderPath(creatorId),
    schema: creatorProfileHeaderResponseSchema,
  });

  return response.data.profile;
}
