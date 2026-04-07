import { z } from "zod";

import { createApiUrl, ApiError } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  FanAuthApiError,
  fanAuthErrorCodeSchema,
  type FanAuthMode,
} from "../model/fan-auth";

const fanAuthChallengeResponseSchema = z.object({
  data: z.object({
    challengeToken: z.string().min(1),
    expiresAt: z.string().datetime(),
  }),
  error: z.null(),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

const fanAuthErrorResponseSchema = z.object({
  data: z.null(),
  error: z.object({
    code: fanAuthErrorCodeSchema,
    message: z.string().min(1),
  }),
  meta: z.object({
    page: z.null(),
    requestId: z.string().min(1),
  }),
});

type AuthenticateFanWithEmailOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

async function sendFanAuthRequest(
  path: `/${string}`,
  body: Record<string, string>,
  options: AuthenticateFanWithEmailOptions,
): Promise<Response> {
  const fetcher = options.fetcher ?? fetch;

  try {
    return await fetcher(createApiUrl(getFanAuthBaseUrl(options.baseUrl), path), {
      body: JSON.stringify(body),
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "POST",
    });
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }
}

function getFanAuthBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

function buildFanAuthPath(mode: FanAuthMode, boundary: "challenges" | "session"): `/${string}` {
  return `/api/fan/auth/${mode}/${boundary}`;
}

async function parseFanAuthError(response: Response): Promise<FanAuthApiError | ApiError> {
  let payload: unknown;

  try {
    payload = await response.json();
  } catch (error) {
    return new ApiError("API response body was not valid JSON.", {
      cause: error,
      code: "parse",
      status: response.status,
    });
  }

  const parsed = fanAuthErrorResponseSchema.safeParse(payload);

  if (!parsed.success) {
    return new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return new FanAuthApiError(parsed.data.error.code, parsed.data.error.message, {
    requestId: parsed.data.meta.requestId,
    status: response.status,
  });
}

async function issueFanAuthChallenge(
  mode: FanAuthMode,
  email: string,
  options: AuthenticateFanWithEmailOptions,
): Promise<string> {
  const response = await sendFanAuthRequest(
    buildFanAuthPath(mode, "challenges"),
    {
      email,
    },
    options,
  );

  if (!response.ok) {
    throw await parseFanAuthError(response);
  }

  let payload: unknown;

  try {
    payload = await response.json();
  } catch (error) {
    throw new ApiError("API response body was not valid JSON.", {
      cause: error,
      code: "parse",
      status: response.status,
    });
  }

  const parsed = fanAuthChallengeResponseSchema.safeParse(payload);

  if (!parsed.success) {
    throw new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return parsed.data.data.challengeToken;
}

async function startFanAuthSession(
  mode: FanAuthMode,
  email: string,
  challengeToken: string,
  options: AuthenticateFanWithEmailOptions,
): Promise<void> {
  const response = await sendFanAuthRequest(
    buildFanAuthPath(mode, "session"),
    {
      challengeToken,
      email,
    },
    options,
  );

  if (!response.ok) {
    throw await parseFanAuthError(response);
  }
}

/**
 * fan auth contract に従って email-only の sign in / sign up を完了する。
 *
 * Contract:
 * - `POST /api/fan/auth/{mode}/challenges` と `POST /api/fan/auth/{mode}/session` を順に実行する
 * - browser cookie を受け取れるよう `credentials: "include"` で送信する
 *
 * Errors:
 * - contract error は `FanAuthApiError`
 * - schema / network / malformed response は `ApiError`
 *
 * Side effects:
 * - backend fan auth endpoint へ network request を送る
 */
export async function authenticateFanWithEmail(
  mode: FanAuthMode,
  email: string,
  options: AuthenticateFanWithEmailOptions = {},
): Promise<void> {
  const challengeToken = await issueFanAuthChallenge(mode, email, options);
  await startFanAuthSession(mode, email, challengeToken, options);
}
