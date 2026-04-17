"use client";

import Link from "next/link";
import { useRef } from "react";

import { Avatar, AvatarFallback, AvatarImage, Button } from "@/shared/ui";

import {
  creatorRegistrationEvidenceKinds,
  type CreatorRegistrationStatus,
} from "../api/contracts";
import { useCreatorRegistration } from "../model/use-creator-registration";
import { CreatorRegistrationStaticWorkspacePreview } from "./creator-registration-static-workspace-preview";
import {
  CreatorRegistrationMessage,
  CreatorRegistrationSectionHeading,
  creatorRegistrationButtonClassName,
  creatorRegistrationFieldLabelClassName,
  creatorRegistrationInlineSurfaceClassName,
  creatorRegistrationInputClassName,
  creatorRegistrationSectionClassName,
  creatorRegistrationShellClassName,
  creatorRegistrationTextareaClassName,
} from "./creator-registration-ui-primitives";

const evidenceFieldLabels = {
  government_id: {
    description: "顔写真付きの確認書類です。画像またはPDFで提出できます。10MBまでです。",
    label: "本人確認書類",
  },
  payout_proof: {
    description: "売上を受け取る名義が分かる書類です。画像またはPDFで提出できます。10MBまでです。",
    label: "受取名義の確認書類",
  },
} as const;

const onboardingChecklist = [
  "本人確認に必要な書類をそろえる",
  "売上の受取名義が分かる書類を用意する",
  "禁止された内容を扱わないことを確認する",
  "確認が終わるまでは投稿や管理画面は使えません",
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

function formatEvidenceFileKind(mimeType: string) {
  if (mimeType === "application/pdf") {
    return "PDF";
  }
  if (mimeType.startsWith("image/")) {
    return "画像";
  }

  return "ファイル";
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
          "差し戻し理由を確認し、必要な項目だけ直してからもう一度申請できます。",
        saveLabel: "修正内容を保存する",
        submitLabel: "再申請する",
        title: "修正して再申請する",
      };
    case "rejected_support":
      return {
        description:
          "現在の内容はこの画面からは再申請できません。必要な案内が出るまでお待ちください。",
        saveLabel: "下書きを保存する",
        submitLabel: "再申請する",
        title: "サポート確認が必要です",
      };
    case "rejected_closed":
      return {
        description:
          "この申請はこの画面からは出し直せません。表示されている理由をご確認ください。",
        saveLabel: "下書きを保存する",
        submitLabel: "再申請する",
        title: "再申請は利用できません",
      };
    case "rejected_unknown":
      return {
        description:
          "申請状況を確認できませんでした。時間を置いてからもう一度開いてください。",
        saveLabel: "下書きを保存する",
        submitLabel: "再申請する",
        title: "審査状態を再確認してください",
      };
    case "suspended":
      return {
        description:
          "利用停止中のため、状態が更新されるまで内容確認のみ行えます。",
        saveLabel: "下書きを保存する",
        submitLabel: "申請を送る",
        title: "クリエイター利用は停止中です",
      };
    case "draft":
      return {
        description:
          "プロフィールを確認し、必要な情報と書類をそろえて申請します。",
        saveLabel: "下書きを保存する",
        submitLabel: "申請を送る",
        title: "クリエイター登録を始める",
      };
    case "submitted":
      return {
        description: "確認が終わるまでしばらくお待ちください。",
        saveLabel: "下書きを保存する",
        submitLabel: "申請を送る",
        title: "申請を受け付けました",
      };
    case "not_started":
    default:
      return {
        description:
          "プロフィールを確認し、必要な情報と書類をそろえて申請します。",
        saveLabel: "下書きを保存する",
        submitLabel: "申請を送る",
        title: "クリエイター登録を始める",
      };
  }
}

