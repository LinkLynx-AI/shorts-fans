"use client";

import Link from "next/link";
import { useRef } from "react";

import { Avatar, AvatarFallback, AvatarImage, Button, SurfacePanel } from "@/shared/ui";

import { creatorRegistrationEvidenceKinds } from "../api/contracts";
import { useCreatorRegistration } from "../model/use-creator-registration";

const evidenceFieldLabels = {
  government_id: {
    description: "本人確認書類。JPEG / PNG / WebP / PDF、10MB まで。",
    label: "Government ID",
  },
  payout_proof: {
    description: "売上受取名義が確認できる書類。JPEG / PNG / WebP / PDF、10MB まで。",
    label: "Payout Proof",
  },
} as const;

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

function formatEvidenceDate(uploadedAt: string) {
  const date = new Date(uploadedAt);
  if (Number.isNaN(date.valueOf())) {
    return "アップロード済み";
  }

  return `${date.getFullYear()}/${String(date.getMonth() + 1).padStart(2, "0")}/${String(date.getDate()).padStart(2, "0")}`;
}

/**
 * fan profile から始める creator registration intake panel を表示する。
 */
export function CreatorRegistrationPanel() {
  const evidenceInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  const {
    acceptsConsentResponsibility,
    birthDate,
    creatorBio,
    declaresNoProhibitedCategory,
    errorMessage,
    evidences,
    hasLoaded,
    isBusy,
    isLoading,
    isReadOnly,
    isSaving,
    isSubmitting,
    legalName,
    payoutRecipientName,
    payoutRecipientType,
    registrationState,
    saveDraft,
    setAcceptsConsentResponsibility,
    setBirthDate,
    setCreatorBio,
    setDeclaresNoProhibitedCategory,
    setLegalName,
    setPayoutRecipientName,
    setPayoutRecipientType,
    sharedProfile,
    submit,
    submitDisabled,
    successMessage,
    uploadEvidence,
  } = useCreatorRegistration();

  return (
    <main className="mx-auto flex min-h-full w-full max-w-[440px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">
          creator onboarding
        </p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          Creator審査申請を始める
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted">
          shared profile の表示名、handle、avatar は fan / creator 共通です。
          この面では preview のみ行い、必要な証跡と creator 固有の bio を追加します。
        </p>

        {isLoading ? (
          <div className="mt-6 rounded-[24px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-5 text-sm leading-6 text-muted">
            申請フォームを読み込んでいます...
          </div>
        ) : null}

        {sharedProfile ? (
          <section className="mt-6 rounded-[24px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-4 text-foreground">
            <div className="flex items-center gap-4">
              <Avatar className="size-[72px] border border-[#d3e2ea] bg-white text-[18px] font-semibold text-[#2f6176] shadow-none">
                {sharedProfile.avatar ? <AvatarImage alt={`${sharedProfile.displayName} avatar`} src={sharedProfile.avatar.url} /> : null}
                <AvatarFallback className="bg-transparent text-inherit">
                  {buildAvatarFallback(sharedProfile.displayName)}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                  Shared profile
                </p>
                <p className="mt-1 text-[17px] font-semibold tracking-[-0.02em] text-foreground">
                  {sharedProfile.displayName}
                </p>
                <p className="mt-1 text-sm text-muted">{sharedProfile.handle}</p>
              </div>
            </div>
            <div className="mt-4 rounded-[18px] border border-white/80 bg-white/90 px-4 py-3 text-sm leading-6 text-muted">
              表示名、handle、avatar を直したい場合は profile settings で更新してください。
            </div>
            <div className="mt-4">
              <Button asChild className="w-full" disabled={isBusy} variant="secondary">
                <Link href="/fan/settings/profile">Profile settings を開く</Link>
              </Button>
            </div>
          </section>
        ) : null}

        <form
          className="mt-6 grid gap-5"
          onSubmit={(event) => {
            event.preventDefault();
            void submit();
          }}
        >
          <section className="grid gap-4">
            <label className="grid gap-1.5" htmlFor="creator-registration-bio">
              <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                Bio
              </span>
              <textarea
                className="min-h-[132px] w-full resize-none rounded-[22px] border border-[#d7e7ef] bg-white px-4 py-3 text-sm leading-6 text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60 disabled:cursor-default disabled:opacity-70"
                disabled={isBusy || isReadOnly}
                id="creator-registration-bio"
                onChange={(event) => setCreatorBio(event.target.value)}
                placeholder="投稿したい世界観や自分の紹介を記入してください。"
                value={creatorBio}
              />
            </label>

            <label className="grid gap-1.5" htmlFor="creator-registration-legal-name">
              <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                Legal name
              </span>
              <input
                autoComplete="name"
                className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60 disabled:cursor-default disabled:opacity-70"
                disabled={isBusy || isReadOnly}
                id="creator-registration-legal-name"
                onChange={(event) => setLegalName(event.target.value)}
                placeholder="Mina Rei"
                type="text"
                value={legalName}
              />
            </label>

            <label className="grid gap-1.5" htmlFor="creator-registration-birth-date">
              <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                Birth date
              </span>
              <input
                className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60 disabled:cursor-default disabled:opacity-70"
                disabled={isBusy || isReadOnly}
                id="creator-registration-birth-date"
                onChange={(event) => setBirthDate(event.target.value)}
                type="date"
                value={birthDate}
              />
            </label>

            <fieldset className="grid gap-2">
              <legend className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                Payout recipient
              </legend>
              <div className="grid gap-2 sm:grid-cols-2">
                {[
                  { label: "自分名義", value: "self" },
                  { label: "事業名義", value: "business" },
                ].map((option) => (
                  <label
                    className="flex items-start gap-3 rounded-[20px] border border-[#d7e7ef] bg-[#f8fbfd] px-4 py-4 text-sm leading-6 text-foreground"
                    htmlFor={`creator-registration-payout-${option.value}`}
                    key={option.value}
                  >
                    <input
                      checked={payoutRecipientType === option.value}
                      className="mt-1 size-5 rounded-[6px] border border-[#cfd7e3] accent-accent-strong"
                      disabled={isBusy || isReadOnly}
                      id={`creator-registration-payout-${option.value}`}
                      name="creator-registration-payout-type"
                      onChange={() => setPayoutRecipientType(option.value)}
                      type="radio"
                    />
                    <span>{option.label}</span>
                  </label>
                ))}
              </div>
            </fieldset>

            <label className="grid gap-1.5" htmlFor="creator-registration-payout-name">
              <span className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
                Payout recipient name
              </span>
              <input
                className="min-h-12 rounded-[18px] border border-[#bae7ff]/90 bg-white/88 px-4 text-sm text-foreground outline-none transition placeholder:text-muted focus:border-accent focus:ring-4 focus:ring-ring/60 disabled:cursor-default disabled:opacity-70"
                disabled={isBusy || isReadOnly}
                id="creator-registration-payout-name"
                onChange={(event) => setPayoutRecipientName(event.target.value)}
                placeholder="Mina Rei"
                type="text"
                value={payoutRecipientName}
              />
            </label>
          </section>

          <section className="grid gap-3">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
              Evidence
            </p>
            {creatorRegistrationEvidenceKinds.map((kind) => {
              const field = evidences[kind];
              const config = evidenceFieldLabels[kind];
              const evidenceUploadDisabled = isBusy || isReadOnly || field.isUploading;

              return (
                <section
                  className="rounded-[22px] border border-[#d7e7ef] bg-[#f8fbfd] px-4 py-4 text-foreground"
                  key={kind}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <p className="text-[15px] font-semibold tracking-[-0.02em] text-foreground">{config.label}</p>
                      <p className="mt-1 text-sm leading-6 text-muted">{config.description}</p>
                    </div>
                    <span className="text-[11px] font-semibold uppercase tracking-[0.16em] text-accent-strong">
                      required
                    </span>
                  </div>

                  <div className="mt-4 rounded-[18px] border border-white/90 bg-white/92 px-4 py-4 text-sm leading-6 text-muted">
                    {field.evidence ? (
                      <>
                        <p className="font-semibold text-foreground">{field.evidence.fileName}</p>
                        <p className="mt-1">
                          {field.evidence.mimeType} / {Math.ceil(field.evidence.fileSizeBytes / 1024)}KB / {formatEvidenceDate(field.evidence.uploadedAt)}
                        </p>
                      </>
                    ) : (
                      <p>まだアップロードされていません。</p>
                    )}
                  </div>

                  {field.errorMessage ? (
                    <p
                      aria-live="polite"
                      className="mt-3 rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
                      role="alert"
                    >
                      {field.errorMessage}
                    </p>
                  ) : null}

                  <div className="mt-4">
                    <Button
                      className="w-full"
                      disabled={evidenceUploadDisabled}
                      onClick={() => {
                        evidenceInputRefs.current[kind]?.click();
                      }}
                      type="button"
                      variant="secondary"
                    >
                      {field.isUploading ? "アップロード中..." : field.evidence ? "証跡を差し替える" : "証跡を選択する"}
                    </Button>
                    <input
                      accept="image/jpeg,image/png,image/webp,application/pdf"
                      className="sr-only"
                      disabled={evidenceUploadDisabled}
                      id={`creator-registration-evidence-${kind}`}
                      key={field.inputKey}
                      onChange={(event) => {
                        void uploadEvidence(kind, event.target.files?.[0] ?? null);
                      }}
                      ref={(node) => {
                        evidenceInputRefs.current[kind] = node;
                      }}
                      type="file"
                    />
                  </div>
                </section>
              );
            })}
          </section>

          <section className="grid gap-3">
            <label className="flex items-start gap-3 rounded-[20px] border border-[#d7e7ef] bg-[#f8fbfd] px-4 py-4">
              <input
                checked={declaresNoProhibitedCategory}
                className="mt-1 size-5 rounded-[6px] border border-[#cfd7e3] accent-accent-strong"
                disabled={isBusy || isReadOnly}
                onChange={(event) => setDeclaresNoProhibitedCategory(event.target.checked)}
                type="checkbox"
              />
              <span className="text-sm leading-6 text-foreground">
                prohibited category に該当する content を扱わないことを確認しました。
              </span>
            </label>

            <label className="flex items-start gap-3 rounded-[20px] border border-[#d7e7ef] bg-[#f8fbfd] px-4 py-4">
              <input
                checked={acceptsConsentResponsibility}
                className="mt-1 size-5 rounded-[6px] border border-[#cfd7e3] accent-accent-strong"
                disabled={isBusy || isReadOnly}
                onChange={(event) => setAcceptsConsentResponsibility(event.target.checked)}
                type="checkbox"
              />
              <span className="text-sm leading-6 text-foreground">
                出演者 consent と権利関係の責任を自分で負うことを確認しました。
              </span>
            </label>
          </section>

          {registrationState && registrationState !== "draft" && registrationState !== "submitted" ? (
            <p className="rounded-[18px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-3 text-sm leading-6 text-muted">
              現在の申請状態は `{registrationState}` です。この状態の詳細 surface は後続 task で実装予定のため、ここでは編集を停止しています。
            </p>
          ) : null}

          {successMessage ? (
            <p className="rounded-[18px] border border-[rgba(26,152,80,0.22)] bg-[rgba(245,255,249,0.95)] px-4 py-3 text-sm leading-6 text-[#197040]">
              {successMessage}
            </p>
          ) : null}

          {errorMessage ? (
            <p
              aria-live="polite"
              className="rounded-[18px] border border-[#ffb3b8] bg-[#fff4f5] px-4 py-3 text-sm leading-6 text-[#b2394f]"
              role="alert"
            >
              {errorMessage}
            </p>
          ) : null}

          {!hasLoaded ? null : (
            <div className="grid gap-3">
              <Button
                className="w-full"
                disabled={isBusy || isReadOnly}
                onClick={() => {
                  void saveDraft();
                }}
                type="button"
                variant="secondary"
              >
                {isSaving ? "下書きを保存中..." : "下書きを保存する"}
              </Button>

              <Button className="w-full" disabled={submitDisabled} type="submit">
                {isSubmitting ? "申請を送信中..." : "審査申請を送信する"}
              </Button>
            </div>
          )}
        </form>

        <div className="mt-4">
          <Button asChild className="w-full" disabled={isBusy} variant="secondary">
            <Link href="/fan">あとで fan に戻る</Link>
          </Button>
        </div>
      </SurfacePanel>
    </main>
  );
}
