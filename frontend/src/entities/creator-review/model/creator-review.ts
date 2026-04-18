import type {
  CreatorReviewDecision,
  CreatorReviewQueueState,
} from "../api/contracts";

export type CreatorReviewReasonOption = {
  code: string;
  description: string;
  label: string;
};

export type CreatorReviewRejectHandling = {
  description: string;
  isResubmitEligible: boolean;
  isSupportReviewRequired: boolean;
  label: string;
  value: CreatorReviewRejectHandlingMode;
};

export type CreatorReviewRejectHandlingMode =
  | "resubmit_eligible"
  | "support_review_required"
  | "manual_follow_up";

export const creatorReviewReasonOptions = [
  {
    code: "missing_documents",
    description: "必要書類が不足しているため再提出が必要です。",
    label: "必要書類不足",
  },
  {
    code: "documents_blurry",
    description: "書類画像が不鮮明で確認できません。",
    label: "書類が不鮮明",
  },
  {
    code: "payout_info_missing",
    description: "売上受取情報が不足しています。",
    label: "受取情報不足",
  },
  {
    code: "profile_mismatch",
    description: "shared profile と申請内容に整合しない点があります。",
    label: "プロフィール不一致",
  },
  {
    code: "age_requirement_mismatch",
    description: "年齢要件を満たしていない、または確認できません。",
    label: "年齢要件不一致",
  },
  {
    code: "fraud_suspected",
    description: "不正利用の疑いがあるため審査を通過できません。",
    label: "不正利用懸念",
  },
  {
    code: "prohibited_category",
    description: "取り扱い対象外カテゴリに該当する可能性があります。",
    label: "禁止カテゴリ",
  },
  {
    code: "consent_or_ownership_issue",
    description: "同意または権利確認に不足があります。",
    label: "同意・権利確認不足",
  },
  {
    code: "safety_high_risk",
    description: "安全性観点のリスクが高く、追加対応が必要です。",
    label: "安全性高リスク",
  },
] as const satisfies readonly CreatorReviewReasonOption[];

const supportReviewReasonCodes = new Set<string>([
  "age_requirement_mismatch",
  "fraud_suspected",
  "prohibited_category",
  "consent_or_ownership_issue",
  "safety_high_risk",
]);

export const creatorReviewRejectHandlingOptions = [
  {
    description: "creator が onboarding flow から修正して再提出できます。",
    isResubmitEligible: true,
    isSupportReviewRequired: false,
    label: "再提出を許可",
    value: "resubmit_eligible",
  },
  {
    description: "support / manual review 完了まで self-serve resubmit を止めます。",
    isResubmitEligible: false,
    isSupportReviewRequired: true,
    label: "support review が必要",
    value: "support_review_required",
  },
  {
    description: "即時の self-serve resubmit は開かず、別導線の follow-up を前提にします。",
    isResubmitEligible: false,
    isSupportReviewRequired: false,
    label: "保留で運用する",
    value: "manual_follow_up",
  },
] as const satisfies readonly CreatorReviewRejectHandling[];

const creatorReviewQueueStateLabels: Record<CreatorReviewQueueState, string> = {
  approved: "承認済み",
  rejected: "却下済み",
  submitted: "審査待ち",
  suspended: "停止中",
};

const creatorReviewDecisionLabels: Record<CreatorReviewDecision, string> = {
  approved: "承認する",
  rejected: "却下する",
  suspended: "停止する",
};

/**
 * admin review state query を既知値へ正規化する。
 */
export function normalizeCreatorReviewState(value: string | string[] | undefined): CreatorReviewQueueState {
  const candidate = Array.isArray(value) ? value[0] : value;
  switch (candidate) {
    case "approved":
    case "rejected":
    case "submitted":
    case "suspended":
      return candidate;
    default:
      return "submitted";
  }
}

/**
 * admin review queue state の表示ラベルを返す。
 */
export function getCreatorReviewStateLabel(state: CreatorReviewQueueState): string {
  return creatorReviewQueueStateLabels[state];
}

/**
 * admin review decision の表示ラベルを返す。
 */
export function getCreatorReviewDecisionLabel(decision: CreatorReviewDecision): string {
  return creatorReviewDecisionLabels[decision];
}

/**
 * 現在 state から許可される decision を返す。
 */
export function getCreatorReviewAvailableDecisions(
  state: CreatorReviewQueueState,
): readonly CreatorReviewDecision[] {
  switch (state) {
    case "submitted":
      return ["approved", "rejected"];
    case "approved":
      return ["suspended"];
    case "rejected":
    case "suspended":
      return [];
  }
}

/**
 * shared profile avatar 未設定時の fallback initials を返す。
 */
export function buildCreatorReviewAvatarFallback(displayName: string): string {
  const trimmed = displayName.trim();
  if (trimmed === "") {
    return "CR";
  }

  return trimmed
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.at(0)?.toUpperCase() ?? "")
    .join("");
}

/**
 * admin review timestamp を UTC 固定表示へ整形する。
 */
export function formatCreatorReviewTimestamp(value: string | null): string {
  if (value === null) {
    return "未記録";
  }

  const date = new Date(value);
  if (Number.isNaN(date.valueOf())) {
    return value;
  }

  return [
    `${date.getUTCFullYear()}/${String(date.getUTCMonth() + 1).padStart(2, "0")}/${String(date.getUTCDate()).padStart(2, "0")}`,
    `${String(date.getUTCHours()).padStart(2, "0")}:${String(date.getUTCMinutes()).padStart(2, "0")} UTC`,
  ].join(" ");
}

/**
 * evidence file size を人が読みやすい単位へ整形する。
 */
export function formatCreatorReviewFileSize(fileSizeBytes: number): string {
  if (fileSizeBytes < 1024) {
    return `${fileSizeBytes} B`;
  }

  if (fileSizeBytes < 1024 * 1024) {
    return `${(fileSizeBytes / 1024).toFixed(1)} KB`;
  }

  return `${(fileSizeBytes / (1024 * 1024)).toFixed(1)} MB`;
}

/**
 * reason code から admin UI 表示用 metadata を返す。
 */
export function getCreatorReviewReasonOption(code: string | null): CreatorReviewReasonOption | null {
  if (code === null) {
    return null;
  }

  return creatorReviewReasonOptions.find((option) => option.code === code) ?? null;
}

/**
 * reject reason に応じた初期 reject handling を返す。
 */
export function getSuggestedCreatorReviewRejectHandling(
  reasonCode: string,
): CreatorReviewRejectHandlingMode {
  if (supportReviewReasonCodes.has(reasonCode)) {
    return "support_review_required";
  }

  return "resubmit_eligible";
}

/**
 * reject handling 選択肢から decision metadata を返す。
 */
export function getCreatorReviewRejectHandling(
  value: CreatorReviewRejectHandlingMode,
): CreatorReviewRejectHandling {
  return creatorReviewRejectHandlingOptions.find((option) => option.value === value) ?? creatorReviewRejectHandlingOptions[0];
}
