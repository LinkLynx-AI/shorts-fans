"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { startTransition, useEffect, useState } from "react";

import { Button, SurfacePanel } from "@/shared/ui";

import { fetchCreatorRegistration } from "../api/fetch-creator-registration";
import type { CreatorRegistrationStatus } from "../api/contracts";
import { getCreatorRegistrationErrorMessage } from "../model/creator-entry";

function formatSubmittedAt(submittedAt: string | null) {
  if (!submittedAt) {
    return null;
  }

  const date = new Date(submittedAt);
  if (Number.isNaN(date.valueOf())) {
    return null;
  }

  return `${date.getFullYear()}/${String(date.getMonth() + 1).padStart(2, "0")}/${String(date.getDate()).padStart(2, "0")} ${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

/**
 * creator registration submitted receipt surface を表示する。
 */
export function CreatorRegistrationSuccessPanel() {
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [registration, setRegistration] = useState<CreatorRegistrationStatus | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadRegistration() {
      setIsLoading(true);
      setErrorMessage(null);

      try {
        const currentRegistration = await fetchCreatorRegistration();
        if (cancelled) {
          return;
        }

        if (currentRegistration?.state !== "submitted") {
          startTransition(() => {
            router.replace("/fan/creator/register");
          });
          return;
        }

        setRegistration(currentRegistration);
      } catch (error) {
        if (!cancelled) {
          setErrorMessage(getCreatorRegistrationErrorMessage(error));
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    }

    void loadRegistration();

    return () => {
      cancelled = true;
    };
  }, [router]);

  const submittedAt = formatSubmittedAt(registration?.review.submittedAt ?? null);

  return (
    <main className="mx-auto flex min-h-full w-full max-w-[408px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">
          creator submitted
        </p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          審査申請を受け付けました
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted">
          申請内容と private な証跡を review queue に送信しました。
          approval までは creator mode や upload workspace は開きません。
        </p>

        {isLoading ? (
          <div className="mt-5 rounded-[22px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-4 text-sm leading-6 text-muted">
            申請状況を確認しています...
          </div>
        ) : null}

        {submittedAt ? (
          <div className="mt-5 rounded-[22px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-4 text-sm leading-6 text-muted">
            <p className="font-semibold text-foreground">受付日時</p>
            <p className="mt-1">{submittedAt}</p>
          </div>
        ) : null}

        {errorMessage ? (
          <p
            aria-live="polite"
            className="mt-5 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
            role="alert"
          >
            {errorMessage}
          </p>
        ) : null}

        <div className="mt-5 grid gap-3">
          <Button asChild className="w-full" variant="secondary">
            <Link href="/fan/settings/profile">Profile settings を開く</Link>
          </Button>

          <Button asChild className="w-full">
            <Link href="/fan">fan hub に戻る</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
  );
}
