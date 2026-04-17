"use client";

import Link from "next/link";
import { useRef } from "react";

import { Avatar, AvatarFallback, AvatarImage, Button, SurfacePanel } from "@/shared/ui";

import {
  creatorRegistrationEvidenceKinds,
  type CreatorRegistrationStatus,
} from "../api/contracts";
import { useCreatorRegistration } from "../model/use-creator-registration";
import { CreatorRegistrationStaticWorkspacePreview } from "./creator-registration-static-workspace-preview";

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

const onboardingChecklist = [
  "年齢要件と本人確認書類をそろえる",
  "売上受取主体が確認できる payout proof を出す",
  "禁止カテゴリ非該当と consent responsibility を確認する",
  "approval 前は creator dashboard / upload は開かない",
] as const;

type RegistrationSurfaceKind =
  | "draft"
  | "not_started"
  | "rejected_closed"
  | "rejected_resubmit"
  | "rejected_support"
  | "rejected_unknown"
  | "submitted"
  | "suspended";

type RegistrationPanelCopy = {
  description: string;
  eyebrow: string;
  saveLabel: string;
  submitLabel: string;
  title: string;
};

type RegistrationStatusCard = {
  body: string;
  checklist: readonly string[];
  ctaLabel: string | null;
  eyebrow: string;
  meta: readonly string[];
  title: string;
};

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

