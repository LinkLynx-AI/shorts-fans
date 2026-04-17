import Link from "next/link";
import { notFound } from "next/navigation";
import { ApiError } from "@/shared/api";
import { assertAdminUiEnabled } from "../../_lib/admin-ui-access";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Button,
  SurfacePanel,
} from "@/shared/ui";
import {
  buildCreatorReviewAvatarFallback,
  formatCreatorReviewFileSize,
  formatCreatorReviewTimestamp,
  getCreatorReviewCase,
  getCreatorReviewReasonOption,
  getCreatorReviewStateLabel,
  isCreatorReviewUserId,
  normalizeCreatorReviewState,
} from "@/entities/creator-review";
import { CreatorReviewDecisionForm } from "@/features/creator-review-decision";

function getStateBadgeClass(state: string) {
  switch (state) {
    case "approved":
      return "border-[#cfe7d7] bg-[#f3fbf4] text-[#1f6a35]";
    case "rejected":
      return "border-[#f1d3d3] bg-[#fff5f5] text-[#9b2c2c]";
    case "suspended":
      return "border-[#f0deba] bg-[#fff8eb] text-[#8a5a00]";
    case "submitted":
    default:
      return "border-[#d7e6f5] bg-[#f5faff] text-[#1f628f]";
  }
}

function isNotFoundApiError(error: unknown): boolean {
  return error instanceof ApiError && error.status === 404;
}

