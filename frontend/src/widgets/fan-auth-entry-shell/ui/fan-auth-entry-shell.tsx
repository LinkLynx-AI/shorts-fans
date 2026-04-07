"use client";

import { useRouter } from "next/navigation";

import { FanAuthEntryPanel } from "@/features/fan-auth";
import { Button } from "@/shared/ui";

/**
 * protected fan surface から到達する fan login entry を表示する。
 */
export function FanAuthEntryShell() {
  const router = useRouter();

  return (
    <main className="mx-auto flex min-h-svh w-full max-w-[408px] items-center px-4 py-10">
      <FanAuthEntryPanel
        dismissAction={({ isSubmitting }) => (
          <Button
            className="w-full"
            disabled={isSubmitting}
            onClick={() => router.push("/")}
            type="button"
            variant="secondary"
          >
            feed に戻る
          </Button>
        )}
      />
    </main>
  );
}
