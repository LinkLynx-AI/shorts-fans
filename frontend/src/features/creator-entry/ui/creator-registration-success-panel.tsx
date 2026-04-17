"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { CheckSquare, ChevronLeft } from "lucide-react";
import { startTransition, useEffect, useState } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui";

import { fetchCreatorRegistration } from "../api/fetch-creator-registration";
import type { CreatorRegistrationStatus } from "../api/contracts";
import {
  getCreatorEntryErrorCode,
  getCreatorRegistrationErrorMessage,
} from "../model/creator-entry";
import {
  CreatorRegistrationMessage,
} from "./creator-registration-ui-primitives";

function buildAvatarFallback(displayName: string) {
  const trimmed = displayName.trim();
  if (trimmed === "") {
    return "ME";
  }

  return trimmed
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.at(0)?.toUpperCase() ?? "")
    .join("");
}

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
    <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
      <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
        現在の状態
      </p>
      <div className="mt-4 space-y-3">
        <div className="rounded-[22px] border border-gray-100 bg-[#f8f9fc] px-4 py-4">
          <p className="text-[12px] font-black tracking-[0.08em] text-[#a3adbc]">
            申請状況
          </p>
          <p className="mt-1 text-[16px] font-bold text-foreground">確認中</p>
        </div>
        <div className="rounded-[22px] border border-gray-100 bg-[#f8f9fc] px-4 py-4">
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
  const profilePreview = registration?.sharedProfile ?? null;

  return (
    <main className="relative mx-auto flex min-h-svh w-full max-w-[408px] flex-col bg-[#f7f8fb] text-foreground">
      <div className="sticky top-0 z-20 flex items-center justify-between border-b border-[#eef1f5] bg-white/90 px-4 pb-4 pt-14 shadow-sm backdrop-blur-md">
        <Link aria-label="戻る" className="text-gray-800" href="/fan">
          <ChevronLeft className="size-7" />
        </Link>
        <span className="text-[17px] font-bold text-foreground">申請完了</span>
        <div className="w-7" />
      </div>

      <div className="flex-1 overflow-y-auto px-5 pb-32 pt-5">
        <div className="space-y-6">
          <section className="rounded-[28px] border border-[#d7ebfb] bg-[#f4fafe] p-5 shadow-sm">
            <div className="flex items-start gap-3">
              <div className="mt-0.5 rounded-full bg-white p-2 text-[#4DA8DA]">
                <CheckSquare className="size-5" strokeWidth={2.2} />
              </div>
              <div>
                <h1 className="text-[18px] font-extrabold leading-tight text-[#1f5f86]">
                  申請を受け付けました
                </h1>
                <p className="mt-1 text-[13px] font-medium leading-relaxed text-[#4e7f9c]">
                  確認が終わるまでしばらくお待ちください。
                </p>
              </div>
            </div>
          </section>

          <SuccessPanelSummary isLoading={isLoading} submittedAt={submittedAt} />

          {profilePreview ? (
            <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
              <div className="flex items-center justify-between gap-4">
                <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
                  プロフィール
                </p>
                <Link
                  className="rounded-full bg-[#eef6ff] px-3 py-1.5 text-[12px] font-bold text-[#3380d6] transition-colors hover:bg-[#e3f0ff]"
                  href="/fan/settings/profile"
                >
                  編集する
                </Link>
              </div>

              <div className="mt-4 flex items-center gap-4">
                <Avatar className="size-16 border border-gray-100 shadow-sm">
                  {profilePreview.avatar ? (
                    <AvatarImage alt={`${profilePreview.displayName} の画像`} src={profilePreview.avatar.url} />
                  ) : null}
                  <AvatarFallback className="bg-[#f3f5f8] text-[17px] font-semibold text-[#486270]">
                    {buildAvatarFallback(profilePreview.displayName)}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <p className="text-[16px] font-extrabold text-foreground">
                    {profilePreview.displayName}
                  </p>
                  <p className="mt-0.5 text-[13px] font-medium text-muted">
                    {profilePreview.handle}
                  </p>
                </div>
              </div>
            </section>
          ) : null}

          {errorMessage ? (
            <CreatorRegistrationMessage kind="error" message={errorMessage} />
          ) : null}
        </div>
      </div>

      <div className="absolute bottom-0 left-0 z-30 w-full border-t border-gray-100 bg-white/95 px-4 pb-8 pt-4 shadow-[0_-10px_20px_rgba(0,0,0,0.03)] backdrop-blur-md">
        <Link
          className="block w-full rounded-full bg-[#4DA8DA] py-4 text-center text-[16px] font-bold text-white shadow-lg shadow-[#4DA8DA]/20 transition-transform active:scale-[0.98]"
          href="/fan"
        >
          ホームに戻る
        </Link>
      </div>
    </main>
  );
}
