"use client";

import { startTransition, useState } from "react";
import { useRouter } from "next/navigation";

import { ApiError } from "@/shared/api";
import { Button, SurfacePanel } from "@/shared/ui";
import {
  doesSubmissionReviewDecisionRequireReason,
  getSubmissionReviewDecisionLabel,
  submissionReviewReasonOptions,
  type SubmissionReviewCase,
  type SubmissionReviewDecision,
} from "@/entities/submission-review";

type SubmissionReviewDecisionFormProps = {
  onSubmitDecision: (input: SubmitSubmissionReviewDecisionInput) => Promise<void>;
  reviewCase: SubmissionReviewCase;
};

type SubmitSubmissionReviewDecisionInput = {
  intakeId: string;
  mainDecision?: {
    decision: SubmissionReviewDecision;
    reasonCode?: string;
    reviewNote?: string;
  };
  shortDecisions: Array<{
    decision: SubmissionReviewDecision;
    reasonCode?: string;
    reviewNote?: string;
    shortId: string;
  }>;
};

type DecisionDraft = {
  decision: SubmissionReviewDecision | "";
  reasonCode: string;
  reviewNote: string;
};

function createEmptyDecisionDraft(): DecisionDraft {
  return {
    decision: "",
    reasonCode: "",
    reviewNote: "",
  };
}

function getSubmissionReviewDecisionErrorMessage(error: unknown): string {
  if (!(error instanceof ApiError)) {
    return "審査更新に失敗しました。時間を置いてから再度お試しください。";
  }

  if (error.code === "network") {
    return "通信に失敗しました。接続を確認してから再度お試しください。";
  }

  switch (error.status) {
    case 400:
      return "入力内容が不正です。decision と reason code を確認してください。";
    case 404:
      return "対象の intake が見つかりませんでした。";
    case 409:
      return "review 対象が更新されたため反映できませんでした。再読み込みしてください。";
    default:
      return "審査更新に失敗しました。時間を置いてから再度お試しください。";
  }
}

function validateDecisionDraft(label: string, draft: DecisionDraft): string | null {
  if (draft.decision === "") {
    return `${label} の decision を選択してください。`;
  }
  if (doesSubmissionReviewDecisionRequireReason(draft.decision) && draft.reasonCode.trim() === "") {
    return `${label} の reason code を選択してください。`;
  }

  return null;
}

function DecisionPicker({
  draft,
  disabled,
  label,
  targetKey,
  onChange,
}: {
  draft: DecisionDraft;
  disabled: boolean;
  label: string;
  targetKey: string;
  onChange: (nextDraft: DecisionDraft) => void;
}) {
  const requiresReason = doesSubmissionReviewDecisionRequireReason(draft.decision);
  const reasonFieldId = `${targetKey}-reason`;
  const reviewNoteFieldId = `${targetKey}-review-note`;

  return (
    <div className="rounded-[22px] border border-border bg-[#fbfdff] px-4 py-4">
      <div className="flex flex-col gap-1">
        <p className="text-sm font-semibold text-foreground">{label}</p>
        <p className="text-xs uppercase tracking-[0.16em] text-accent-ink">decision required</p>
      </div>

      <div className="mt-4 grid gap-2 sm:grid-cols-3">
        {(["approved", "revision_requested", "rejected"] as const).map((decision) => {
          const active = draft.decision === decision;

          return (
            <button
              aria-pressed={active}
              className={[
                "min-h-12 rounded-[18px] border px-4 py-3 text-left text-sm font-semibold transition",
                active
                  ? "border-[#7bb6e8] bg-[#eff8ff] text-[#195784] shadow-[0_10px_24px_rgba(80,159,224,0.12)]"
                  : "border-border bg-white text-foreground hover:border-[#cfe2f3]",
              ].join(" ")}
              disabled={disabled}
              key={decision}
              onClick={() => {
                onChange({
                  ...draft,
                  decision,
                });
              }}
              type="button"
            >
              {getSubmissionReviewDecisionLabel(decision)}
            </button>
          );
        })}
      </div>

      {requiresReason ? (
        <label className="mt-4 grid gap-2" htmlFor={reasonFieldId}>
          <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-ink">
            Reason code
          </span>
          <select
            aria-label={reasonFieldId}
            className="min-h-12 rounded-[18px] border border-border bg-white px-4 text-sm text-foreground outline-none transition focus:border-accent focus:ring-4 focus:ring-ring/60"
            disabled={disabled}
            id={reasonFieldId}
            onChange={(event) => {
              onChange({
                ...draft,
                reasonCode: event.target.value,
              });
            }}
            value={draft.reasonCode}
          >
            <option value="">選択してください</option>
            {submissionReviewReasonOptions.map((option) => (
              <option key={option.code} value={option.code}>
                {option.label}
              </option>
            ))}
          </select>
          {draft.reasonCode !== "" ? (
            <p className="text-sm leading-6 text-muted">
              {submissionReviewReasonOptions.find((option) => option.code === draft.reasonCode)?.description}
            </p>
          ) : (
            <p className="text-sm leading-6 text-muted">reason code を明示的に選択してください。</p>
          )}
        </label>
      ) : null}

      <label className="mt-4 grid gap-2" htmlFor={reviewNoteFieldId}>
        <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-ink">
          Review note
        </span>
        <textarea
          aria-label={reviewNoteFieldId}
          className="min-h-28 rounded-[18px] border border-border bg-white px-4 py-3 text-sm leading-6 text-foreground outline-none transition focus:border-accent focus:ring-4 focus:ring-ring/60"
          disabled={disabled}
          id={reviewNoteFieldId}
          onChange={(event) => {
            onChange({
              ...draft,
              reviewNote: event.target.value,
            });
          }}
          placeholder="任意メモを残す場合のみ入力"
          value={draft.reviewNote}
        />
      </label>
    </div>
  );
}

