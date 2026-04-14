import { z } from "zod";

import { ApiError } from "@/shared/api";

export const fanLoginPath = "/login" as const;
export const fanAuthModes = ["sign-in", "sign-up"] as const;

export const fanAuthErrorCodeSchema = z.enum([
  "email_already_registered",
  "email_not_found",
  "handle_already_taken",
  "internal_error",
  "invalid_challenge",
  "invalid_display_name",
  "invalid_email",
  "invalid_handle",
]);

const authRequiredResponseSchema = z.object({
  error: z.object({
    code: z.literal("auth_required"),
    message: z.string().min(1),
  }),
});

export type AuthRequiredResponse = z.infer<typeof authRequiredResponseSchema>;
export type FanAuthErrorCode = z.infer<typeof fanAuthErrorCodeSchema>;
export type FanAuthMode = (typeof fanAuthModes)[number];

type FanAuthApiErrorOptions = {
  requestId?: string;
  status?: number;
};

const fanAuthErrorMessages: Record<FanAuthErrorCode, string> = {
  email_already_registered: "このメールアドレスは既に登録されています。サインインに切り替えてください。",
  email_not_found: "このメールアドレスのアカウントが見つかりません。サインアップに切り替えてください。",
  handle_already_taken: "そのhandleは既に使われています。別のhandleを入力してください。",
  internal_error: "認証を完了できませんでした。少し時間を置いてからやり直してください。",
  invalid_challenge: "認証を完了できませんでした。もう一度やり直してください。",
  invalid_display_name: "表示名を入力してください。",
  invalid_email: "メールアドレスの形式を確認してください。",
  invalid_handle: "handleは英数字・`.`・`_`のみ使えます。`@`は先頭に付けても構いません。",
};

/**
 * fan auth contract error を表す。
 */
export class FanAuthApiError extends Error {
  readonly code: FanAuthErrorCode;
  readonly requestId: string | undefined;
  readonly status: number | undefined;

  constructor(code: FanAuthErrorCode, message: string, options: FanAuthApiErrorOptions = {}) {
    super(message);
    this.name = "FanAuthApiError";
    this.code = code;
    this.requestId = options.requestId;
    this.status = options.status;
  }
}

/**
 * fan login entry の route path を返す。
 */
export function buildFanLoginHref(): string {
  return fanLoginPath;
}

/**
 * payload が auth_required 応答かを判定する。
 */
export function isAuthRequiredResponse(value: unknown): value is AuthRequiredResponse {
  return authRequiredResponseSchema.safeParse(value).success;
}

/**
 * API error が auth_required 応答を含むかを判定する。
 */
export function isAuthRequiredApiError(error: unknown): boolean {
  if (!(error instanceof ApiError) || error.status !== 401 || !error.details) {
    return false;
  }

  try {
    return isAuthRequiredResponse(JSON.parse(error.details) as unknown);
  } catch {
    return false;
  }
}

/**
 * fan auth mode を切り替えたときの補助ラベルを返す。
 */
export function getFanAuthModeSwitchLabel(mode: FanAuthMode): string {
  return mode === "sign-in" ? "サインアップへ" : "サインインへ";
}

/**
 * fan auth mode に対応する primary action 文言を返す。
 */
export function getFanAuthSubmitLabel(mode: FanAuthMode, isSubmitting: boolean): string {
  if (isSubmitting) {
    return mode === "sign-in" ? "サインイン中..." : "登録中...";
  }

  return mode === "sign-in" ? "サインインを続ける" : "新規登録を続ける";
}

/**
 * fan auth mode に対応する補足文言を返す。
 */
export function getFanAuthModeHint(mode: FanAuthMode): string {
  return mode === "sign-in"
    ? "アカウントがまだない場合"
    : "すでに登録済みの場合";
}

/**
 * fan auth contract error を UI 表示用メッセージへ変換する。
 */
export function getFanAuthErrorMessage(code: FanAuthErrorCode): string {
  return fanAuthErrorMessages[code];
}

/**
 * fan logout failure を UI 表示用メッセージへ変換する。
 */
export function getFanLogoutErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return "ログアウトできませんでした。通信状態を確認してから再度お試しください。";
  }

  return "ログアウトできませんでした。少し時間を置いてからやり直してください。";
}
