import Link from "next/link";
import { notFound } from "next/navigation";

import { ApiError } from "@/shared/api";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Button,
  SurfacePanel,
} from "@/shared/ui";
import { SubmissionReviewDecisionForm } from "@/features/submission-review-decision";
import {
  buildSubmissionReviewAvatarFallback,
  formatSubmissionReviewDecisionSource,
  formatSubmissionReviewPrice,
  formatSubmissionReviewTimestamp,
  getSubmissionReviewCase,
  getSubmissionReviewReasonOption,
  getSubmissionReviewStateLabel,
  getSubmissionReviewSubmitKindLabel,
  isSubmissionReviewIntakeId,
  type SubmissionReviewDecisionLog,
  type SubmissionReviewObjectState,
} from "@/entities/submission-review";
import { assertAdminUiAccess } from "../../_lib/admin-ui-access";
import { createAdminAPIFetcher } from "../../_lib/admin-api";
import { AdminReviewNavigation } from "../../_ui/admin-review-navigation";
import { applySubmissionReviewDecisionFromAdmin } from "./actions";

function getStateBadgeClass(state: SubmissionReviewObjectState | "decision_applied" | "pending_review") {
  switch (state) {
    case "approved_for_publish":
    case "approved_for_unlock":
      return "border-[#cfe7d7] bg-[#f3fbf4] text-[#1f6a35]";
    case "revision_requested":
      return "border-[#d7e6f5] bg-[#f5faff] text-[#1f628f]";
    case "rejected":
      return "border-[#f1d3d3] bg-[#fff5f5] text-[#9b2c2c]";
    case "decision_applied":
      return "border-[#f0deba] bg-[#fff8eb] text-[#8a5a00]";
    case "pending_review":
    default:
      return "border-[#d7e6f5] bg-[#f5faff] text-[#1f628f]";
  }
}

function isNotFoundApiError(error: unknown): boolean {
  return error instanceof ApiError && error.status === 404;
}

function ReviewMetadata({
  currentDecisionSource,
  currentDecisionedAt,
  decisionLog,
  reasonCode,
  reviewNote,
  title,
}: {
  currentDecisionSource: string | null;
  currentDecisionedAt: string | null;
  decisionLog: SubmissionReviewDecisionLog | null;
  reasonCode: string | null;
  reviewNote: string | null;
  title: string;
}) {
  const reason = getSubmissionReviewReasonOption(reasonCode);
  const intakeReason = getSubmissionReviewReasonOption(decisionLog?.reasonCode ?? null);

  return (
    <div className="grid gap-3">
      <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
        <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">{title}</p>
        <dl className="mt-3 grid gap-2 text-sm leading-6 text-muted">
          <div>
            <dt className="font-semibold text-foreground">current source</dt>
            <dd>{formatSubmissionReviewDecisionSource(currentDecisionSource)}</dd>
          </div>
          <div>
            <dt className="font-semibold text-foreground">current decisioned</dt>
            <dd>{formatSubmissionReviewTimestamp(currentDecisionedAt)}</dd>
          </div>
          <div>
            <dt className="font-semibold text-foreground">current reason</dt>
            <dd>{reason ? reason.label : (reasonCode ?? "未記録")}</dd>
          </div>
          {reason ? <p className="text-sm leading-6 text-muted">{reason.description}</p> : null}
          <div>
            <dt className="font-semibold text-foreground">current note</dt>
            <dd>{reviewNote ?? "未記録"}</dd>
          </div>
        </dl>
      </div>

      <div className="rounded-[18px] border border-border bg-[#fffaf4] px-4 py-3">
        <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">intake decision log</p>
        {decisionLog ? (
          <dl className="mt-3 grid gap-2 text-sm leading-6 text-muted">
            <div>
              <dt className="font-semibold text-foreground">target state</dt>
              <dd>{getSubmissionReviewStateLabel(decisionLog.targetState)}</dd>
            </div>
            <div>
              <dt className="font-semibold text-foreground">decision source</dt>
              <dd>{formatSubmissionReviewDecisionSource(decisionLog.decisionSource)}</dd>
            </div>
            <div>
              <dt className="font-semibold text-foreground">decisioned</dt>
              <dd>{formatSubmissionReviewTimestamp(decisionLog.decisionedAt)}</dd>
            </div>
            <div>
              <dt className="font-semibold text-foreground">reason</dt>
              <dd>{intakeReason ? intakeReason.label : (decisionLog.reasonCode ?? "未記録")}</dd>
            </div>
            {intakeReason ? <p className="text-sm leading-6 text-muted">{intakeReason.description}</p> : null}
            <div>
              <dt className="font-semibold text-foreground">note</dt>
              <dd>{decisionLog.reviewNote ?? "未記録"}</dd>
            </div>
          </dl>
        ) : (
          <p className="mt-3 text-sm leading-6 text-muted">この intake ではまだ decision log がありません。</p>
        )}
      </div>
    </div>
  );
}

