import Link from "next/link";
import { ArrowLeft } from "lucide-react";

import { CreatorUploadForm } from "@/features/creator-upload";
import { Button } from "@/shared/ui";

/**
 * `/creator/upload` の creator-private upload shell を表示する。
 */
export function CreatorUploadShell() {
  return (
    <main className="bg-background">
      <div className="mx-auto min-h-svh w-full max-w-[408px] overflow-hidden bg-white text-foreground">
        <section className="relative z-[2] min-h-svh overflow-y-auto px-4 pb-10 pt-[14px] text-foreground">
          <div className="flex items-center justify-between gap-3">
            <Button asChild className="-ml-2" size="icon" variant="ghost">
              <Link href="/creator">
                <span className="sr-only">Back</span>
                <ArrowLeft className="size-5" strokeWidth={2.1} />
              </Link>
            </Button>
          </div>

          <section className="mt-[18px] pb-10">
            <CreatorUploadForm />
          </section>
        </section>
      </div>
    </main>
  );
}
