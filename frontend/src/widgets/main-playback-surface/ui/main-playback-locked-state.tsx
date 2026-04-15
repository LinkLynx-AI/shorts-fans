import Link from "next/link";

import { Button } from "@/shared/ui";

type MainPlaybackLockedStateProps = {
  description?: string;
  fallbackHref: string;
  fallbackLabel?: string;
  title?: string;
};

/**
 * 無効な access で main playback に入ったときの locked state を表示する。
 */
export function MainPlaybackLockedState({
  description = "再生権限のある short から unlock を通る導線で入り直してください。",
  fallbackHref,
  fallbackLabel = "short に戻る",
  title = "この main はまだ unlock されていません。",
}: MainPlaybackLockedStateProps) {
  return (
    <main className="relative flex min-h-full items-center justify-center overflow-hidden bg-[#07111b] px-6 py-12 text-white">
      <div className="absolute inset-0 bg-[linear-gradient(180deg,#11263a_0%,#0d1b2c_38%,#07111b_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(146,217,255,0.22),transparent_34%)]" />
      <div className="relative w-full max-w-sm rounded-[32px] border border-white/12 bg-white/10 p-7 shadow-[0_28px_72px_rgba(0,0,0,0.34)] backdrop-blur-xl">
        <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-white/62">
          main locked
        </p>
        <h1 className="mt-4 text-[30px] font-semibold leading-[1.05] tracking-[-0.05em] text-white">
          {title}
        </h1>
        <p className="mt-4 text-sm leading-7 text-white/76">
          {description}
        </p>
        <div className="mt-8">
          <Button asChild className="w-full bg-white text-[#07111b] hover:bg-white/92">
            <Link href={fallbackHref}>{fallbackLabel}</Link>
          </Button>
        </div>
      </div>
    </main>
  );
}