export default async function AdminSubmissionReviewCasePage({
  params,
}: {
  params: Promise<{ intakeId: string }>;
}) {
  await assertAdminUiAccess();
  const { intakeId } = await params;
  if (!isSubmissionReviewIntakeId(intakeId)) {
    notFound();
  }

  let reviewCase;
  try {
    reviewCase = await getSubmissionReviewCase({
      fetcher: createAdminAPIFetcher(),
      intakeId,
    });
  } catch (error) {
    if (isNotFoundApiError(error)) {
      notFound();
    }
    throw error;
  }

  return (
    <main className="mx-auto flex min-h-full w-full max-w-6xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <AdminReviewNavigation active="submission-reviews" />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <Button asChild variant="secondary">
          <Link href="/admin/submission-reviews" prefetch={false}>
            一覧へ戻る
          </Link>
        </Button>
        <span
          className={[
            "rounded-full border px-3 py-1 text-[11px] font-bold uppercase tracking-[0.16em]",
            getStateBadgeClass(reviewCase.intake.status),
          ].join(" ")}
        >
          {getSubmissionReviewStateLabel(reviewCase.intake.status)}
        </span>
      </div>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
        <SurfacePanel className="overflow-hidden border-none bg-[linear-gradient(155deg,#271812_0%,#8a4b2d_52%,#f3e5d8_100%)] px-6 py-6 text-white shadow-[0_28px_56px_rgba(74,35,18,0.2)]">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
            <Avatar className="size-[84px] border border-white/28 bg-white/12 text-[24px] font-semibold text-white shadow-none">
              {reviewCase.creator.avatar ? (
                <AvatarImage
                  alt={`${reviewCase.creator.displayName} avatar`}
                  src={reviewCase.creator.avatar.url}
                />
              ) : null}
              <AvatarFallback className="bg-transparent text-inherit">
                {buildSubmissionReviewAvatarFallback(reviewCase.creator.displayName)}
              </AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-white/72">submission package</p>
              <h1 className="mt-2 font-display text-[34px] font-semibold leading-[1.02] tracking-[-0.05em]">
                {reviewCase.creator.displayName}
              </h1>
              <p className="mt-2 text-sm text-white/80">{reviewCase.creator.handle}</p>
            </div>
          </div>
          <p className="mt-5 max-w-3xl text-sm leading-6 text-white/82">{reviewCase.creator.bio}</p>
        </SurfacePanel>

        <SurfacePanel className="px-5 py-5 text-foreground">
          <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">intake</p>
          <dl className="mt-4 grid gap-3 text-sm leading-6 text-muted">
            <div className="rounded-[18px] border border-border bg-[#fffaf4] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">submit kind</dt>
              <dd className="mt-1 text-foreground">{getSubmissionReviewSubmitKindLabel(reviewCase.intake.submitKind)}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#fffaf4] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">submitted</dt>
              <dd className="mt-1 text-foreground">{formatSubmissionReviewTimestamp(reviewCase.intake.submittedAt)}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#fffaf4] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">price</dt>
              <dd className="mt-1 text-foreground">{formatSubmissionReviewPrice(reviewCase.intake.mainPriceJpy)}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#fffaf4] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">ownership / consent</dt>
              <dd className="mt-1 text-foreground">
                {reviewCase.intake.ownershipConfirmed ? "ownership ok" : "ownership ng"} /{" "}
                {reviewCase.intake.consentConfirmed ? "consent ok" : "consent ng"}
              </dd>
            </div>
          </dl>
        </SurfacePanel>
      </section>

      <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <div className="grid gap-6">
          <SurfacePanel className="px-5 py-5 text-foreground">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">main</p>
                <h2 className="mt-2 font-display text-[24px] font-semibold leading-[1.12] tracking-[-0.03em]">
                  canonical main
                </h2>
              </div>
              <span
                className={[
                  "rounded-full border px-3 py-1 text-[11px] font-bold uppercase tracking-[0.16em]",
                  getStateBadgeClass(reviewCase.main.state),
                ].join(" ")}
              >
                {getSubmissionReviewStateLabel(reviewCase.main.state)}
              </span>
            </div>

            <div className="mt-4 overflow-hidden rounded-[24px] bg-black">
              <video
                className="block aspect-[4/5] w-full object-cover"
                controls
                playsInline
                poster={reviewCase.main.media.posterUrl}
                preload="metadata"
                src={reviewCase.main.media.url}
              />
            </div>

            <dl className="mt-4 grid gap-3 text-sm leading-6 text-muted sm:grid-cols-2">
              <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
                <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">price</dt>
                <dd className="mt-1 text-foreground">{formatSubmissionReviewPrice(reviewCase.main.priceJpy)}</dd>
              </div>
              <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
                <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">duration</dt>
                <dd className="mt-1 text-foreground">{reviewCase.main.media.durationSeconds} sec</dd>
              </div>
            </dl>
          </SurfacePanel>

          <ReviewMetadata
            currentDecisionSource={reviewCase.main.review.decisionSource}
            currentDecisionedAt={reviewCase.main.review.decisionedAt}
            decisionLog={reviewCase.main.intakeDecisionLog}
            reasonCode={reviewCase.main.review.reasonCode}
            reviewNote={reviewCase.main.review.reviewNote}
            title="main provenance"
          />
        </div>

        <SubmissionReviewDecisionForm
          key={`${reviewCase.intake.id}:${reviewCase.intake.status}`}
          onSubmitDecision={applySubmissionReviewDecisionFromAdmin}
          reviewCase={reviewCase}
        />
      </section>

      <section className="grid gap-4">
        <div className="flex items-center justify-between gap-3">
          <div>
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">linked shorts</p>
            <h2 className="mt-2 font-display text-[24px] font-semibold leading-[1.12] tracking-[-0.03em]">
              intake に含まれる short
            </h2>
          </div>
        </div>

        <div className="grid gap-4 lg:grid-cols-2">
          {reviewCase.shorts.map((item, index) => (
            <SurfacePanel className="px-5 py-5 text-foreground" key={item.id}>
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">short {index + 1}</p>
                  <p className="mt-2 text-sm leading-6 text-muted">{item.caption ?? "caption 未設定"}</p>
                </div>
                <span
                  className={[
                    "rounded-full border px-3 py-1 text-[11px] font-bold uppercase tracking-[0.16em]",
                    getStateBadgeClass(item.state),
                  ].join(" ")}
                >
                  {getSubmissionReviewStateLabel(item.state)}
                </span>
              </div>

              <div className="mt-4 overflow-hidden rounded-[24px] bg-black">
                <video
                  className="block aspect-[4/5] w-full object-cover"
                  controls
                  playsInline
                  poster={item.media.posterUrl}
                  preload="metadata"
                  src={item.media.url}
                />
              </div>

              <dl className="mt-4 grid gap-3 text-sm leading-6 text-muted">
                <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
                  <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">duration</dt>
                  <dd className="mt-1 text-foreground">{item.media.durationSeconds} sec</dd>
                </div>
              </dl>

              <div className="mt-4">
                <ReviewMetadata
                  currentDecisionSource={item.review.decisionSource}
                  currentDecisionedAt={item.review.decisionedAt}
                  decisionLog={item.intakeDecisionLog}
                  reasonCode={item.review.reasonCode}
                  reviewNote={item.review.reviewNote}
                  title="short provenance"
                />
              </div>
            </SurfacePanel>
          ))}
        </div>
      </section>
    </main>
  );
}
