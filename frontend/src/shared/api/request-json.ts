import { z } from "zod";

import { getClientEnv } from "@/shared/config";

import { ApiError } from "./errors";

export type RequestJsonOptions<TSchema extends z.ZodTypeAny> = {
  baseUrl?: string;
  fetcher?: typeof fetch;
  init?: RequestInit;
  path: `/${string}` | URL;
  schema: TSchema;
};

/**
 * API のベース URL と path から絶対 URL を組み立てる。
 */
export function createApiUrl(baseUrl: string, path: `/${string}` | URL): URL {
  if (path instanceof URL) {
    return path;
  }

  const normalizedBaseUrl = baseUrl.endsWith("/") ? baseUrl : `${baseUrl}/`;

  return new URL(path.slice(1), normalizedBaseUrl);
}

/**
 * JSON API 応答を取得して Zod で検証する。
 */
export async function requestJson<TSchema extends z.ZodTypeAny>({
  baseUrl = getClientEnv().NEXT_PUBLIC_API_BASE_URL,
  fetcher = fetch,
  init,
  path,
  schema,
}: RequestJsonOptions<TSchema>): Promise<z.output<TSchema>> {
  const headers = new Headers(init?.headers);
  headers.set("Accept", "application/json");

  let response: Response;

  try {
    response = await fetcher(createApiUrl(baseUrl, path), {
      ...init,
      headers,
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }

  if (!response.ok) {
    const details = await response.text().catch(() => response.statusText);

    throw new ApiError("API request failed with a non-success status.", {
      code: "http",
      details,
      status: response.status,
    });
  }

  let payload: unknown;

  try {
    payload = response.status === 204 ? undefined : ((await response.json()) as unknown);
  } catch (error) {
    throw new ApiError("API response body was not valid JSON.", {
      cause: error,
      code: "parse",
    });
  }

  const parsed = schema.safeParse(payload);

  if (!parsed.success) {
    throw new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
    });
  }

  return parsed.data;
}
