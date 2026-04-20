import type {
  SubmissionReviewDecision,
  SubmissionReviewDecisionTargetState,
  SubmissionReviewIntakeStatus,
  SubmissionReviewObjectState,
  SubmissionReviewSubmitKind,
} from "../api/contracts";

export type SubmissionReviewReasonOption = {
  code: string;
  description: string;
  label: string;
};

export const submissionReviewReasonOptions = [
  {
    code: "content_safety_issue",
    description: "コンテンツ安全性の観点で追加対応が必要です。",
    label: "安全性の懸念",
  },
  {
    code: "consent_or_ownership_issue",
    description: "権利確認または同意確認が不足しています。",
    label: "同意・権利確認不足",
  },
  {
    code: "prohibited_category",
    description: "取扱禁止カテゴリに該当、または該当の懸念があります。",
    label: "禁止カテゴリ",
  },
  {
    code: "continuity_mismatch",
    description: "submission package の continuity が崩れており再提出が必要です。",
    label: "連続性不一致",
  },
  {
    code: "metadata_incomplete",
    description: "caption / price / confirmation など review に必要な metadata が不足しています。",
    label: "metadata 不足",
  },
  {
    code: "quality_issue",
    description: "画質や内容品質の観点で修正が必要です。",
    label: "品質上の問題",
  },
] as const satisfies readonly SubmissionReviewReasonOption[];

const submissionReviewDecisionLabels: Record<SubmissionReviewDecision, string> = {
  approved: "承認する",
  rejected: "却下する",
  revision_requested: "修正依頼にする",
};

const submissionReviewStateLabels: Record<
  SubmissionReviewDecisionTargetState | SubmissionReviewIntakeStatus | SubmissionReviewObjectState,
  string
> = {
  approved_for_publish: "公開承認済み",
  approved_for_unlock: "解放承認済み",
  decision_applied: "反映済み",
  draft: "下書き",
  pending_review: "審査待ち",
  rejected: "却下",
  revision_requested: "修正依頼",
};

const submissionReviewSubmitKindLabels: Record<SubmissionReviewSubmitKind, string> = {
  initial_submit: "初回 submit",
  resubmit: "再 submit",
};

/**
 * review state の表示ラベルを返す。
 */
export function getSubmissionReviewStateLabel(
  state: SubmissionReviewDecisionTargetState | SubmissionReviewIntakeStatus | SubmissionReviewObjectState,
): string {
  return submissionReviewStateLabels[state];
}

/**
 * decision action の表示ラベルを返す。
 */
export function getSubmissionReviewDecisionLabel(decision: SubmissionReviewDecision): string {
  return submissionReviewDecisionLabels[decision];
}

/**
 * submit kind の表示ラベルを返す。
 */
export function getSubmissionReviewSubmitKindLabel(kind: SubmissionReviewSubmitKind): string {
  return submissionReviewSubmitKindLabels[kind];
}

/**
 * decision に reason code が必要かを返す。
 */
export function doesSubmissionReviewDecisionRequireReason(decision: SubmissionReviewDecision | ""): boolean {
  return decision === "revision_requested" || decision === "rejected";
}

/**
 * reason code から option metadata を返す。
 */
export function getSubmissionReviewReasonOption(code: string | null): SubmissionReviewReasonOption | null {
  if (code === null) {
    return null;
  }

  return submissionReviewReasonOptions.find((option) => option.code === code) ?? null;
}

/**
 * avatar 未設定時の fallback initials を返す。
 */
export function buildSubmissionReviewAvatarFallback(displayName: string): string {
  const trimmed = displayName.trim();
  if (trimmed === "") {
    return "SR";
  }

  return trimmed
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.at(0)?.toUpperCase() ?? "")
    .join("");
}

/**
 * UTC timestamp を admin review 表示向けに整形する。
 */
export function formatSubmissionReviewTimestamp(value: string | null): string {
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
 * decision source を reviewer 向け表示に正規化する。
 */
export function formatSubmissionReviewDecisionSource(value: string | null): string {
  if (value === null) {
    return "未記録";
  }
  if (value === "manual_override") {
    return "manual";
  }

  return value;
}

/**
 * 金額を JPY 固定で整形する。
 */
export function formatSubmissionReviewPrice(priceJpy: number): string {
  return `${priceJpy.toLocaleString("ja-JP")} 円`;
}
