"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  CheckCircle2,
  Clock3,
} from "lucide-react";
import { startTransition, useEffect, useState } from "react";

import { Button, SurfacePanel } from "@/shared/ui";

import { fetchCreatorRegistration } from "../api/fetch-creator-registration";
import type { CreatorRegistrationStatus } from "../api/contracts";
import {
  getCreatorEntryErrorCode,
  getCreatorRegistrationErrorMessage,
} from "../model/creator-entry";

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

function getSuccessPanelErrorMessage(error: unknown) {
  if (getCreatorEntryErrorCode(error) === "not_found") {
    return "プロフィール設定をご確認ください。";
  }

  return getCreatorRegistrationErrorMessage(error);
}

function SuccessPanelSummary({
  isLoading,
  submittedAt,
}: {
  isLoading: boolean;
  submittedAt: string | null;
}) {
  if (isLoading) {
    return (
      <div className="mt-5 rounded-[22px] border border-[#dce6ee] bg-[linear-gradient(180deg,#f8fbfd_0%,#f4f9fc_100%)] px-4 py-4">
        <p aria-live="polite" className="text-[13px] font-semibold text-foreground" role="status">
          受付情報を確認しています
        </p>
        <div aria-hidden="true" className="mt-3 h-3 w-28 animate-pulse rounded-full bg-[rgba(167,220,249,0.34)]" />
      </div>
    );
  }

  if (!submittedAt) {
    return null;
  }

  return (
    <div className="mt-5 rounded-[22px] border border-[#dce6ee] bg-[linear-gradient(180deg,#f8fbfd_0%,#f4f9fc_100%)] px-4 py-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-[12px] font-semibold text-muted">現在の状態</p>
          <p className="mt-1 text-[15px] font-semibold text-foreground">確認待ち</p>
        </div>
        <span className="inline-flex size-10 items-center justify-center rounded-full bg-white text-[#0f6172] shadow-[0_8px_18px_rgba(15,23,42,0.04)]">
          <Clock3 aria-hidden="true" className="size-[18px]" strokeWidth={2.1} />
        </span>
      </div>
      <div className="mt-4 rounded-[18px] bg-white/88 px-4 py-3">
        <p className="text-[12px] font-semibold text-muted">受付日時</p>
        <p className="mt-1 text-[14px] font-semibold text-foreground">{submittedAt}</p>
      </div>
    </div>
  );
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
          setErrorMessage(getSuccessPanelErrorMessage(error));
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
    <main className="mx-auto flex min-h-full w-full max-w-[408px] flex-col px-4 pb-28 pt-6">
      <SurfacePanel className="w-full overflow-hidden px-5 py-5 text-foreground">
        <div className="rounded-[24px] border border-[rgba(113,180,234,0.22)] bg-[linear-gradient(180deg,rgba(236,246,255,0.96)_0%,rgba(248,251,255,0.98)_100%)] px-4 py-4">
          <p className="inline-flex rounded-full border border-white/80 bg-white/88 px-3 py-1 text-[11px] font-semibold tracking-[0.16em] text-[#0f6172] shadow-[0_8px_20px_rgba(15,23,42,0.04)]">
            受付完了
          </p>
          <div className="mt-4 flex items-start gap-3">
            <span className="inline-flex size-12 shrink-0 items-center justify-center rounded-full bg-[linear-gradient(135deg,var(--accent)_0%,var(--accent-strong)_100%)] text-white shadow-[0_12px_24px_rgba(80,159,224,0.22)]">
              <CheckCircle2 aria-hidden="true" className="size-[22px]" strokeWidth={2.4} />
            </span>
            <div className="min-w-0">
              <h1 className="font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em] text-foreground">
                申請を受け付けました
              </h1>
              <p className="mt-2 text-sm leading-6 text-muted">
                確認が終わるまで、クリエイターとしての利用はまだ始まりません。
              </p>
            </div>
          </div>
        </div>

        <SuccessPanelSummary isLoading={isLoading} submittedAt={submittedAt} />

        {errorMessage ? (
          <p
            aria-live="polite"
            className="mt-5 rounded-[18px] border border-[#ffb3b8] bg-[linear-gradient(180deg,#fff7f8_0%,#fff2f4_100%)] px-4 py-3 text-sm leading-6 text-[#b2394f]"
            role="alert"
          >
            {errorMessage}
          </p>
        ) : null}

        <div className="mt-5 grid gap-3">
          <Button asChild className="w-full" variant="secondary">
            <Link href="/fan/settings/profile">プロフィール設定</Link>
          </Button>

          <Button asChild className="w-full">
            <Link href="/fan">ホームへ戻る</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
  );
}