function resolveReasonChecklist(reasonCode: string | null): readonly string[] {
  switch (reasonCode) {
    case "documents_incomplete":
      return [
        "不足している書類をアップロードし直す",
        "氏名、生年月日、受取名義の空欄を埋める",
        "保存後にこの画面から再申請する",
      ];
    case "documents_blurry":
      return [
        "判別できる解像度で本人確認書類を再提出する",
        "見切れや反射がないファイルへ差し替える",
        "更新後にもう一度申請する",
      ];
    case "payout_info_incomplete":
      return [
        "受取名義の種類と受取名を見直す",
        "受取名義の確認書類を最新のものに差し替える",
        "修正後に再申請する",
      ];
    case "profile_mismatch":
      return [
        "プロフィールの表示名と提出情報の不一致を直す",
        "必要ならプロフィール設定で表示名やユーザー名を更新する",
        "反映後に再申請する",
      ];
    case "impersonation_suspected":
      return [
        "現在はこの画面から再申請できません",
        "運営側の確認が終わるまでお待ちください",
      ];
    default:
      return [
        "表示されている案内を確認する",
        "必要な項目が分かる場合だけ修正し、もう一度申請する",
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
        "詳しい差し戻し理由は読み込めませんでしたが、再申請はできます。気になる項目を見直してからもう一度送ってください。",
      checklist: [
        "差し戻しになった入力と証跡を見直す",
        "保存後にこの画面から再申請する",
      ],
      ctaLabel: "修正して再申請する",
      eyebrow: "再申請できます",
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
      rejectedAt ? `却下日時: ${rejectedAt}` : null,
      showRemaining
        ? `再申請できる残り回数: ${remaining}`
        : null,
    ].filter((value): value is string => value !== null);

    if (registration.actions.canResubmit) {
      return {
        body:
          "差し戻しになった項目だけ直せば、もう一度申請できます。",
        checklist: resolveReasonChecklist(reasonCode),
        ctaLabel: "修正して再申請する",
        eyebrow: "再申請できます",
        meta,
        title: "再申請できます",
      };
    }

    return {
      body:
        registration.rejection?.isSupportReviewRequired
          ? "いまはこの画面から再申請できません。運営からの確認をお待ちください。"
          : "この申請はこの画面からは出し直せません。必要な案内がある場合のみ従ってください。",
      checklist: resolveReasonChecklist(reasonCode),
      ctaLabel: null,
      eyebrow: registration.rejection?.isSupportReviewRequired ? "運営確認が必要です" : "再申請は利用できません",
      meta: registration.rejection?.isSupportReviewRequired
        ? [...meta, "次の案内: 運営確認が必要です"]
        : meta,
      title: registration.rejection?.isSupportReviewRequired ? "運営確認が必要です" : "再申請は利用できません",
    };
  }

  if (registration.state === "suspended") {
    const suspendedAt = formatStatusTimestamp(registration.review.suspendedAt);

    return {
      body:
        "利用停止中のため、いまは内容確認のみ行えます。",
      checklist: [
        "投稿や管理画面は再開まで使えません",
        "この画面では内容確認のみ行えます",
      ],
      ctaLabel: null,
      eyebrow: "利用停止中",
      meta: suspendedAt ? [`停止日時: ${suspendedAt}`] : [],
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
    <main className="mx-auto flex min-h-svh w-full max-w-[408px] px-4 py-10">
      <div className={creatorRegistrationShellClassName}>
        <div className="mt-2">
          <h1 className="font-display text-[32px] font-semibold leading-[1.12] tracking-[-0.05em] text-foreground">
            {panelCopy.title}
          </h1>
          <p className="mt-3 max-w-[34ch] text-sm leading-6 text-muted">{panelCopy.description}</p>
        </div>

        {isLoading ? (
          <CreatorRegistrationMessage className="mt-6" kind="info" message="入力内容を読み込んでいます。" />
        ) : null}

        {statusCard ? (
          <section className={`mt-6 ${creatorRegistrationSectionClassName}`}>
            <CreatorRegistrationSectionHeading>
              {statusCard.eyebrow}
            </CreatorRegistrationSectionHeading>
            <h2 className="mt-3 font-display text-[24px] font-semibold leading-[1.12] tracking-[-0.04em] text-foreground">
              {statusCard.title}
            </h2>
            <p className="mt-2 text-sm leading-6 text-muted">{statusCard.body}</p>

            {statusCard.meta.length > 0 ? (
              <div className={`mt-4 ${creatorRegistrationInlineSurfaceClassName} text-sm leading-6 text-muted`}>
                {statusCard.meta.map((item) => (
                  <p key={item}>{item}</p>
                ))}
              </div>
            ) : null}

            <div className="mt-4 grid gap-3">
              {statusCard.checklist.map((item) => (
                <div
                  className={`${creatorRegistrationInlineSurfaceClassName} text-sm leading-6 text-foreground`}
                  key={item}
                >
                  {item}
                </div>
              ))}
            </div>

            {statusCard.ctaLabel ? (
              <div className="mt-4">
                <Button className={creatorRegistrationButtonClassName} disabled={isBusy || isReadOnly} onClick={scrollToForm} type="button">
                  {statusCard.ctaLabel}
                </Button>
              </div>
            ) : null}
          </section>
        ) : null}

        <section className={`mt-6 ${creatorRegistrationSectionClassName}`}>
          <CreatorRegistrationSectionHeading>
            準備しておくこと
          </CreatorRegistrationSectionHeading>
          <div className="mt-4 grid gap-3">
            {onboardingChecklist.map((item) => (
              <div
                className={`${creatorRegistrationInlineSurfaceClassName} text-sm leading-6 text-foreground`}
                key={item}
              >
                {item}
              </div>
            ))}
          </div>
        </section>

        <CreatorRegistrationStaticWorkspacePreview />

        {sharedProfile ? (
          <section className={`mt-6 ${creatorRegistrationSectionClassName}`}>
            <div className="flex items-center gap-4">
              <Avatar className="size-[72px] border border-[#dfe8ef] bg-white text-[18px] font-semibold text-[#2f6176] shadow-none">
                {sharedProfile.avatar ? <AvatarImage alt={`${sharedProfile.displayName} の画像`} src={sharedProfile.avatar.url} /> : null}
                <AvatarFallback className="bg-transparent text-inherit">
                  {buildAvatarFallback(sharedProfile.displayName)}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <CreatorRegistrationSectionHeading>
                  現在のプロフィール
                </CreatorRegistrationSectionHeading>
                <p className="mt-2 text-[17px] font-bold tracking-[-0.02em] text-foreground">
                  {sharedProfile.displayName}
                </p>
                <p className="mt-1 text-sm text-muted">{sharedProfile.handle}</p>
              </div>
            </div>
            <div className={`mt-4 ${creatorRegistrationInlineSurfaceClassName} text-sm leading-6 text-muted`}>
              表示名、ユーザー名、画像はプロフィール設定で変更できます。
            </div>
            <div className="mt-4">
              <Button asChild className={creatorRegistrationButtonClassName} disabled={isBusy} variant="secondary">
                <Link href="/fan/settings/profile">プロフィール設定を開く</Link>
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
          <section className={`grid gap-4 ${creatorRegistrationSectionClassName}`}>
            <label className="grid gap-1.5" htmlFor="creator-registration-bio">
              <span className={creatorRegistrationFieldLabelClassName}>
                紹介文
              </span>
              <textarea
                className={creatorRegistrationTextareaClassName}
                disabled={isBusy || isReadOnly}
                id="creator-registration-bio"
                onChange={(event) => setCreatorBio(event.target.value)}
                placeholder="投稿したい雰囲気や、見てもらいたい内容を入力してください。"
                value={creatorBio}
              />
            </label>

            <label className="grid gap-1.5" htmlFor="creator-registration-legal-name">
              <span className={creatorRegistrationFieldLabelClassName}>
                氏名
              </span>
              <input
                autoComplete="name"
                className={creatorRegistrationInputClassName}
                disabled={isBusy || isReadOnly}
                id="creator-registration-legal-name"
                onChange={(event) => setLegalName(event.target.value)}
                placeholder="Mina Rei"
                type="text"
                value={legalName}
              />
            </label>

            <label className="grid gap-1.5" htmlFor="creator-registration-birth-date">
              <span className={creatorRegistrationFieldLabelClassName}>
                生年月日
              </span>
              <input
                className={creatorRegistrationInputClassName}
                disabled={isBusy || isReadOnly}
                id="creator-registration-birth-date"
                onChange={(event) => setBirthDate(event.target.value)}
                type="date"
                value={birthDate}
              />
            </label>

            <fieldset className="grid gap-2">
              <legend className={creatorRegistrationFieldLabelClassName}>
                受取名義の種類
              </legend>
              <div className="grid gap-2 sm:grid-cols-2">
                {[
                  { label: "自分名義", value: "self" },
                  { label: "事業名義", value: "business" },
                ].map((option) => (
                  <label
                    className={`${creatorRegistrationInlineSurfaceClassName} flex items-start gap-3 text-sm leading-6 text-foreground`}
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
              <span className={creatorRegistrationFieldLabelClassName}>
                受取名
              </span>
              <input
                className={creatorRegistrationInputClassName}
                disabled={isBusy || isReadOnly}
                id="creator-registration-payout-name"
                onChange={(event) => setPayoutRecipientName(event.target.value)}
                placeholder="Mina Rei"
                type="text"
                value={payoutRecipientName}
              />
            </label>
          </section>

          <section className={`grid gap-3 ${creatorRegistrationSectionClassName}`}>
            <CreatorRegistrationSectionHeading>
              提出書類
            </CreatorRegistrationSectionHeading>
            {creatorRegistrationEvidenceKinds.map((kind) => {
              const field = evidences[kind];
              const config = evidenceFieldLabels[kind];
              const evidenceUploadDisabled = isBusy || isReadOnly || field.isUploading;

              return (
                <section
                  className={`${creatorRegistrationInlineSurfaceClassName} text-foreground`}
                  key={kind}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <p className="text-[15px] font-bold tracking-[-0.02em] text-foreground">{config.label}</p>
                      <p className="mt-1 text-sm leading-6 text-muted">{config.description}</p>
                    </div>
                    <span className="text-[11px] font-black tracking-[0.08em] text-[#a3adbc]">
                      必須
                    </span>
                  </div>

                  <div className={`mt-4 ${creatorRegistrationSectionClassName} text-sm leading-6 text-muted`}>
                    {field.evidence ? (
                      <>
                        <p className="font-bold text-foreground">{field.evidence.fileName}</p>
                        <p className="mt-1">
                          {formatEvidenceFileKind(field.evidence.mimeType)} / {Math.ceil(field.evidence.fileSizeBytes / 1024)}KB / {formatEvidenceDate(field.evidence.uploadedAt)}
                        </p>
                      </>
                    ) : (
                      <p>まだアップロードされていません。</p>
                    )}
                  </div>

                  {field.errorMessage ? (
                    <CreatorRegistrationMessage className="mt-3" kind="error" message={field.errorMessage} />
                  ) : null}

                  <div className="mt-4">
                    <Button
                      className={creatorRegistrationButtonClassName}
                      disabled={evidenceUploadDisabled}
                      onClick={() => {
                        evidenceInputRefs.current[kind]?.click();
                      }}
                      type="button"
                      variant="secondary"
                    >
                      {field.isUploading ? "アップロード中..." : field.evidence ? "書類を差し替える" : "書類を選択する"}
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

          <section className={`grid gap-3 ${creatorRegistrationSectionClassName}`}>
            <label className={`${creatorRegistrationInlineSurfaceClassName} flex items-start gap-3`}>
              <input
                checked={declaresNoProhibitedCategory}
                className="mt-1 size-5 rounded-[6px] border border-[#cfd7e3] accent-accent-strong"
                disabled={isBusy || isReadOnly}
                onChange={(event) => setDeclaresNoProhibitedCategory(event.target.checked)}
                type="checkbox"
              />
              <span className="text-sm leading-6 text-foreground">
                禁止されている内容を扱わないことを確認しました。
              </span>
            </label>

            <label className={`${creatorRegistrationInlineSurfaceClassName} flex items-start gap-3`}>
              <input
                checked={acceptsConsentResponsibility}
                className="mt-1 size-5 rounded-[6px] border border-[#cfd7e3] accent-accent-strong"
                disabled={isBusy || isReadOnly}
                onChange={(event) => setAcceptsConsentResponsibility(event.target.checked)}
                type="checkbox"
              />
              <span className="text-sm leading-6 text-foreground">
                出演者の同意と権利確認の責任を自分で負うことを確認しました。
              </span>
            </label>
          </section>

          {successMessage ? (
            <CreatorRegistrationMessage kind="success" message={successMessage} />
          ) : null}

          {errorMessage ? (
            <CreatorRegistrationMessage kind="error" message={errorMessage} />
          ) : null}

          {showFormActions ? (
            <div className="grid gap-3">
              <Button
                className={creatorRegistrationButtonClassName}
                disabled={isBusy || isReadOnly}
                onClick={() => {
                  void saveDraft();
                }}
                type="button"
                variant="secondary"
              >
                {isSaving ? "保存中..." : panelCopy.saveLabel}
              </Button>

              <Button className={creatorRegistrationButtonClassName} disabled={submitDisabled} type="submit">
                {isSubmitting ? "送信中..." : panelCopy.submitLabel}
              </Button>
            </div>
          ) : null}
        </form>

        <div className="mt-6">
          <Button asChild className={creatorRegistrationButtonClassName} disabled={isBusy} variant="secondary">
            <Link href="/fan">あとでホームに戻る</Link>
          </Button>
        </div>
      </div>
    </main>
  );
}
