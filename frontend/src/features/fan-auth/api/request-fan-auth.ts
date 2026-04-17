import { z } from "zod";

import { ApiError, createApiUrl } from "@/shared/api";
import { getClientEnv } from "@/shared/config";

import {
  fanAuthAcceptedNextStepSchema,
  FanAuthApiError,
  fanAuthErrorCodeSchema,
  type FanAuthAcceptedNextStep,
} from "../model/fan-auth";

const fanAuthAcceptedStepResponseSchema = z.object({
  data: z.object({
    deliveryDestinationHint: z.string().min(1).nullable(),
    nextStep: fanAuthAcceptedNextStepSchema,
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

type FanAuthRequestOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

type SignInFanInput = {
  email: string;
  password: string;
};

type SignUpFanInput = {
  displayName: string;
  email: string;
  handle: string;
  password: string;
};

type ConfirmFanSignUpInput = {
  confirmationCode: string;
  email: string;
};

type StartFanPasswordResetInput = {
  email: string;
};

type ConfirmFanPasswordResetInput = {
  confirmationCode: string;
  email: string;
  newPassword: string;
};

type ReAuthenticateFanInput = {
  password: string;
};

export type FanAuthAcceptedStep = {
  deliveryDestinationHint: string | null;
  nextStep: FanAuthAcceptedNextStep;
};

function getFanAuthBaseUrl(baseUrl?: string): string {
  return baseUrl ?? getClientEnv().NEXT_PUBLIC_API_BASE_URL;
}

async function sendFanAuthRequest(
  path: `/${string}`,
  {
    baseUrl,
    fetcher = fetch,
  }: FanAuthRequestOptions,
  init: {
    body?: Record<string, string>;
    method?: "DELETE" | "POST";
  },
): Promise<Response> {
  const requestInit: RequestInit = {
    credentials: "include",
    headers: init.body
      ? {
          Accept: "application/json",
          "Content-Type": "application/json",
        }
      : {
          Accept: "application/json",
        },
    method: init.method ?? "POST",
    ...(init.body
      ? {
          body: JSON.stringify(init.body),
        }
      : {}),
  };

  try {
    return await fetcher(createApiUrl(getFanAuthBaseUrl(baseUrl), path), requestInit);
  } catch (error) {
    throw new ApiError("API request failed before a response was received.", {
      cause: error,
      code: "network",
    });
  }
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

async function expectNoContent(
  path: `/${string}`,
  options: FanAuthRequestOptions,
  body: Record<string, string>,
): Promise<void> {
  const response = await sendFanAuthRequest(path, options, {
    body,
    method: "POST",
  });

  if (!response.ok) {
    throw await parseFanAuthError(response);
  }
}

async function expectAcceptedStep(
  path: `/${string}`,
  options: FanAuthRequestOptions,
  body: Record<string, string>,
): Promise<FanAuthAcceptedStep> {
  const response = await sendFanAuthRequest(path, options, {
    body,
    method: "POST",
  });

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

  const parsed = fanAuthAcceptedStepResponseSchema.safeParse(payload);

  if (!parsed.success) {
    throw new ApiError("API response body did not match the expected schema.", {
      cause: parsed.error,
      code: "parse",
      details: parsed.error.message,
      status: response.status,
    });
  }

  return parsed.data.data;
}

/**
 * email/password で fan sign in を完了する。
 */
export async function signInFan(
  input: SignInFanInput,
  options: FanAuthRequestOptions = {},
): Promise<void> {
  await expectNoContent("/api/fan/auth/sign-in", options, {
    email: input.email,
    password: input.password,
  });
}

/**
 * fan sign up を開始し、確認コード入力待ち step を返す。
 */
export async function signUpFan(
  input: SignUpFanInput,
  options: FanAuthRequestOptions = {},
): Promise<FanAuthAcceptedStep> {
  return expectAcceptedStep("/api/fan/auth/sign-up", options, {
    displayName: input.displayName,
    email: input.email,
    handle: input.handle,
    password: input.password,
  });
}

/**
 * sign-up confirmation code を消費して fan session を開始する。
 */
export async function confirmFanSignUp(
  input: ConfirmFanSignUpInput,
  options: FanAuthRequestOptions = {},
): Promise<void> {
  await expectNoContent("/api/fan/auth/sign-up/confirm", options, {
    confirmationCode: input.confirmationCode,
    email: input.email,
  });
}

/**
 * password reset の確認コード送信を開始する。
 */
export async function startFanPasswordReset(
  input: StartFanPasswordResetInput,
  options: FanAuthRequestOptions = {},
): Promise<FanAuthAcceptedStep> {
  return expectAcceptedStep("/api/fan/auth/password-reset", options, {
    email: input.email,
  });
}

/**
 * password reset を confirmation code と新 password で完了する。
 */
export async function confirmFanPasswordReset(
  input: ConfirmFanPasswordResetInput,
  options: FanAuthRequestOptions = {},
): Promise<void> {
  await expectNoContent("/api/fan/auth/password-reset/confirm", options, {
    confirmationCode: input.confirmationCode,
    email: input.email,
    newPassword: input.newPassword,
  });
}

/**
 * current session の fresh re-auth を更新する。
 */
export async function reAuthenticateFan(
  input: ReAuthenticateFanInput,
  options: FanAuthRequestOptions = {},
): Promise<void> {
  await expectNoContent("/api/fan/auth/re-auth", options, {
    password: input.password,
  });
}
