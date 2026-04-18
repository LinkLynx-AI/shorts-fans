"use client";

import { startTransition, useState } from "react";
import { useRouter } from "next/navigation";

import { ApiError } from "@/shared/api";
import { Button, SurfacePanel } from "@/shared/ui";
import {
  applyCreatorReviewDecision,
  creatorReviewReasonOptions,
  creatorReviewRejectHandlingOptions,
  getCreatorReviewAvailableDecisions,
  getCreatorReviewDecisionLabel,
  getCreatorReviewRejectHandling,
  getSuggestedCreatorReviewRejectHandling,
  type CreatorReviewCase,
  type CreatorReviewDecision,
  type CreatorReviewRejectHandlingMode,
} from "@/entities/creator-review";

type CreatorReviewDecisionFormProps = {
  reviewCase: CreatorReviewCase;
};

function getDefaultDecision(reviewCase: CreatorReviewCase): CreatorReviewDecision {
  if (reviewCase.state === "approved") {
    return "suspended";
  }

  return "approved";
}

function getCreatorReviewDecisionErrorMessage(error: unknown): string {
  if (!(error instanceof ApiError)) {
    return "審査更新に失敗しました。時間を置いてから再度お試しください。";
  }

  if (error.code === "network") {
    return "通信に失敗しました。接続を確認してから再度お試しください。";
  }

  switch (error.status) {
    case 400:
      return "入力内容が不正です。却下理由と reject metadata を確認してください。";
    case 404:
      return "対象の申請が見つかりませんでした。";
    case 409:
      return "申請状態が更新されたため処理できませんでした。再読み込みしてください。";
    default:
      return "審査更新に失敗しました。時間を置いてから再度お試しください。";
  }
}

/**
 * admin creator review detail から decision mutation を実行する。
 */
