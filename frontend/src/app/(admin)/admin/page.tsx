import Link from "next/link";
import { ShieldCheck } from "lucide-react";

import { Button } from "@/shared/ui";

export default function AdminPage() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-col gap-6 px-6 py-8 sm:px-10">
      <section className="rounded-[2rem] border border-white/80 bg-[#18100d] p-8 text-white shadow-[0_24px_80px_rgba(35,16,8,0.26)]">
        <div className="flex items-center gap-3">
          <span className="flex size-12 items-center justify-center rounded-2xl bg-white/10">
            <ShieldCheck className="size-6" />
          </span>
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-200/80">admin / isolated shell</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">consumer shell とは分離した admin route。</h1>
          </div>
        </div>
        <p className="mt-5 max-w-3xl text-sm leading-7 text-stone-300">
          moderation、creator support、運営 KPI などは consumer 導線と混ぜない前提で route group を分けています。
          今回は画面箱だけを用意し、権限制御や業務 UI には踏み込みません。
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Button asChild>
            <Link href="/shorts">consumer に戻る</Link>
          </Button>
          <Button asChild variant="secondary">
            <Link href="/home">home を確認</Link>
          </Button>
        </div>
      </section>
    </main>
  );
}
