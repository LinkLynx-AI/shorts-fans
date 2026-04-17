import Link from "next/link";

import { assertAdminUiEnabled } from "../_lib/admin-ui-access";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  SegmentedControl,
  SurfacePanel,
} from "@/shared/ui";
import {
  buildCreatorReviewAvatarFallback,
  formatCreatorReviewTimestamp,
  getCreatorReviewQueue,
  getCreatorReviewStateLabel,
  normalizeCreatorReviewState,
} from "@/entities/creator-review";

const queueTabs = [
  { key: "submitted", label: "審査待ち" },
  { key: "approved", label: "承認済み" },
  { key: "rejected", label: "却下済み" },
  { key: "suspended", label: "停止中" },
] as const;

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

export default async function AdminCreatorReviewsPage({
  searchParams,
}: {
  searchParams: Promise<{ state?: string | string[] }>;
}) {
  assertAdminUiEnabled();
  const { state } = await searchParams;
  const activeState = normalizeCreatorReviewState(state);
  const queue = await getCreatorReviewQueue({ state: activeState });

  return (
    <main className="mx-auto flex min-h-full w-full max-w-6xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)]">
        <SurfacePanel className="overflow-hidden border-none bg-[linear-gradient(160deg,#18324d_0%,#22557f_48%,#d8edf9_100%)] px-6 py-6 text-white shadow-[0_28px_56px_rgba(16,42,67,0.2)]">
          <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-white/72">admin creator review</p>
          <h1 className="mt-3 max-w-xl font-display text-[34px] font-semibold leading-[1.02] tracking-[-0.05em]">
            Creator 審査申請を local admin から確認する
          </h1>
          <p className="mt-3 max-w-2xl text-sm leading-6 text-white/80">
            submit 済みの申請者データを確認し、approve / reject / suspend を backend transport に沿って反映します。
          </p>
        </SurfacePanel>

        <SurfacePanel className="bg-[linear-gradient(180deg,#ffffff,#f8fbff)] px-5 py-5 text-foreground">
          <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">flow</p>
          <ol className="mt-4 grid gap-3 text-sm leading-6 text-muted">
            <li className="rounded-[18px] border border-border bg-white px-4 py-3">
              1. queue から対象 creator を選択
            </li>
            <li className="rounded-[18px] border border-border bg-white px-4 py-3">
              2. intake / evidence / review timeline を確認
            </li>
            <li className="rounded-[18px] border border-border bg-white px-4 py-3">
              3. 必要なら reject 理由を指定して decision を反映
            </li>
          </ol>
        </SurfacePanel>
      </section>

      <section className="flex justify-start">
        <SegmentedControl
          ariaLabel="Admin creator review state"
          items={queueTabs.map((tab) => ({
            active: activeState === tab.key,
            href: `/admin/creator-reviews?state=${tab.key}`,
            key: tab.key,
            label: tab.label,
          }))}
        />
      </section>

      <section className="grid gap-4">
        {queue.items.length === 0 ? (
          <SurfacePanel className="px-6 py-10 text-center text-foreground">
            <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">
              {getCreatorReviewStateLabel(queue.state)}
            </p>
            <h2 className="mt-3 font-display text-[28px] font-semibold tracking-[-0.04em]">
              対象の申請はありません
            </h2>
            <p className="mt-3 text-sm leading-6 text-muted">
              dev seed を適用すると submitted queue のサンプル申請を確認できます。
            </p>
          </SurfacePanel>
        ) : (
          queue.items.map((item) => (
            <Link
              className="block"
              href={`/admin/creator-reviews/${item.userId}?state=${queue.state}`}
              key={item.userId}
            >
              <SurfacePanel className="group px-5 py-5 text-foreground transition hover:-translate-y-0.5 hover:shadow-[0_24px_48px_rgba(15,23,42,0.1)]">
                <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                  <div className="flex min-w-0 items-start gap-4">
                    <Avatar className="size-[68px] border border-border bg-[#eef6fd] text-[17px] font-semibold text-[#255a80] shadow-none">
                      {item.sharedProfile.avatar ? (
                        <AvatarImage
                          alt={`${item.sharedProfile.displayName} avatar`}
                          src={item.sharedProfile.avatar.url}
                        />
                      ) : null}
                      <AvatarFallback className="bg-transparent text-inherit">
                        {buildCreatorReviewAvatarFallback(item.sharedProfile.displayName)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <p className="text-[18px] font-semibold tracking-[-0.03em] text-foreground">
                          {item.sharedProfile.displayName}
                        </p>
                        <span
                          className={[
                            "rounded-full border px-3 py-1 text-[11px] font-bold uppercase tracking-[0.16em]",
                            getStateBadgeClass(item.state),
                          ].join(" ")}
                        >
                          {getCreatorReviewStateLabel(item.state)}
                        </span>
                      </div>
                      <p className="mt-1 text-sm text-muted">{item.sharedProfile.handle}</p>
                      <p className="mt-3 line-clamp-2 text-sm leading-6 text-muted">{item.creatorBio}</p>
                    </div>
                  </div>

                  <dl className="grid gap-3 text-sm leading-6 text-muted sm:grid-cols-2 lg:min-w-[360px]">
                    <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
                      <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">legal name</dt>
                      <dd className="mt-1 text-foreground">{item.legalName}</dd>
                    </div>
                    <div className="rounded-[18px] border border-border bg-[#f8fbfe] px-4 py-3">
                      <dt className="text-[11px] font-bold uppercase tracking-[0.18em] text-accent-ink">submitted</dt>
                      <dd className="mt-1 text-foreground">
                        {formatCreatorReviewTimestamp(item.review.submittedAt)}
                      </dd>
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
