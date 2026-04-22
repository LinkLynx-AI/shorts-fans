import {
  ArrowRight,
  Clapperboard,
  IdCard,
} from "lucide-react";
import Link from "next/link";

import { assertAdminUiAccess } from "./_lib/admin-ui-access";
import { AdminReviewNavigation } from "./_ui/admin-review-navigation";
import {
  Button,
  SurfacePanel,
} from "@/shared/ui";

const reviewSurfaces = [
  {
    accentClass: "border-[#d7e6f5] bg-[#f5faff] text-[#1f628f]",
    description: "creator registration intake と evidence を確認し、approve / reject / suspend を反映します。",
    href: "/admin/creator-reviews",
    icon: IdCard,
    label: "Creator 審査",
    title: "登録申請を確認する",
  },
  {
    accentClass: "border-[#f0deba] bg-[#fff8eb] text-[#8a5a00]",
    description: "canonical main と linked short を確認し、object-level decision を反映します。",
    href: "/admin/submission-reviews",
    icon: Clapperboard,
    label: "Video 審査",
    title: "投稿 video を審査する",
  },
] as const;

export default async function AdminPage() {
  await assertAdminUiAccess();

  return (
    <main className="mx-auto flex min-h-full w-full max-w-6xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
      <AdminReviewNavigation active="home" />

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)]">
        <SurfacePanel className="overflow-hidden border-none bg-[linear-gradient(155deg,#162938_0%,#376b75_48%,#e0f0e7_100%)] px-6 py-6 text-white shadow-[0_28px_56px_rgba(22,41,56,0.18)]">
          <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-white/72">local admin</p>
          <h1 className="mt-3 max-w-xl font-display text-[34px] font-semibold leading-[1.02] tracking-[-0.05em]">
            審査 surface を選択する
          </h1>
          <p className="mt-3 max-w-2xl text-sm leading-6 text-white/82">
            creator 登録審査と video submission 審査を local admin から開けるようにします。
          </p>
        </SurfacePanel>

        <SurfacePanel className="bg-[linear-gradient(180deg,#ffffff,#f7fbf8)] px-5 py-5 text-foreground">
          <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent-ink">review queues</p>
          <div className="mt-4 grid gap-3 text-sm leading-6 text-muted">
            <div className="rounded-[18px] border border-border bg-white px-4 py-3">
              Creator 審査は user / intake / evidence を確認します。
            </div>
            <div className="rounded-[18px] border border-border bg-white px-4 py-3">
              Video 審査は main / short を package 単位で確認します。
            </div>
          </div>
        </SurfacePanel>
      </section>

      <section className="grid gap-4 md:grid-cols-2">
        {reviewSurfaces.map((surface) => {
          const Icon = surface.icon;

          return (
            <SurfacePanel className="px-5 py-5 text-foreground" key={surface.href}>
              <div className="flex items-start gap-4">
                <span
                  className={[
                    "flex size-11 shrink-0 items-center justify-center rounded-full border",
                    surface.accentClass,
                  ].join(" ")}
                >
                  <Icon aria-hidden="true" size={20} strokeWidth={2.2} />
                </span>
                <div className="min-w-0">
                  <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-accent-ink">{surface.label}</p>
                  <h2 className="mt-2 font-display text-[24px] font-semibold leading-[1.12] tracking-[-0.03em]">
                    {surface.title}
                  </h2>
                  <p className="mt-3 text-sm leading-6 text-muted">{surface.description}</p>
                </div>
              </div>
              <Button asChild className="mt-5" variant="secondary">
                <Link href={surface.href} prefetch={false}>
                  開く
                  <ArrowRight aria-hidden="true" size={16} strokeWidth={2.2} />
                </Link>
              </Button>
            </SurfacePanel>
          );
        })}
      </section>
    </main>
  );
}