function formatStatusTimestamp(timestamp: string | null) {
  if (!timestamp) {
    return null;
  }

  const date = new Date(timestamp);
  if (Number.isNaN(date.valueOf())) {
    return null;
  }

  return `${date.getFullYear()}/${String(date.getMonth() + 1).padStart(2, "0")}/${String(date.getDate()).padStart(2, "0")} ${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function resolveRegistrationSurfaceKind(
  registration: CreatorRegistrationStatus | null,
  registrationState: string | null,
  isReadOnly: boolean,
): RegistrationSurfaceKind {
  if (registration?.state === "submitted" || registrationState === "submitted") {
    return "submitted";
  }
  if (registration?.state === "suspended" || registrationState === "suspended") {
    return "suspended";
  }
  if (registration?.state === "rejected" || registrationState === "rejected") {
    if (registration === null) {
      return isReadOnly ? "rejected_unknown" : "rejected_resubmit";
    }
    if (registration?.actions.canResubmit || !isReadOnly) {
      return "rejected_resubmit";
    }
    if (registration?.rejection?.isSupportReviewRequired) {
      return "rejected_support";
    }

    return "rejected_closed";
  }
  if (registrationState === "draft") {
    return "draft";
  }

  return "not_started";
}

function resolvePanelCopy(surfaceKind: RegistrationSurfaceKind): RegistrationPanelCopy {
  switch (surfaceKind) {
    case "rejected_resubmit":
      return {
        description:
          "却下理由を確認し、必要な項目だけ修正して同じ onboarding flow から再申請できます。creator mode や upload workspace は引き続き開きません。",
        eyebrow: "creator rejected",
        saveLabel: "修正内容を保存する",
        submitLabel: "再申請を送信する",
        title: "修正して再申請する",
      };
    case "rejected_support":
      return {
        description:
          "現在の却下内容は self-serve では解消できません。この surface は read-only のまま next action を確認するために使います。",
        eyebrow: "creator rejected",
        saveLabel: "下書きを保存する",
        submitLabel: "再申請を送信する",
        title: "サポート確認が必要です",
      };
    case "rejected_closed":
      return {
        description:
          "この onboarding case では self-serve resubmit を開けません。却下理由と残回数を確認するための read-only surface として扱います。",
        eyebrow: "creator rejected",
        saveLabel: "下書きを保存する",
        submitLabel: "再申請を送信する",
        title: "再申請は利用できません",
      };
    case "rejected_unknown":
      return {
        description:
          "rejected status の詳細をいま確認できません。support review required か上限到達かをこの surface だけでは判定できないため、時間を置いて再読み込みしてください。",
        eyebrow: "creator rejected",
        saveLabel: "下書きを保存する",
        submitLabel: "再申請を送信する",
        title: "審査状態を再確認してください",
      };
    case "suspended":
      return {
        description:
          "Creator 利用は現在停止中です。review 状態が更新されるまで、creator mode や upload workspace は引き続き開きません。",
        eyebrow: "creator suspended",
        saveLabel: "下書きを保存する",
        submitLabel: "審査申請を送信する",
        title: "Creator利用は停止中です",
      };
    case "draft":
      return {
        description:
          "shared profile の表示名、handle、avatar は fan / creator 共通です。この面では preview のみ行い、必要な証跡と creator 固有の bio を追加します。",
        eyebrow: "creator onboarding",
        saveLabel: "下書きを保存する",
        submitLabel: "審査申請を送信する",
        title: "Creator審査申請を始める",
      };
    case "submitted":
      return {
        description: "審査状況の確認が完了するまでお待ちください。",
        eyebrow: "creator submitted",
        saveLabel: "下書きを保存する",
        submitLabel: "審査申請を送信する",
        title: "審査申請を受け付けました",
      };
    case "not_started":
    default:
      return {
        description:
          "shared profile の表示名、handle、avatar は fan / creator 共通です。この面では preview のみ行い、必要な証跡と creator 固有の bio を追加します。",
        eyebrow: "creator onboarding",
        saveLabel: "下書きを保存する",
        submitLabel: "審査申請を送信する",
        title: "Creator審査申請を始める",
      };
  }
}

function resolveReasonChecklist(reasonCode: string | null): readonly string[] {
  switch (reasonCode) {
    case "documents_incomplete":
      return [
        "不足している証跡をアップロードし直す",
        "legal name / birth date / payout recipient の空欄を埋める",
        "保存後に同じ flow から再申請する",
      ];
    case "documents_blurry":
      return [
        "判別できる解像度で本人確認書類を再提出する",
        "見切れや反射がないファイルへ差し替える",
        "更新後に再申請する",
      ];
    case "payout_info_incomplete":
      return [
        "payout recipient type と受取名義を見直す",
        "payout proof を最新ファイルへ差し替える",
        "修正後に再申請する",
      ];
    case "profile_mismatch":
      return [
        "shared profile の表示名と提出情報の不一致を直す",
        "必要なら profile settings で表示名や handle を更新する",
        "反映後に再申請する",
      ];
    case "impersonation_suspected":
      return [
        "現在は self-serve での再申請はできません",
        "review 側の確認が終わるまで read-only のままです",
      ];
    default:
      return [
        "status explanation を確認する",
        "必要な項目が分かる場合だけ修正し、案内に従って再申請する",
      ];
  }
}

function resolveStatusCard(
  registration: CreatorRegistrationStatus | null,
  surfaceKind: RegistrationSurfaceKind,
): RegistrationStatusCard | null {
  if (registration === null) {
    if (surfaceKind !== "rejected_resubmit") {
      return null;
    }

    return {
      body:
        "rejected detail の再取得には失敗しましたが、self-serve resubmit 自体は開いています。差し戻しになった項目を中心に見直し、同じ onboarding surface から再申請してください。",
      checklist: [
        "差し戻しになった入力と証跡を見直す",
        "保存後に同じ flow から再申請する",
      ],
      ctaLabel: "Edit And Resubmit",
      eyebrow: "Resubmit available",
      meta: [],
      title: "修正ポイントを確認して再申請する",
    };
  }

  if (!registration) {
    return null;
  }

  if (registration.state === "rejected") {
    const rejectedAt = formatStatusTimestamp(registration.review.rejectedAt);
    const remaining = registration.rejection?.selfServeResubmitRemaining ?? 0;
    const reasonCode = registration.rejection?.reasonCode ?? null;
    const showRemaining = registration.rejection !== null && !registration.rejection.isSupportReviewRequired;
    const meta = [
      rejectedAt ? `Rejected at: ${rejectedAt}` : null,
      showRemaining
        ? `Self-serve remaining: ${remaining}`
        : null,
    ].filter((value): value is string => value !== null);

    if (registration.actions.canResubmit) {
      return {
        body:
          "fixable reject として判定されています。必要な項目だけ修正し、同じ onboarding surface から self-serve で再申請できます。",
        checklist: resolveReasonChecklist(reasonCode),
        ctaLabel: "Edit And Resubmit",
        eyebrow: "Resubmit available",
        meta,
        title: "再申請できます",
      };
    }

    return {
      body:
        registration.rejection?.isSupportReviewRequired
          ? "support review required の状態です。いまは self-serve では再申請できず、この surface は確認用の read-only です。"
          : "self-serve resubmit の残回数がなく、この onboarding case では再申請できません。next action の案内がある場合だけ従ってください。",
      checklist: resolveReasonChecklist(reasonCode),
      ctaLabel: null,
      eyebrow: registration.rejection?.isSupportReviewRequired ? "Review required" : "Resubmit unavailable",
      meta: registration.rejection?.isSupportReviewRequired
        ? [...meta, "Next action: support review required"]
        : meta,
      title: registration.rejection?.isSupportReviewRequired ? "運営確認が必要です" : "再申請は利用できません",
    };
  }

  if (registration.state === "suspended") {
    const suspendedAt = formatStatusTimestamp(registration.review.suspendedAt);

    return {
      body:
        "creator capability は停止中です。approval 前 surface だけを read-only で表示し、creator mode / upload / submission package は引き続き閉じたままにします。",
      checklist: [
        "creator mode は再開まで利用できません",
        "この surface では内容確認のみ行えます",
      ],
      ctaLabel: null,
      eyebrow: "Suspended",
      meta: suspendedAt ? [`Suspended at: ${suspendedAt}`] : [],
      title: "停止中のため再申請できません",
    };
  }

  return null;
}

/**
 * fan profile から始める creator registration intake panel を表示する。
 */
export function CreatorRegistrationPanel({
  initialRegistration = null,
}: {
  initialRegistration?: CreatorRegistrationStatus | null;
}) {
  const evidenceInputRefs = useRef<Record<string, HTMLInputElement | null>>({});
  const formRef = useRef<HTMLFormElement | null>(null);
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
    registration,
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
  } = useCreatorRegistration(initialRegistration);

  const surfaceKind = resolveRegistrationSurfaceKind(registration, registrationState, isReadOnly);
  const panelCopy = resolvePanelCopy(surfaceKind);
  const statusCard = resolveStatusCard(registration, surfaceKind);
  const showFormActions = hasLoaded && !isReadOnly && surfaceKind !== "submitted";

  const scrollToForm = () => {
    formRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
    document.getElementById("creator-registration-bio")?.focus();
  };

  return (
    <main className="mx-auto flex min-h-full w-full max-w-[440px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">
          {panelCopy.eyebrow}
        </p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          {panelCopy.title}
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted">{panelCopy.description}</p>

        {isLoading ? (
          <div className="mt-6 rounded-[24px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-5 text-sm leading-6 text-muted">
            申請フォームを読み込んでいます...
          </div>
        ) : null}

        {statusCard ? (
          <section className="mt-6 rounded-[24px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-4 text-foreground">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
              {statusCard.eyebrow}
            </p>
            <h2 className="mt-2 text-[20px] font-semibold tracking-[-0.02em] text-foreground">
              {statusCard.title}
            </h2>
            <p className="mt-2 text-sm leading-6 text-muted">{statusCard.body}</p>

            {statusCard.meta.length > 0 ? (
              <div className="mt-4 rounded-[18px] border border-white/90 bg-white/92 px-4 py-4 text-sm leading-6 text-muted">
                {statusCard.meta.map((item) => (
                  <p key={item}>{item}</p>
                ))}
              </div>
            ) : null}

            <div className="mt-4 grid gap-2">
              {statusCard.checklist.map((item) => (
                <div
                  className="rounded-[18px] border border-white/85 bg-white/90 px-4 py-3 text-sm leading-6 text-foreground"
                  key={item}
                >
                  {item}
                </div>
              ))}
            </div>

            {statusCard.ctaLabel ? (
              <div className="mt-4">
                <Button className="w-full" disabled={isBusy || isReadOnly} onClick={scrollToForm} type="button">
                  {statusCard.ctaLabel}
                </Button>
              </div>
            ) : null}
          </section>
        ) : null}

        <section className="mt-6 rounded-[24px] border border-[#d7e7ef] bg-[#f7fbfd] px-4 py-4 text-foreground">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-accent-strong">
            Onboarding checklist
          </p>
          <div className="mt-4 grid gap-2">
            {onboardingChecklist.map((item) => (
              <div
                className="rounded-[18px] border border-white/90 bg-white/92 px-4 py-3 text-sm leading-6 text-foreground"
                key={item}
              >
                {item}
              </div>
            ))}
          </div>
        </section>

        <CreatorRegistrationStaticWorkspacePreview />

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
          ref={formRef}
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

          {showFormActions ? (
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
                {isSaving ? "保存中..." : panelCopy.saveLabel}
              </Button>

              <Button className="w-full" disabled={submitDisabled} type="submit">
                {isSubmitting ? "送信中..." : panelCopy.submitLabel}
              </Button>
            </div>
          ) : null}
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
