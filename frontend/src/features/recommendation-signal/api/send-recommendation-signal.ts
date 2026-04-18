import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import type { RecommendationSignalInput } from "../model/recommendation-signal";

type SendRecommendationSignalOptions = {
  baseUrl?: string | undefined;
  fetcher?: typeof fetch | undefined;
  input: RecommendationSignalInput;
};

/**
 * recommendation signal を non-blocking 前提で backend に送る。
 */
export async function sendRecommendationSignal({
  baseUrl = getClientEnv().NEXT_PUBLIC_API_BASE_URL,
  fetcher = fetch,
  input,
}: SendRecommendationSignalOptions): Promise<void> {
  let response: Response;

  try {
    response = await fetcher(createApiUrl(baseUrl, "/api/fan/recommendation/events"), {
      body: JSON.stringify(input),
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      keepalive: true,
      method: "POST",
    });
  } catch (error) {
    throw new ApiError("Recommendation signal request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (response.ok) {
    return;
  }

  const details = await response.text().catch(() => response.statusText);

  throw new ApiError("Recommendation signal request failed with a non-success status.", {
    code: "http",
    details,
    status: response.status,
  });
}