/**
 * admin submission review detail から object-level decision を実行する。
 */
export function SubmissionReviewDecisionForm({
  onSubmitDecision,
  reviewCase,
}: SubmissionReviewDecisionFormProps) {
  const router = useRouter();
  const pendingShorts = reviewCase.shorts.filter((item) => item.decisionRequired);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [mainDraft, setMainDraft] = useState<DecisionDraft>(createEmptyDecisionDraft);
  const [shortDrafts, setShortDrafts] = useState<Record<string, DecisionDraft>>(() => {
    return Object.fromEntries(
      pendingShorts.map((item) => [item.id, createEmptyDecisionDraft()]),
    );
  });

  if (reviewCase.intake.status !== "pending_review" || (!reviewCase.main.decisionRequired && pendingShorts.length === 0)) {
    return (
      <SurfacePanel className="px-5 py-5 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">decision</p>
        <p className="mt-3 text-sm leading-6 text-muted">
          この intake では追加の admin decision はありません。
        </p>
      </SurfacePanel>
    );
  }

  const submit = async () => {
    if (isSubmitting) {
      return;
    }

    if (reviewCase.main.decisionRequired) {
      const validationMessage = validateDecisionDraft("本編", mainDraft);
      if (validationMessage) {
        setErrorMessage(validationMessage);
        return;
      }
    }

    for (const [index, item] of pendingShorts.entries()) {
      const shortDraft = shortDrafts[item.id] ?? createEmptyDecisionDraft();
      const validationMessage = validateDecisionDraft(`short ${index + 1}`, shortDraft);
      if (validationMessage) {
        setErrorMessage(validationMessage);
        return;
      }
    }

    setErrorMessage(null);
    setIsSubmitting(true);

    try {
      const request = {
        intakeId: reviewCase.intake.id,
        ...(reviewCase.main.decisionRequired
          ? {
              mainDecision: {
                decision: mainDraft.decision as SubmissionReviewDecision,
                reasonCode: doesSubmissionReviewDecisionRequireReason(mainDraft.decision)
                  ? mainDraft.reasonCode
                  : "",
                reviewNote: mainDraft.reviewNote,
              },
            }
          : {}),
        shortDecisions: pendingShorts.map((item) => {
          const draft = shortDrafts[item.id] ?? createEmptyDecisionDraft();

          return {
            decision: draft.decision as SubmissionReviewDecision,
            reasonCode: doesSubmissionReviewDecisionRequireReason(draft.decision)
              ? draft.reasonCode
              : "",
            reviewNote: draft.reviewNote,
            shortId: item.id,
          };
        }),
      };

      await onSubmitDecision(request);

      startTransition(() => {
        router.replace("/admin/submission-reviews");
      });
    } catch (error) {
      setErrorMessage(getSubmissionReviewDecisionErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <SurfacePanel className="px-5 py-5 text-foreground">
      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">decision</p>
      <h2 className="mt-3 font-display text-[24px] font-semibold leading-[1.12] tracking-[-0.03em]">
        object-level decision を反映する
      </h2>
      <p className="mt-2 text-sm leading-6 text-muted">
        pending target ごとに decision を明示し、revision requested / rejected では reason code を必須にします。
      </p>

      <div className="mt-5 grid gap-4">
        {reviewCase.main.decisionRequired ? (
          <DecisionPicker
            disabled={isSubmitting}
            draft={mainDraft}
            label="main"
            targetKey="main"
            onChange={(nextDraft) => {
              setMainDraft(nextDraft);
              setErrorMessage(null);
            }}
          />
        ) : null}

        {pendingShorts.map((item, index) => (
          <DecisionPicker
            disabled={isSubmitting}
            draft={shortDrafts[item.id] ?? createEmptyDecisionDraft()}
            key={item.id}
            label={`short ${index + 1}`}
            targetKey={`short-${index + 1}`}
            onChange={(nextDraft) => {
              setShortDrafts((current) => ({
                ...current,
                [item.id]: nextDraft,
              }));
              setErrorMessage(null);
            }}
          />
        ))}
      </div>

      {errorMessage ? (
        <div
          className="mt-5 rounded-[18px] border border-[rgba(255,184,189,0.84)] bg-[linear-gradient(180deg,rgba(255,247,248,0.98),rgba(255,241,243,0.96))] px-4 py-4 text-sm leading-6 text-foreground"
          role="alert"
        >
          {errorMessage}
        </div>
      ) : null}

      <div className="mt-5 flex justify-end">
        <Button disabled={isSubmitting} onClick={() => void submit()} type="button">
          {isSubmitting ? "更新中..." : "decision を反映する"}
        </Button>
      </div>
    </SurfacePanel>
  );
}
