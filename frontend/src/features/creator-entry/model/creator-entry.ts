import { z } from "zod";

import { ApiError } from "@/shared/api";

const creatorEntryErrorResponseSchema = z.object({
  error: z.object({
    code: z.string().min(1),
    message: z.string().min(1),
  }),
  meta: z.object({
    requestId: z.string().min(1),
  }),
});

function getCreatorEntryErrorCode(error: unknown): string | null {
  if (!(error instanceof ApiError) || !error.details) {
    return null;
  }

  try {
    const payload = JSON.parse(error.details) as unknown;
    const parsed = creatorEntryErrorResponseSchema.safeParse(payload);

    if (!parsed.success) {
      return null;
    }

    return parsed.data.error.code;
  } catch {
    return null;
  }
}

/**
 * creator registration 失敗時の UI message を返す。
 */
export function getCreatorRegistrationErrorMessage(error: unknown): string {
  const code = getCreatorEntryErrorCode(error);

  if (code === "invalid_display_name") {
    return "表示名を入力してください。";
  }

  if (error instanceof ApiError && error.code === "network") {
    return "登録を完了できませんでした。通信状態を確認してから再度お試しください。";
  }

  return "登録を完了できませんでした。少し時間を置いてからやり直してください。";
}

/**
 * creator mode 遷移失敗時の UI message を返す。
 */
export function getCreatorModeEntryErrorMessage(error: unknown): string {
  const code = getCreatorEntryErrorCode(error);

  if (code === "creator_mode_unavailable") {
    return "creator mode を開けませんでした。状態反映を確認してからもう一度お試しください。";
  }

  if (error instanceof ApiError && error.code === "network") {
    return "creator mode を開けませんでした。通信状態を確認してから再度お試しください。";
  }

  return "creator mode を開けませんでした。少し時間を置いてからやり直してください。";
}
