import Link from "next/link";

import { FanAuthEntryPanel } from "@/features/fan-auth";
import { Button } from "@/shared/ui";

/**
 * protected fan surface から到達する fan login entry を表示する。
 */
export function FanAuthEntryShell() {
  return (
    <main className="mx-auto flex min-h-svh w-full max-w-[408px] items-center px-4 py-10">
      <FanAuthEntryPanel
        dismissAction={
          <Button asChild className="w-full" variant="secondary">
            <Link href="/">feed に戻る</Link>
          </Button>
        }
      />
    </main>
  );
}
