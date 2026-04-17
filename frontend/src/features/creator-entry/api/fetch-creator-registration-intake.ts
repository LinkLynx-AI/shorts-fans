import { requestJson } from "@/shared/api";

import {
  creatorRegistrationIntakeResponseSchema,
  type CreatorRegistrationIntake,
} from "./contracts";

type FetchCreatorRegistrationIntakeOptions = {
  baseUrl?: string;
  credentials?: RequestCredentials;
  fetcher?: typeof fetch;
};

/**
 * current viewer の creator registration intake draft を取得する。
 */
export async function fetchCreatorRegistrationIntake(
  {
    baseUrl,
    credentials = "include",
    fetcher,
  }: FetchCreatorRegistrationIntakeOptions = {},
): Promise<CreatorRegistrationIntake> {
  const response = await requestJson({
    ...(baseUrl ? { baseUrl } : {}),
    ...(fetcher ? { fetcher } : {}),
    init: {
      cache: "no-store",
      credentials,
      method: "GET",
    },
    path: "/api/viewer/creator-registration/intake",
    schema: creatorRegistrationIntakeResponseSchema,
  });

  return response.data.intake;
}
