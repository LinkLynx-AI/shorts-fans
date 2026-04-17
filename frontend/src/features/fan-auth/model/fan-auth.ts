import { z } from "zod";

import { ApiError } from "@/shared/api";

export const fanLoginPath = "/login" as const;
export const fanAuthModes = [
  "sign-in",
  "sign-up",
  "confirm-sign-up",
  "password-reset-request",
  "confirm-password-reset",
  "re-auth",
] as const;
export const fanAuthAcceptedNextSteps = [
  "confirm_sign_up",
  "confirm_password_reset",
] as const;

export const fanAuthAcceptedNextStepSchema = z.enum(fanAuthAcceptedNextSteps);
export const fanAuthErrorCodeSchema = z.enum([
  "auth_required",
  "confirmation_code_expired",
  "confirmation_required",
  "fresh_auth_required",
  "handle_already_taken",
  "internal_error",
  "invalid_confirmation_code",
  "invalid_credentials",
  "invalid_display_name",
  "invalid_email",
  "invalid_handle",
  "invalid_password",
  "password_policy_violation",
  "rate_limited",
]);

const authRequiredResponseSchema = z.object({
  error: z.object({
    code: z.literal("auth_required"),
    message: z.string().min(1),
  }),
});

const freshAuthRequiredResponseSchema = z.object({
  error: z.object({
    code: z.literal("fresh_auth_required"),
    message: z.string().min(1),
  }),
});

export type AuthRequiredResponse = z.infer<typeof authRequiredResponseSchema>;
export type FreshAuthRequiredResponse = z.infer<typeof freshAuthRequiredResponseSchema>;
export type FanAuthAcceptedNextStep = z.infer<typeof fanAuthAcceptedNextStepSchema>;
export type FanAuthErrorCode = z.infer<typeof fanAuthErrorCodeSchema>;
export type FanAuthMode = (typeof fanAuthModes)[number];

type FanAuthApiErrorOptions = {
  requestId?: string;
  status?: number;
};

const fanAuthErrorMessages: Record<FanAuthErrorCode, string> = {
  auth_required: "セッションが確認できませんでした。もう一度ログインしてください。",
  confirmation_code_expired: "確認コードの有効期限が切れました。コードを再送してやり直してください。",
  confirmation_required: "確認コードを入力して登録を完了してください。",
  fresh_auth_required: "続けるには、もう一度パスワードを入力して認証を確認してください。",
  handle_already_taken: "そのhandleは既に使われています。別のhandleを入力してください。",
  internal_error: "認証を完了できませんでした。少し時間を置いてからやり直してください。",
  invalid_confirmation_code: "確認コードを確認して、もう一度入力してください。",
  invalid_credentials: "メールアドレスまたはパスワードを確認してください。",
  invalid_display_name: "表示名を入力してください。",
  invalid_email: "メールアドレスの形式を確認してください。",
  invalid_handle: "handleは英数字・`.`・`_`のみ使えます。`@`は先頭に付けても構いません。",
  invalid_password: "パスワードを確認してください。",
  password_policy_violation: "パスワードが条件を満たしていません。より強いパスワードを入力してください。",
  rate_limited: "試行回数が上限に達しました。少し時間を置いてから再度お試しください。",
};

const fanAuthModeTitles: Record<FanAuthMode, string> = {
  "confirm-password-reset": "新しいパスワードを設定します",
  "confirm-sign-up": "確認コードを入力してください",
  "password-reset-request": "パスワードを再設定します",
  "re-auth": "認証を確認してください",
  "sign-in": "続けるにはログインが必要です",
  "sign-up": "アカウントを作成します",
};

const fanAuthModeDescriptions: Record<FanAuthMode, string> = {
  "confirm-password-reset":
    "メールで受け取った確認コードと新しいパスワードを入力すると、サインインへ戻れます。",
  "confirm-sign-up":
    "メールで受け取った確認コードを入力すると、fan session を開始して元の文脈へ戻れます。",
  "password-reset-request":
    "登録済みメールアドレスに確認コードを送って、modal のままパスワードを更新します。",
  "re-auth":
    "高い権限が必要な操作を続ける前に、現在の fan session をもう一度確認します。",
  "sign-in":
    "fan session を開始すると、フォロー中、library、main 再生のような protected surface をそのまま続けられます。",
  "sign-up":
    "email とプロフィール初期値を入力すると、確認コードの入力に進みます。",
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
 * payload が fresh_auth_required 応答かを判定する。
 */
export function isFreshAuthRequiredResponse(value: unknown): value is FreshAuthRequiredResponse {
  return freshAuthRequiredResponseSchema.safeParse(value).success;
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
 * API error が fresh_auth_required 応答を含むかを判定する。
 */
export function isFreshAuthRequiredApiError(error: unknown): boolean {
  if (!(error instanceof ApiError) || error.status !== 403 || !error.details) {
    return false;
  }

  try {
    return isFreshAuthRequiredResponse(JSON.parse(error.details) as unknown);
  } catch {
    return false;
  }
}

/**
 * accepted auth nextStep を modal mode へ対応づける。
 */
export function mapFanAuthNextStepToMode(nextStep: FanAuthAcceptedNextStep): FanAuthMode {
  switch (nextStep) {
    case "confirm_password_reset":
      return "confirm-password-reset";
    case "confirm_sign_up":
      return "confirm-sign-up";
  }
}

/**
 * fan auth mode に対応する panel title を返す。
 */
export function getFanAuthModeTitle(mode: FanAuthMode): string {
  return fanAuthModeTitles[mode];
}

/**
 * fan auth mode に対応する補足文言を返す。
 */
export function getFanAuthModeDescription(mode: FanAuthMode): string {
  return fanAuthModeDescriptions[mode];
}

/**
 * fan auth mode に対応する primary action 文言を返す。
 */
export function getFanAuthSubmitLabel(mode: FanAuthMode, isSubmitting: boolean): string {
  if (isSubmitting) {
    switch (mode) {
      case "confirm-password-reset":
        return "更新中...";
      case "confirm-sign-up":
        return "登録中...";
      case "password-reset-request":
      case "sign-up":
        return "送信中...";
      case "re-auth":
        return "認証中...";
      case "sign-in":
        return "サインイン中...";
    }
  }

  switch (mode) {
    case "confirm-password-reset":
      return "パスワードを更新する";
    case "confirm-sign-up":
      return "登録を完了する";
    case "password-reset-request":
      return "確認コードを送る";
    case "re-auth":
      return "認証を続ける";
    case "sign-in":
      return "サインインを続ける";
    case "sign-up":
      return "確認コードを送る";
  }
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
