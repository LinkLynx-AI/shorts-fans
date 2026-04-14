import { z } from "zod";

import { ApiError } from "@/shared/api";

export const viewerProfileErrorCodeSchema = z.enum([
  "auth_required",
  "avatar_file_too_large",
  "avatar_upload_expired",
  "avatar_upload_incomplete",
  "avatar_upload_not_found",
  "creator_mode_unavailable",
  "handle_already_taken",
  "internal_error",
  "invalid_avatar_file_size",
  "invalid_avatar_mime_type",
  "invalid_avatar_upload_token",
  "invalid_display_name",
  "invalid_handle",
  "not_found",
]);

const viewerProfileErrorResponseSchema = z.object({
  error: z.object({
    code: viewerProfileErrorCodeSchema,
    message: z.string().min(1),
  }),
  meta: z.object({
    requestId: z.string().min(1),
  }),
});

export type ViewerProfileErrorCode = z.infer<typeof viewerProfileErrorCodeSchema>;
export type ViewerProfileInitialValues = {
  avatarUrl: string | null;
  displayName: string;
  handle: string;
};

export function getViewerProfileErrorCode(error: unknown): ViewerProfileErrorCode | null {
  if (!(error instanceof ApiError) || !error.details) {
    return null;
  }

  try {
    const payload = JSON.parse(error.details) as unknown;
    const parsed = viewerProfileErrorResponseSchema.safeParse(payload);

    if (!parsed.success) {
      return null;
    }

    return parsed.data.error.code;
  } catch {
    return null;
  }
}

export function getViewerProfileSaveErrorMessage(error: unknown): string {
  const code = getViewerProfileErrorCode(error);

  if (code === "auth_required") {
    return "ログイン状態を確認してから再度お試しください。";
  }
  if (code === "invalid_display_name") {
    return "表示名を入力してください。";
  }
  if (code === "invalid_handle") {
    return "handleは英数字・`.`・`_`のみ使えます。`@`は先頭に付けても構いません。";
  }
  if (code === "handle_already_taken") {
    return "そのhandleは既に使われています。別のhandleを入力してください。";
  }
  if (code === "invalid_avatar_mime_type") {
    return "avatar は JPEG / PNG / WebP のみ選択できます。";
  }
  if (code === "invalid_avatar_file_size") {
    return "avatar file を読み取れませんでした。別の画像を選択してください。";
  }
  if (code === "avatar_file_too_large") {
    return "avatar は 5MB 以下の画像を選択してください。";
  }
  if (code === "avatar_upload_not_found" || code === "avatar_upload_incomplete") {
    return "avatar のアップロードを確認できませんでした。もう一度お試しください。";
  }
  if (code === "avatar_upload_expired" || code === "invalid_avatar_upload_token") {
    return "avatar upload の有効期限が切れました。もう一度設定してください。";
  }
  if (code === "creator_mode_unavailable") {
    return "creator mode の状態を確認してから再度お試しください。";
  }
  if (code === "not_found") {
    return "プロフィールを読み込めませんでした。少し時間を置いてから再度お試しください。";
  }
  if (error instanceof ApiError && error.code === "network") {
    return "プロフィールを保存できませんでした。通信状態を確認してから再度お試しください。";
  }

  return "プロフィールを保存できませんでした。少し時間を置いてからやり直してください。";
}
