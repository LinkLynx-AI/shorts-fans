import Link from "next/link";

import { assertAdminUiEnabled } from "../_lib/admin-ui-access";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  SurfacePanel,
} from "@/shared/ui";
import {
  buildSubmissionReviewAvatarFallback,
  formatSubmissionReviewTimestamp,
  getSubmissionReviewQueue,
  getSubmissionReviewSubmitKindLabel,
} from "@/entities/submission-review";

export default async function AdminSubmissionReviewsPage() {
  assertAdminUiEnabled();
  const queue = await getSubmissionReviewQueue();

  return (
    <main className="mx-auto flex min-h-full w-full max-w-6xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)]">
        <SurfacePanel className="overflow-hidden border-none bg-[linear-gradient(160deg,#352019_0%,#8a4b2d_46%,#f4dfcf_100%)] px-6 py-6 text-white shadow-[0_28px_56px_rgba(74,35,18,0.18)]">
          <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-white/72">admin submission review</p>
          <h1 className="mt-3 max-w-xl font-display text-[34px] font-semibold leading-[1.02] tracking-[-0.05em]">
            pending review の package を object-level で裁く
          </h1>
          <p className="mt-3 max-w-2xl text-sm leading-6 text-white/82">
            canonical main と linked short を同時に確認し、main / short ごとに approved、
            revision requested、rejected を反映します。
          </p>
        </SurfacePanel>

        <SurfacePanel className="bg-[linear-gradient(180deg,#ffffff,#fff8f1)] px-5 py-5 text-foreground">
          <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">flow</p>
          <ol className="mt-4 grid gap-3 text-sm leading-6 text-muted">
            <li className="rounded-[18px] border border-border bg-white px-4 py-3">
              1. pending intake を開いて creator / package 文脈を確認
            </li>
            <li className="rounded-[18px] border border-border bg-white px-4 py-3">
              2. main と pending short ごとに decision を選択
            </li>
            <li className="rounded-[18px] border border-border bg-white px-4 py-3">
              3. 必要なら reason code / review note を残して反映
            </li>
          </ol>
        </SurfacePanel>
      </section>

      <section className="grid gap-4">
        {queue.items.length === 0 ? (
          <SurfacePanel className="px-6 py-10 text-center text-foreground">
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">pending review</p>
            <h2 className="mt-3 font-display text-[28px] font-semibold tracking-[-0.04em]">
              対象の intake はありません
            </h2>
            <p className="mt-3 text-sm leading-6 text-muted">
              creator workspace から submission package を submit すると queue に表示されます。
            </p>
          </SurfacePanel>
        ) : (
          queue.items.map((item) => (
            <Link
              className="block"
              href={`/admin/submission-reviews/${item.intakeId}`}
              key={item.intakeId}
            >
              <SurfacePanel className="group px-5 py-5 text-foreground transition hover:-translate-y-0.5 hover:shadow-[0_24px_48px_rgba(74,35,18,0.08)]">
                <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                  <div className="flex min-w-0 items-start gap-4">
                    <Avatar className="size-[68px] border border-border bg-[#fff4ea] text-[17px] font-semibold text-[#8a4b2d] shadow-none">
                      {item.creator.avatar ? (
                        <AvatarImage
                          alt={`${item.creator.displayName} avatar`}
                          src={item.creator.avatar.url}
                        />
                      ) : null}
                      <AvatarFallback className="bg-transparent text-inherit">
                        {buildSubmissionReviewAvatarFallback(item.creator.displayName)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <p className="text-[18px] font-semibold tracking-[-0.03em] text-foreground">
                          {item.creator.displayName}
                        </p>
                        <span className="rounded-full border border-[#f0deba] bg-[#fff8eb] px-3 py-1 text-[11px] font-bold uppercase tracking-[0.16em] text-[#8a5a00]">
                          {getSubmissionReviewSubmitKindLabel(item.submitKind)}
                        </span>
                        {item.mainDecisionRequired ? (
                          <span className="rounded-full border border-[#d7e6f5] bg-[#f5faff] px-3 py-1 text-[11px] font-bold uppercase tracking-[0.16em] text-[#1f628f]">
                            main pending
                          </span>
                        ) : null}
                      </div>
                      <p className="mt-1 text-sm text-muted">{item.creator.handle}</p>
                      <p className="mt-3 line-clamp-2 text-sm leading-6 text-muted">{item.creator.bio}</p>
                    </div>
                  </div>

                  <dl className="grid gap-3 text-sm leading-6 text-muted sm:grid-cols-2 lg:min-w-[360px]">
                    <div className="rounded-[18px] border border-border bg-[#fffaf4] px-4 py-3">
                      <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">submitted</dt>
                      <dd className="mt-1 text-foreground">{formatSubmissionReviewTimestamp(item.submittedAt)}</dd>
                    </div>
                    <div className="rounded-[18px] border border-border bg-[#fffaf4] px-4 py-3">
                      <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">pending shorts</dt>
                      <dd className="mt-1 text-foreground">{item.pendingShortCount} / {item.shortCount}</dd>
                    </div>
                  </dl>
                </div>
              </SurfacePanel>
            </Link>
          ))
        )}
      </section>
    </main>
  );
}
