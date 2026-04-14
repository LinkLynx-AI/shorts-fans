import { requestJson } from "@/shared/api";

import {
  viewerProfileResponseSchema,
  type ViewerProfile,
} from "./contracts";

type GetViewerProfileOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
  sessionToken?: string;
};

function buildViewerProfileRequestHeaders(sessionToken?: string): Headers {
  const headers = new Headers();

  if (sessionToken) {
    headers.set("Cookie", `shorts_fans_session=${sessionToken}`);
  }

  return headers;
}

export async function getViewerProfile({
  baseUrl,
  credentials,
  fetcher,
  sessionToken,
}: GetViewerProfileOptions = {}): Promise<ViewerProfile> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      ...(credentials ? { credentials } : {}),
      headers: buildViewerProfileRequestHeaders(sessionToken),
    },
    path: "/api/viewer/profile",
    schema: viewerProfileResponseSchema,
  });

  return response.data.profile;
}