export default async function AdminCreatorReviewCasePage({
  params,
  searchParams,
}: {
  params: Promise<{ userId: string }>;
  searchParams: Promise<{ state?: string | string[] }>;
}) {
  assertAdminUiEnabled();
  const [{ userId }, { state }] = await Promise.all([params, searchParams]);
  const activeState = normalizeCreatorReviewState(state);
  if (!isCreatorReviewUserId(userId)) {
    notFound();
  }

  let reviewCase;
  try {
    reviewCase = await getCreatorReviewCase({ userId });
  } catch (error) {
    if (isNotFoundApiError(error)) {
      notFound();
    }
    throw error;
  }

  const rejectionReason = getCreatorReviewReasonOption(reviewCase.rejection?.reasonCode ?? null);

  return (
    <main className="mx-auto flex min-h-full w-full max-w-6xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Button asChild variant="secondary">
          <Link href={`/admin/creator-reviews?state=${activeState}`}>一覧へ戻る</Link>
        </Button>
        <span
          className={[
            "rounded-full border px-3 py-1 text-[11px] font-bold uppercase tracking-[0.16em]",
            getStateBadgeClass(reviewCase.state),
          ].join(" ")}
        >
          {getCreatorReviewStateLabel(reviewCase.state)}
        </span>
      </div>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
        <SurfacePanel className="overflow-hidden border-none bg-[linear-gradient(155deg,#14283d_0%,#285b84_52%,#ddeef8_100%)] px-6 py-6 text-white shadow-[0_28px_56px_rgba(16,42,67,0.22)]">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
            <Avatar className="size-[84px] border border-white/28 bg-white/12 text-[24px] font-semibold text-white shadow-none">
              {reviewCase.sharedProfile.avatar ? (
                <AvatarImage
                  alt={`${reviewCase.sharedProfile.displayName} avatar`}
                  src={reviewCase.sharedProfile.avatar.url}
                />
              ) : null}
              <AvatarFallback className="bg-transparent text-inherit">
                {buildCreatorReviewAvatarFallback(reviewCase.sharedProfile.displayName)}
              </AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-white/72">creator applicant</p>
              <h1 className="mt-2 font-display text-[34px] font-semibold leading-[1.02] tracking-[-0.05em]">
                {reviewCase.sharedProfile.displayName}
              </h1>
              <p className="mt-2 text-sm text-white/80">{reviewCase.sharedProfile.handle}</p>
            </div>
          </div>
          <p className="mt-5 max-w-3xl text-sm leading-6 text-white/82">{reviewCase.creatorBio}</p>
        </SurfacePanel>

        <SurfacePanel className="px-5 py-5 text-foreground">
          <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">review timeline</p>
          <dl className="mt-4 grid gap-3">
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">submitted</dt>
              <dd className="mt-1 text-sm text-foreground">{formatCreatorReviewTimestamp(reviewCase.review.submittedAt)}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">approved</dt>
              <dd className="mt-1 text-sm text-foreground">{formatCreatorReviewTimestamp(reviewCase.review.approvedAt)}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">rejected</dt>
              <dd className="mt-1 text-sm text-foreground">{formatCreatorReviewTimestamp(reviewCase.review.rejectedAt)}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">suspended</dt>
              <dd className="mt-1 text-sm text-foreground">{formatCreatorReviewTimestamp(reviewCase.review.suspendedAt)}</dd>
            </div>
          </dl>
        </SurfacePanel>
      </section>

      <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <SurfacePanel className="px-5 py-5 text-foreground">
          <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">shared profile</p>
          <dl className="mt-4 grid gap-3 text-sm leading-6 text-muted">
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">display name</dt>
              <dd className="mt-1 text-foreground">{reviewCase.sharedProfile.displayName}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">handle</dt>
              <dd className="mt-1 text-foreground">{reviewCase.sharedProfile.handle}</dd>
            </div>
          </dl>

          <p className="mt-5 text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">intake</p>
          <dl className="mt-4 grid gap-3 text-sm leading-6 text-muted sm:grid-cols-2">
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">legal name</dt>
              <dd className="mt-1 text-foreground">{reviewCase.intake.legalName}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">birth date</dt>
              <dd className="mt-1 text-foreground">{reviewCase.intake.birthDate ?? "未入力"}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">payout type</dt>
              <dd className="mt-1 text-foreground">{reviewCase.intake.payoutRecipientType ?? "未入力"}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">payout name</dt>
              <dd className="mt-1 text-foreground">{reviewCase.intake.payoutRecipientName || "未入力"}</dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">consent</dt>
              <dd className="mt-1 text-foreground">
                {reviewCase.intake.acceptsConsentResponsibility ? "確認済み" : "未確認"}
              </dd>
            </div>
            <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
              <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">prohibited category</dt>
              <dd className="mt-1 text-foreground">
                {reviewCase.intake.declaresNoProhibitedCategory ? "非該当を宣言済み" : "未確認"}
              </dd>
            </div>
          </dl>
        </SurfacePanel>

        <div className="grid gap-6">
          <SurfacePanel className="px-5 py-5 text-foreground">
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">evidence</p>
            <div className="mt-4 grid gap-3">
              {reviewCase.evidences.map((evidence) => (
                <div
                  className="rounded-[20px] border border-border bg-[#f8fbfe] px-4 py-4"
                  key={evidence.kind}
                >
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div>
                      <p className="text-[15px] font-semibold tracking-[-0.02em] text-foreground">{evidence.fileName}</p>
                      <p className="mt-1 text-sm leading-6 text-muted">
                        {evidence.kind} / {evidence.mimeType} / {formatCreatorReviewFileSize(evidence.fileSizeBytes)}
                      </p>
                      <p className="mt-1 text-sm leading-6 text-muted">
                        uploaded: {formatCreatorReviewTimestamp(evidence.uploadedAt)}
                      </p>
                    </div>
                    <Button asChild variant="secondary">
                      <a href={evidence.accessUrl} rel="noreferrer" target="_blank">
                        Open evidence
                      </a>
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </SurfacePanel>

          {reviewCase.rejection ? (
            <SurfacePanel className="px-5 py-5 text-foreground">
              <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">rejection</p>
              <div className="mt-4 grid gap-3 text-sm leading-6 text-muted">
                <div className="rounded-[18px] border border-border bg-[#fff8f8] px-4 py-3">
                  <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-[#9b2c2c]">reason code</p>
                  <p className="mt-1 text-foreground">
                    {rejectionReason ? rejectionReason.label : (reviewCase.rejection.reasonCode ?? "未設定")}
                  </p>
                  {rejectionReason ? (
                    <p className="mt-1 text-sm leading-6 text-muted">{rejectionReason.description}</p>
                  ) : null}
                </div>
                <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
                  <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">resubmit</p>
                  <p className="mt-1 text-foreground">
                    {reviewCase.rejection.isResubmitEligible ? "許可" : "未許可"}
                  </p>
                </div>
                <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
                  <p className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">support review</p>
                  <p className="mt-1 text-foreground">
                    {reviewCase.rejection.isSupportReviewRequired ? "必要" : "不要"}
                  </p>
                </div>
              </div>
            </SurfacePanel>
          ) : null}

          <CreatorReviewDecisionForm
            key={`${reviewCase.userId}:${reviewCase.state}`}
            reviewCase={reviewCase}
          />
        </div>
      </section>
    </main>
  );
}
