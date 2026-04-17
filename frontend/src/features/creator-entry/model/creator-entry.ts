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

export function getCreatorEntryErrorCode(error: unknown): string | null {
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

  if (code === "auth_required") {
    return "ログイン状態を確認してから再度お試しください。";
  }
  if (code === "not_found") {
    return "プロフィールが見つかりませんでした。profile settings を確認してください。";
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
  if (code === "invalid_legal_name") {
    return "本人確認に使う氏名を入力してください。";
  }
  if (code === "invalid_birth_date") {
    return "生年月日は `YYYY-MM-DD` 形式で入力してください。";
  }
  if (code === "invalid_payout_recipient_type") {
    return "受取人種別を選択してください。";
  }
  if (code === "invalid_payout_recipient_name") {
    return "売上受取名義を入力してください。";
  }
  if (code === "registration_incomplete") {
    return "必須項目と証跡をそろえてから申請してください。";
  }
  if (code === "registration_state_conflict") {
    return "現在の申請状態ではこの操作を実行できません。";
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
    return "avatar upload の有効期限が切れました。もう一度申し込んでください。";
  }
  if (code === "invalid_evidence_kind") {
    return "この証跡タイプは現在アップロードできません。";
  }
  if (code === "invalid_evidence_mime_type") {
    return "証跡は JPEG / PNG / WebP / PDF のみ選択できます。";
  }
  if (code === "invalid_evidence_file_size") {
    return "証跡ファイルを読み取れませんでした。別のファイルを選択してください。";
  }
  if (code === "evidence_file_too_large") {
    return "証跡ファイルは 10MB 以下で選択してください。";
  }
  if (code === "evidence_upload_not_found" || code === "evidence_upload_incomplete") {
    return "証跡アップロードを確認できませんでした。もう一度お試しください。";
  }
  if (code === "evidence_upload_expired") {
    return "証跡アップロードの有効期限が切れました。もう一度ファイルを選択してください。";
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

/**
 * fan mode 遷移失敗時の UI message を返す。
 */
export function getFanModeEntryErrorMessage(error: unknown): string {
  if (error instanceof ApiError && error.code === "network") {
    return "fan mode に戻れませんでした。通信状態を確認してから再度お試しください。";
  }

  return "fan mode に戻れませんでした。少し時間を置いてからやり直してください。";
}
