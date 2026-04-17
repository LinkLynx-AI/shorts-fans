"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { startTransition, useEffect, useState } from "react";

import { Button } from "@/shared/ui";

import { fetchCreatorRegistration } from "../api/fetch-creator-registration";
import type { CreatorRegistrationStatus } from "../api/contracts";
import {
  getCreatorEntryErrorCode,
  getCreatorRegistrationErrorMessage,
} from "../model/creator-entry";
import {
  CreatorRegistrationMessage,
  CreatorRegistrationSectionHeading,
  creatorRegistrationButtonClassName,
  creatorRegistrationInlineSurfaceClassName,
  creatorRegistrationSectionClassName,
  creatorRegistrationShellClassName,
} from "./creator-registration-ui-primitives";

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
      <CreatorRegistrationMessage
        className="mt-6"
        kind="info"
        message="受付情報を確認しています。"
      />
    );
  }

  if (!submittedAt) {
    return null;
  }

  return (
    <section className={`mt-6 ${creatorRegistrationSectionClassName}`}>
      <CreatorRegistrationSectionHeading>
        現在の状態
      </CreatorRegistrationSectionHeading>
      <div className="mt-3 grid gap-3">
        <div className={creatorRegistrationInlineSurfaceClassName}>
          <p className="text-[12px] font-black tracking-[0.08em] text-[#a3adbc]">
            申請状況
          </p>
          <p className="mt-1 text-[16px] font-bold text-foreground">確認中</p>
        </div>
        <div className={creatorRegistrationInlineSurfaceClassName}>
          <p className="text-[12px] font-black tracking-[0.08em] text-[#a3adbc]">
            受付日時
          </p>
          <p className="mt-1 text-[15px] font-bold text-foreground">{submittedAt}</p>
        </div>
      </div>
    </section>
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
    <main className="mx-auto flex min-h-svh w-full max-w-[408px] px-4 py-10">
      <div className={creatorRegistrationShellClassName}>
        <div className="mt-2">
          <h1 className="font-display text-[32px] font-semibold leading-[1.12] tracking-[-0.05em] text-foreground">
            申請を受け付けました
          </h1>
          <p className="mt-3 max-w-[34ch] text-sm leading-6 text-muted">
            確認が終わるまでしばらくお待ちください。利用開始はそのあとです。
          </p>
        </div>

        <SuccessPanelSummary isLoading={isLoading} submittedAt={submittedAt} />

        {errorMessage ? (
          <CreatorRegistrationMessage
            className="mt-6"
            kind="error"
            message={errorMessage}
          />
        ) : null}

        <div className="mt-6 grid gap-3">
          <Button asChild className={creatorRegistrationButtonClassName} variant="secondary">
            <Link href="/fan/settings/profile">プロフィール設定を開く</Link>
          </Button>

          <Button asChild className={creatorRegistrationButtonClassName}>
            <Link href="/fan">ホームに戻る</Link>
          </Button>
        </div>
      </div>
    </main>
  );
}
