import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { CreatorUploadForm } from "@/features/creator-upload";

/**
 * `/creator/upload` の creator-private upload shell を表示する。
 */
export function CreatorUploadShell() {
  return (
    <main className="min-h-svh bg-background sm:px-4">
      <div className="mx-auto flex min-h-svh w-full max-w-[408px] flex-col overflow-hidden bg-white text-foreground sm:border sm:border-border sm:shadow-[var(--device-shadow)]">
        <header className="sticky top-0 z-20 border-b border-[#eef2f6] bg-white/95 backdrop-blur-sm">
          <div className="grid min-h-[74px] grid-cols-[40px_1fr_40px] items-center px-4 pt-[12px]">
            <Link
              aria-label="Back"
              className="inline-flex size-9 items-center justify-center rounded-full text-foreground transition hover:bg-[#f5f8fb]"
              href="/creator"
            >
              <ArrowLeft className="size-5" strokeWidth={2.1} />
            </Link>
            <p className="text-center text-[17px] font-semibold tracking-[-0.02em] text-foreground">新しい動画</p>
            <div className="size-9" />
          </div>
        </header>

        <CreatorUploadForm />
      </div>
    </main>
  );
}