export function CreatorReviewDecisionForm({
  reviewCase,
}: CreatorReviewDecisionFormProps) {
  const router = useRouter();
  const availableDecisions = getCreatorReviewAvailableDecisions(reviewCase.state);
  const singleDecision = availableDecisions[0];
  const defaultReasonCode = creatorReviewReasonOptions[0]?.code ?? "";
  const [decision, setDecision] = useState<CreatorReviewDecision>(getDefaultDecision(reviewCase));
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [reasonCode, setReasonCode] = useState<string>(defaultReasonCode);
  const [rejectHandlingMode, setRejectHandlingMode] = useState<CreatorReviewRejectHandlingMode>(
    getSuggestedCreatorReviewRejectHandling(defaultReasonCode),
  );
  const [isRejectHandlingCustomized, setIsRejectHandlingCustomized] = useState(false);

  if (availableDecisions.length === 0) {
    return (
      <SurfacePanel className="px-5 py-5 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">decision</p>
        <p className="mt-3 text-sm leading-6 text-muted">
          この状態では追加の admin decision はありません。
        </p>
      </SurfacePanel>
    );
  }

  const requiresReason = decision === "rejected";

  const submit = async () => {
    if (isSubmitting) {
      return;
    }
    if (requiresReason && reasonCode.trim() === "") {
      setErrorMessage("却下理由を選択してください。");
      return;
    }

    setErrorMessage(null);
    setIsSubmitting(true);

    const rejectHandling = getCreatorReviewRejectHandling(rejectHandlingMode);

    try {
      await applyCreatorReviewDecision({
        decision,
        isResubmitEligible: requiresReason ? rejectHandling.isResubmitEligible : false,
        isSupportReviewRequired: requiresReason ? rejectHandling.isSupportReviewRequired : false,
        reasonCode: requiresReason ? reasonCode : "",
        userId: reviewCase.userId,
      });

      startTransition(() => {
        router.refresh();
      });
    } catch (error) {
      setErrorMessage(getCreatorReviewDecisionErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <SurfacePanel className="px-5 py-5 text-foreground">
      <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">decision</p>
      <h2 className="mt-3 font-display text-[24px] font-semibold leading-[1.12] tracking-[-0.03em]">
        審査を更新する
      </h2>
      <p className="mt-2 text-sm leading-6 text-muted">
        submitted では承認または却下、approved では停止のみ実行できます。
      </p>

      {availableDecisions.length > 1 ? (
        <fieldset className="mt-5 grid gap-2">
          <legend className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-ink">
            Action
          </legend>
          <div className="grid gap-2 sm:grid-cols-2">
            {availableDecisions.map((option) => {
              const active = decision === option;

              return (
                <button
                  aria-pressed={active}
                  className={[
                    "min-h-12 rounded-[18px] border px-4 py-3 text-left text-sm font-semibold transition",
                    active
                      ? "border-[#7bb6e8] bg-[#eff8ff] text-[#195784] shadow-[0_10px_24px_rgba(80,159,224,0.12)]"
                      : "border-border bg-white text-foreground hover:border-[#cfe2f3]",
                  ].join(" ")}
                  key={option}
                  onClick={() => {
                    setDecision(option);
                    setErrorMessage(null);
                  }}
                  type="button"
                >
                  {getCreatorReviewDecisionLabel(option)}
                </button>
              );
            })}
          </div>
        </fieldset>
      ) : (
        <div className="mt-5 rounded-[20px] border border-border bg-[#f8fbfe] px-4 py-4 text-sm leading-6 text-muted">
          現在の状態では
          <span className="font-semibold text-foreground">
            {" "}
            {singleDecision ? getCreatorReviewDecisionLabel(singleDecision) : "更新"}{" "}
          </span>
          のみ実行できます。
        </div>
      )}

      {requiresReason ? (
        <>
          <label className="mt-5 grid gap-2" htmlFor="creator-review-reason-code">
            <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-ink">
              Reject reason
            </span>
            <select
              className="min-h-12 rounded-[18px] border border-border bg-white px-4 text-sm text-foreground outline-none transition focus:border-accent focus:ring-4 focus:ring-ring/60"
              disabled={isSubmitting}
              id="creator-review-reason-code"
              onChange={(event) => {
                const nextReasonCode = event.target.value;
                setReasonCode(nextReasonCode);
                if (!isRejectHandlingCustomized) {
                  setRejectHandlingMode(getSuggestedCreatorReviewRejectHandling(nextReasonCode));
                }
                setErrorMessage(null);
              }}
              value={reasonCode}
            >
              {creatorReviewReasonOptions.map((option) => (
                <option key={option.code} value={option.code}>
                  {option.label}
                </option>
              ))}
            </select>
            <p className="text-sm leading-6 text-muted">
              {creatorReviewReasonOptions.find((option) => option.code === reasonCode)?.description}
            </p>
          </label>

          <fieldset className="mt-5 grid gap-2">
            <legend className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-ink">
              Reject handling
            </legend>
            <div className="grid gap-2">
              {creatorReviewRejectHandlingOptions.map((option) => {
                const active = rejectHandlingMode === option.value;

                return (
                  <button
                    aria-pressed={active}
                    className={[
                      "rounded-[18px] border px-4 py-4 text-left transition",
                      active
                        ? "border-[#7bb6e8] bg-[#eff8ff] text-[#195784] shadow-[0_10px_24px_rgba(80,159,224,0.12)]"
                        : "border-border bg-white text-foreground hover:border-[#cfe2f3]",
                    ].join(" ")}
                    key={option.value}
                    onClick={() => {
                      setRejectHandlingMode(option.value);
                      setIsRejectHandlingCustomized(true);
                      setErrorMessage(null);
                    }}
                    type="button"
                  >
                    <p className="text-sm font-semibold">{option.label}</p>
                    <p className={active ? "mt-1 text-sm leading-6 text-[#3b6f95]" : "mt-1 text-sm leading-6 text-muted"}>
                      {option.description}
                    </p>
                  </button>
                );
              })}
            </div>
            <p className="text-sm leading-6 text-muted">
              reject metadata は単一選択にして、resubmit と support review の矛盾入力を防ぎます。
            </p>
          </fieldset>
        </>
      ) : null}

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
          {isSubmitting ? "更新中..." : getCreatorReviewDecisionLabel(decision)}
        </Button>
      </div>
    </SurfacePanel>
  );
}
