"use client";

import Link from "next/link";
import { AlertCircle, CheckSquare, ChevronLeft, FileBadge } from "lucide-react";
import { useRef } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui";

import {
  creatorRegistrationEvidenceKinds,
  type CreatorRegistrationEvidence,
  type CreatorRegistrationEvidenceKind,
  type CreatorRegistrationStatus,
} from "../api/contracts";
import { useCreatorRegistration } from "../model/use-creator-registration";
import { CreatorRegistrationStaticWorkspacePreview } from "./creator-registration-static-workspace-preview";
import {
  buildCreatorRegistrationAvatarFallback,
  CreatorRegistrationMessage,
  creatorRegistrationFocusRingClassName,
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

type RegistrationEvidenceFieldState = {
  errorMessage: string | null;
  evidence: CreatorRegistrationEvidence | null;
  inputKey: number;
  isUploading: boolean;
};

type ResubmitIssueSummary = {
  description: string;
  needsAttentionKind: CreatorRegistrationEvidenceKind | null;
  title: string;
};

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
        title: "運営確認が必要です",
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
        title: "停止中のため再申請できません",
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

function resolveResubmitIssueSummary(
  reasonCode: string | null,
  evidences: Record<CreatorRegistrationEvidenceKind, RegistrationEvidenceFieldState>,
): ResubmitIssueSummary {
  const missingKinds = creatorRegistrationEvidenceKinds.filter((kind) => evidences[kind].evidence === null);

  if (reasonCode === "payout_info_incomplete") {
    return {
      description:
        "受取名義の確認書類または入力内容に不足があります。内容を見直して再度申請してください。",
      needsAttentionKind: "payout_proof",
      title: "申請が差し戻されました",
    };
  }

  if (reasonCode === "documents_blurry") {
    return {
      description:
        "書類が不鮮明で読み取れません。新しい書類に差し替えて再度申請してください。",
      needsAttentionKind: missingKinds[0] ?? null,
      title: "申請が差し戻されました",
    };
  }

  if (reasonCode === "documents_incomplete") {
    return {
      description:
        "必要な書類または入力内容に不足があります。内容を見直して再度申請してください。",
      needsAttentionKind: missingKinds[0] ?? null,
      title: "申請が差し戻されました",
    };
  }

  if (reasonCode === "profile_mismatch") {
    return {
      description:
        "プロフィール情報と提出内容が一致していません。内容を見直して再度申請してください。",
      needsAttentionKind: missingKinds[0] ?? null,
      title: "申請が差し戻されました",
    };
  }

  return {
    description:
      "表示された内容を見直し、必要な修正をしてから再度申請してください。",
    needsAttentionKind: missingKinds[0] ?? null,
    title: "申請が差し戻されました",
  };
}

function resolveResubmitEvidenceActionLabel({
  hasEvidence,
  isAttention,
  isUploading,
}: {
  hasEvidence: boolean;
  isAttention: boolean;
  isUploading: boolean;
}) {
  if (isUploading) {
    return "アップロード中...";
  }

  if (isAttention) {
    return "新しい書類をアップロード";
  }

  if (hasEvidence) {
    return "書類を差し替える";
  }

  return "書類をアップロード";
}

function resolveRegistrationPageTitle(surfaceKind: RegistrationSurfaceKind) {
  switch (surfaceKind) {
    case "draft":
    case "not_started":
      return "クリエイター登録";
    case "rejected_closed":
    case "rejected_support":
    case "rejected_unknown":
    case "suspended":
      return "申請状況";
    case "submitted":
      return "申請完了";
    case "rejected_resubmit":
      return "再申請";
    default:
      return "クリエイター登録";
  }
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
  const showStatusEyebrow = statusCard !== null && statusCard.eyebrow !== statusCard.title;
  const pageTitle = resolveRegistrationPageTitle(surfaceKind);
  const showIntroCard = surfaceKind === "draft" || surfaceKind === "not_started";
  const profilePreview = sharedProfile ?? registration?.sharedProfile ?? null;
  const resubmitIssueSummary = resolveResubmitIssueSummary(
    registration?.rejection?.reasonCode ?? null,
    evidences,
  );
  const resubmitRemaining = registration?.rejection?.selfServeResubmitRemaining ?? null;

  if (surfaceKind === "rejected_resubmit") {
    return (
      <main className="relative mx-auto flex min-h-svh w-full max-w-[408px] flex-col bg-[#f7f8fb] text-foreground">
        <div className="sticky top-0 z-20 flex items-center justify-between border-b border-[#eef1f5] bg-white/90 px-4 pb-4 pt-14 shadow-sm backdrop-blur-md">
          <Link aria-label="戻る" className="text-gray-800" href="/fan">
            <ChevronLeft className="size-7" />
          </Link>
          <span className="text-[17px] font-bold text-foreground">再申請</span>
          <div className="w-7" />
        </div>

        <div className="flex-1 overflow-y-auto px-5 pb-36 pt-5">
          <div className="space-y-6">
            <section className="rounded-[28px] border border-[#f7d9de] bg-[#fff4f6] p-5 shadow-sm">
              <div className="flex items-start gap-3">
                <div className="mt-0.5 rounded-full bg-[#fee4e8] p-2 text-[#ef476f]">
                  <AlertCircle className="size-5" strokeWidth={2.2} />
                </div>
                <div>
                  <h1 className="text-[18px] font-extrabold leading-tight text-[#8f2239]">
                    {resubmitIssueSummary.title}
                  </h1>
                  <p className="mt-1 text-[13px] font-medium leading-relaxed text-[#b24358]">
                    {resubmitIssueSummary.description}
                  </p>
                  {resubmitRemaining !== null ? (
                    <p className="mt-2 text-[12px] font-bold text-[#c76174]">
                      （残り申請回数：{resubmitRemaining}回）
                    </p>
                  ) : null}
                </div>
              </div>
            </section>

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
                      {buildCreatorRegistrationAvatarFallback(profilePreview.displayName)}
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

            <form
              className="space-y-6"
              id="creator-registration-resubmit-form"
              onSubmit={(event) => {
                event.preventDefault();
                void submit();
              }}
              ref={formRef}
            >
              <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
                <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
                  申請情報
                </p>

                <div className="mt-4 space-y-5">
                  <div>
                    <label
                      className="ml-1 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]"
                      htmlFor="creator-registration-bio"
                    >
                      紹介文
                    </label>
                    <textarea
                      className="mt-2 h-24 w-full resize-none rounded-[20px] border-2 border-transparent bg-[#f6f7fb] px-5 py-4 text-[15px] font-bold text-foreground outline-none transition focus:border-[#dcebff] focus:bg-white"
                      disabled={isBusy || isReadOnly}
                      id="creator-registration-bio"
                      onChange={(event) => setCreatorBio(event.target.value)}
                      placeholder="紹介したい内容を入力してください。"
                      value={creatorBio}
                    />
                  </div>

                  <div>
                    <label
                      className="ml-1 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]"
                      htmlFor="creator-registration-legal-name"
                    >
                      氏名
                    </label>
                    <input
                      autoComplete="name"
                      className="mt-2 w-full rounded-[20px] border-2 border-transparent bg-[#f6f7fb] px-5 py-4 text-[15px] font-bold text-foreground outline-none transition focus:border-[#dcebff] focus:bg-white"
                      disabled={isBusy || isReadOnly}
                      id="creator-registration-legal-name"
                      onChange={(event) => setLegalName(event.target.value)}
                      placeholder="阿部壮一郎"
                      type="text"
                      value={legalName}
                    />
                  </div>

                  <div>
                    <label
                      className="ml-1 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]"
                      htmlFor="creator-registration-birth-date"
                    >
                      生年月日
                    </label>
                    <input
                      className="mt-2 w-full rounded-[20px] border-2 border-transparent bg-[#f6f7fb] px-5 py-4 text-[15px] font-bold text-foreground outline-none transition focus:border-[#dcebff] focus:bg-white"
                      disabled={isBusy || isReadOnly}
                      id="creator-registration-birth-date"
                      onChange={(event) => setBirthDate(event.target.value)}
                      type="date"
                      value={birthDate}
                    />
                  </div>

                  <div>
                    <label className="ml-1 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]">
                      受取名義の種類
                    </label>
                    <div className="mt-3 grid gap-3 sm:grid-cols-2">
                      {[
                        { label: "自分名義", value: "self" },
                        { label: "事業名義", value: "business" },
                      ].map((option) => {
                        const isChecked = payoutRecipientType === option.value;

                        return (
                          <label
                            className={`flex cursor-pointer items-center justify-center rounded-[20px] border-2 px-4 py-3.5 transition-colors ${
                              isChecked
                                ? "border-[#dcebff] bg-[#eef6ff] text-[#134b80]"
                                : "border-transparent bg-[#f6f7fb] text-foreground hover:bg-[#eef2f7]"
                            }`}
                            htmlFor={`creator-registration-payout-${option.value}`}
                            key={option.value}
                          >
                            <input
                              checked={isChecked}
                              className="size-4 border-gray-300"
                              disabled={isBusy || isReadOnly}
                              id={`creator-registration-payout-${option.value}`}
                              name="creator-registration-payout-type"
                              onChange={() => setPayoutRecipientType(option.value)}
                              style={{ accentColor: "#4DA8DA" }}
                              type="radio"
                            />
                            <span className="ml-2 text-[14px] font-bold">
                              {option.label}
                            </span>
                          </label>
                        );
                      })}
                    </div>

                    <label
                      className="ml-1 mt-4 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]"
                      htmlFor="creator-registration-payout-name"
                    >
                      受取名
                    </label>
                    <input
                      className="mt-2 w-full rounded-[20px] border-2 border-transparent bg-[#f6f7fb] px-5 py-4 text-[15px] font-bold text-foreground outline-none transition focus:border-[#dcebff] focus:bg-white"
                      disabled={isBusy || isReadOnly}
                      id="creator-registration-payout-name"
                      onChange={(event) => setPayoutRecipientName(event.target.value)}
                      placeholder="阿部壮一郎"
                      type="text"
                      value={payoutRecipientName}
                    />
                  </div>
                </div>
              </section>

              <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
                <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
                  提出書類
                </p>

                <div className="mt-4 space-y-5">
                  {creatorRegistrationEvidenceKinds.map((kind) => {
                    const field = evidences[kind];
                    const config = evidenceFieldLabels[kind];
                    const isAttention = resubmitIssueSummary.needsAttentionKind === kind;
                    const evidenceUploadDisabled = isBusy || isReadOnly || field.isUploading;

                    return (
                      <section
                        className={`relative overflow-hidden rounded-[22px] p-4 ${
                          isAttention
                            ? "border border-[#f3cfd6] bg-[#fff4f6] shadow-sm"
                            : "border border-gray-100 bg-[#f8f9fc]"
                        }`}
                        key={kind}
                      >
                        {isAttention ? (
                          <div className="absolute inset-y-0 left-0 w-1 bg-[#ef476f]" />
                        ) : null}

                        <div className={`flex items-start gap-3 ${isAttention ? "ml-2" : ""}`}>
                          <div
                            className={`rounded-full p-2 ${
                              isAttention
                                ? "bg-[#fee4e8] text-[#ef476f]"
                                : "bg-[#e7f7ef] text-[#1f9d61]"
                            }`}
                          >
                            {isAttention ? (
                              <FileBadge className="size-4" strokeWidth={2.2} />
                            ) : (
                              <CheckSquare className="size-4" strokeWidth={2.2} />
                            )}
                          </div>
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2">
                              <p
                                className={`text-[13px] font-bold ${
                                  isAttention ? "text-[#8f2239]" : "text-foreground"
                                }`}
                              >
                                {config.label}
                              </p>
                              {isAttention ? (
                                <span className="rounded bg-[#fee4e8] px-1.5 py-0.5 text-[10px] font-black text-[#c03556]">
                                  要修正
                                </span>
                              ) : null}
                            </div>
                            <p
                              className={`mt-1 text-[11px] font-medium ${
                                isAttention ? "text-[#c35d71]" : "text-muted"
                              }`}
                            >
                              {field.evidence?.fileName ?? "未提出"}
                            </p>
                          </div>
                        </div>

                        <button
                          className={`mt-3 w-full rounded-[14px] border py-3 text-[14px] font-bold transition-colors ${
                            isAttention
                              ? "border-[#e34d6f] bg-[#ef476f] text-white hover:bg-[#df3d61]"
                              : "border-gray-200 bg-white text-gray-700 hover:bg-gray-50"
                          } ${isAttention ? "ml-1 w-[calc(100%-4px)]" : ""} disabled:cursor-not-allowed disabled:opacity-60`}
                          disabled={evidenceUploadDisabled}
                          onClick={() => {
                            evidenceInputRefs.current[kind]?.click();
                          }}
                          type="button"
                        >
                          {resolveResubmitEvidenceActionLabel({
                            hasEvidence: field.evidence !== null,
                            isAttention,
                            isUploading: field.isUploading,
                          })}
                        </button>
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

                        {field.errorMessage ? (
                          <CreatorRegistrationMessage
                            className="mt-3"
                            kind="error"
                            message={field.errorMessage}
                          />
                        ) : null}
                      </section>
                    );
                  })}
                </div>
              </section>

              <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
                <div className="space-y-3">
                  <label className="flex cursor-pointer items-start gap-3 p-1">
                    <input
                      checked={declaresNoProhibitedCategory}
                      className="mt-0.5 size-5 rounded border-gray-300 bg-[#f6f7fb]"
                      disabled={isBusy || isReadOnly}
                      onChange={(event) => setDeclaresNoProhibitedCategory(event.target.checked)}
                      style={{ accentColor: "#4DA8DA" }}
                      type="checkbox"
                    />
                    <span className="text-[13px] font-medium leading-snug text-gray-700">
                      禁止されている内容を扱わないことを確認しました。
                    </span>
                  </label>

                  <div className="h-px w-full bg-gray-100" />

                  <label className="flex cursor-pointer items-start gap-3 p-1">
                    <input
                      checked={acceptsConsentResponsibility}
                      className="mt-0.5 size-5 rounded border-gray-300 bg-[#f6f7fb]"
                      disabled={isBusy || isReadOnly}
                      onChange={(event) => setAcceptsConsentResponsibility(event.target.checked)}
                      style={{ accentColor: "#4DA8DA" }}
                      type="checkbox"
                    />
                    <span className="text-[13px] font-medium leading-snug text-gray-700">
                      出演者の同意と権利確認の責任を自分で負うことを確認しました。
                    </span>
                  </label>
                </div>
              </section>
            </form>
          </div>
        </div>

        <div className="absolute bottom-0 left-0 z-30 w-full border-t border-gray-100 bg-white/95 px-4 pb-8 pt-4 shadow-[0_-10px_20px_rgba(0,0,0,0.03)] backdrop-blur-md">
          {successMessage ? (
            <CreatorRegistrationMessage className="mb-3" kind="success" message={successMessage} />
          ) : null}
        {errorMessage ? (
          <CreatorRegistrationMessage className="mb-3" kind="error" message={errorMessage} />
        ) : null}
        <button
          className={`w-full rounded-full border border-gray-200 bg-white py-4 text-[16px] font-bold text-gray-900 shadow-sm transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60 ${creatorRegistrationFocusRingClassName}`}
          disabled={isBusy || isReadOnly}
          onClick={() => {
            void saveDraft();
          }}
          type="button"
        >
          {isSaving ? "保存中..." : "修正内容を保存する"}
        </button>
        <button
          className={`mt-3 w-full rounded-full bg-[#4DA8DA] py-4 text-[16px] font-bold text-white shadow-lg shadow-[#4DA8DA]/20 transition-transform active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60 ${creatorRegistrationFocusRingClassName}`}
          disabled={submitDisabled}
          form="creator-registration-resubmit-form"
          type="submit"
        >
            {isSubmitting ? "送信中..." : "再申請する"}
          </button>
          <Link
            className={`mt-2 block w-full rounded-full py-3 text-center text-[14px] font-bold text-gray-500 transition-colors hover:text-gray-800 ${creatorRegistrationFocusRingClassName}`}
            href="/fan"
          >
            閉じる
          </Link>
        </div>
      </main>
    );
  }

  return (
    <main className="relative mx-auto flex min-h-svh w-full max-w-[408px] flex-col bg-[#f7f8fb] text-foreground">
      <div className="sticky top-0 z-20 flex items-center justify-between border-b border-[#eef1f5] bg-white/90 px-4 pb-4 pt-14 shadow-sm backdrop-blur-md">
        <Link aria-label="戻る" className="text-gray-800" href="/fan">
          <ChevronLeft className="size-7" />
        </Link>
        <span className="text-[17px] font-bold text-foreground">{pageTitle}</span>
        <div className="w-7" />
      </div>

      <div className="flex-1 overflow-y-auto px-5 pb-44 pt-5">
        <div className="space-y-6">
          {statusCard ? (
            <section
              className={`rounded-[28px] border p-5 shadow-sm ${
                surfaceKind === "suspended"
                  ? "border-[#e2e8f0] bg-[#f4f6f8]"
                  : "border-[#f6dfc6] bg-[#fff7f1]"
              }`}
            >
              <div className="flex items-start gap-3">
                <div
                  className={`mt-0.5 rounded-full p-2 ${
                    surfaceKind === "suspended"
                      ? "bg-white text-[#6b7280]"
                      : "bg-[#ffe7cf] text-[#d97706]"
                  }`}
                >
                  <AlertCircle className="size-5" strokeWidth={2.2} />
                </div>
                <div>
                  {showStatusEyebrow ? (
                    <p
                      className={`text-[12px] font-black tracking-[0.15em] ${
                        surfaceKind === "suspended" ? "text-[#6b7280]" : "text-[#c98c44]"
                      }`}
                    >
                      {statusCard.eyebrow}
                    </p>
                  ) : null}
                  <h1
                    className={`font-extrabold leading-tight ${
                      showStatusEyebrow ? "mt-1 text-[18px]" : "text-[18px]"
                    } ${surfaceKind === "suspended" ? "text-[#374151]" : "text-[#8a5317]"}`}
                  >
                    {statusCard.title}
                  </h1>
                  <p
                    className={`mt-1 text-[13px] font-medium leading-relaxed ${
                      surfaceKind === "suspended" ? "text-[#6b7280]" : "text-[#b16e2c]"
                    }`}
                  >
                    {statusCard.body}
                  </p>
                </div>
              </div>

              {statusCard.meta.length > 0 ? (
                <div className="mt-4 space-y-2">
                  {statusCard.meta.map((item) => (
                    <div
                      className="rounded-[18px] border border-white/70 bg-white/80 px-4 py-3 text-[13px] font-medium leading-relaxed text-gray-600"
                      key={item}
                    >
                      {item}
                    </div>
                  ))}
                </div>
              ) : null}

              <div className="mt-4 space-y-2">
                {statusCard.checklist.map((item) => (
                  <div
                    className="rounded-[18px] border border-white/70 bg-white/80 px-4 py-3 text-[13px] font-medium leading-relaxed text-foreground"
                    key={item}
                  >
                    {item}
                  </div>
                ))}
              </div>
            </section>
          ) : (
            <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
              <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
                {showIntroCard ? "はじめに" : pageTitle}
              </p>
              <h1 className="mt-3 text-[22px] font-extrabold leading-tight text-foreground">
                {panelCopy.title}
              </h1>
              <p className="mt-1 text-[13px] font-medium leading-relaxed text-muted">
                {panelCopy.description}
              </p>
            </section>
          )}

          {isLoading ? (
            <CreatorRegistrationMessage
              kind="info"
              message="入力内容を読み込んでいます。"
            />
          ) : null}

          {showIntroCard ? (
            <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
              <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
                申請前に確認すること
              </p>
              <div className="mt-4 space-y-3">
                {onboardingChecklist.map((item) => (
                  <div
                    className="rounded-[22px] border border-gray-100 bg-[#f8f9fc] px-4 py-4 text-[13px] font-medium leading-relaxed text-foreground"
                    key={item}
                  >
                    {item}
                  </div>
                ))}
              </div>
            </section>
          ) : null}

          {showIntroCard ? <CreatorRegistrationStaticWorkspacePreview /> : null}

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
                      {buildCreatorRegistrationAvatarFallback(profilePreview.displayName)}
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

              <p className="mt-4 text-[13px] font-medium leading-relaxed text-muted">
                表示名、ユーザー名、画像はプロフィール設定で変更できます。
              </p>
            </section>
          ) : null}

          <form
            className="space-y-6"
            id="creator-registration-form"
            onSubmit={(event) => {
              event.preventDefault();
              void submit();
            }}
            ref={formRef}
          >
            <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
              <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
                申請情報
              </p>

              <div className="mt-4 space-y-5">
                <div>
                  <label
                    className="ml-1 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]"
                    htmlFor="creator-registration-bio"
                  >
                    紹介文
                  </label>
                  <textarea
                    className="mt-2 h-24 w-full resize-none rounded-[20px] border-2 border-transparent bg-[#f6f7fb] px-5 py-4 text-[15px] font-bold text-foreground outline-none transition focus:border-[#dcebff] focus:bg-white"
                    disabled={isBusy || isReadOnly}
                    id="creator-registration-bio"
                    onChange={(event) => setCreatorBio(event.target.value)}
                    placeholder="紹介したい内容を入力してください。"
                    value={creatorBio}
                  />
                </div>

                <div>
                  <label
                    className="ml-1 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]"
                    htmlFor="creator-registration-legal-name"
                  >
                    氏名
                  </label>
                  <input
                    autoComplete="name"
                    className="mt-2 w-full rounded-[20px] border-2 border-transparent bg-[#f6f7fb] px-5 py-4 text-[15px] font-bold text-foreground outline-none transition focus:border-[#dcebff] focus:bg-white"
                    disabled={isBusy || isReadOnly}
                    id="creator-registration-legal-name"
                    onChange={(event) => setLegalName(event.target.value)}
                    placeholder="阿部壮一郎"
                    type="text"
                    value={legalName}
                  />
                </div>

                <div>
                  <label
                    className="ml-1 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]"
                    htmlFor="creator-registration-birth-date"
                  >
                    生年月日
                  </label>
                  <input
                    className="mt-2 w-full rounded-[20px] border-2 border-transparent bg-[#f6f7fb] px-5 py-4 text-[15px] font-bold text-foreground outline-none transition focus:border-[#dcebff] focus:bg-white"
                    disabled={isBusy || isReadOnly}
                    id="creator-registration-birth-date"
                    onChange={(event) => setBirthDate(event.target.value)}
                    type="date"
                    value={birthDate}
                  />
                </div>

                <div>
                  <label className="ml-1 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]">
                    受取名義の種類
                  </label>
                  <div className="mt-3 grid gap-3 sm:grid-cols-2">
                    {[
                      { label: "自分名義", value: "self" },
                      { label: "事業名義", value: "business" },
                    ].map((option) => {
                      const isChecked = payoutRecipientType === option.value;

                      return (
                        <label
                          className={`flex cursor-pointer items-center justify-center rounded-[20px] border-2 px-4 py-3.5 transition-colors ${
                            isChecked
                              ? "border-[#dcebff] bg-[#eef6ff] text-[#134b80]"
                              : "border-transparent bg-[#f6f7fb] text-foreground hover:bg-[#eef2f7]"
                          }`}
                          htmlFor={`creator-registration-payout-${option.value}`}
                          key={option.value}
                        >
                          <input
                            checked={isChecked}
                            className="size-4 border-gray-300"
                            disabled={isBusy || isReadOnly}
                            id={`creator-registration-payout-${option.value}`}
                            name="creator-registration-payout-type"
                            onChange={() => setPayoutRecipientType(option.value)}
                            style={{ accentColor: "#4DA8DA" }}
                            type="radio"
                          />
                          <span className="ml-2 text-[14px] font-bold">
                            {option.label}
                          </span>
                        </label>
                      );
                    })}
                  </div>

                  <label
                    className="ml-1 mt-4 block text-[12px] font-black tracking-[0.08em] text-[#a3adbc]"
                    htmlFor="creator-registration-payout-name"
                  >
                    受取名
                  </label>
                  <input
                    className="mt-2 w-full rounded-[20px] border-2 border-transparent bg-[#f6f7fb] px-5 py-4 text-[15px] font-bold text-foreground outline-none transition focus:border-[#dcebff] focus:bg-white"
                    disabled={isBusy || isReadOnly}
                    id="creator-registration-payout-name"
                    onChange={(event) => setPayoutRecipientName(event.target.value)}
                    placeholder="阿部壮一郎"
                    type="text"
                    value={payoutRecipientName}
                  />
                </div>
              </div>
            </section>

            <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
              <p className="text-[12px] font-black tracking-[0.15em] text-[#a3adbc]">
                提出書類
              </p>

              <div className="mt-4 space-y-5">
                {creatorRegistrationEvidenceKinds.map((kind) => {
                  const field = evidences[kind];
                  const config = evidenceFieldLabels[kind];
                  const evidenceUploadDisabled = isBusy || isReadOnly || field.isUploading;
                  const hasEvidence = field.evidence !== null;

                  return (
                    <section
                      className="rounded-[22px] border border-gray-100 bg-[#f8f9fc] p-4"
                      key={kind}
                    >
                      <div className="flex items-start gap-3">
                        <div
                          className={`rounded-full p-2 ${
                            hasEvidence
                              ? "bg-[#e7f7ef] text-[#1f9d61]"
                              : "bg-[#eef6ff] text-[#4DA8DA]"
                          }`}
                        >
                          {hasEvidence ? (
                            <CheckSquare className="size-4" strokeWidth={2.2} />
                          ) : (
                            <FileBadge className="size-4" strokeWidth={2.2} />
                          )}
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center justify-between gap-3">
                            <p className="text-[13px] font-bold text-foreground">
                              {config.label}
                            </p>
                            <span className="text-[10px] font-black tracking-[0.08em] text-[#a3adbc]">
                              必須
                            </span>
                          </div>
                          <p className="mt-1 text-[12px] font-medium leading-relaxed text-muted">
                            {config.description}
                          </p>
                        </div>
                      </div>

                      <div className="mt-3 rounded-[16px] border border-gray-100 bg-white px-4 py-3 text-[12px] font-medium leading-relaxed text-muted">
                        {field.evidence ? (
                          <>
                            <p className="font-bold text-foreground">
                              {field.evidence.fileName}
                            </p>
                            <p className="mt-1">
                              {formatEvidenceFileKind(field.evidence.mimeType)} /{" "}
                              {Math.ceil(field.evidence.fileSizeBytes / 1024)}KB /{" "}
                              {formatEvidenceDate(field.evidence.uploadedAt)}
                            </p>
                          </>
                        ) : (
                          <p>まだ提出されていません。</p>
                        )}
                      </div>

                      <button
                        className="mt-3 w-full rounded-[14px] border border-gray-200 bg-white py-3 text-[14px] font-bold text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60"
                        disabled={evidenceUploadDisabled}
                        onClick={() => {
                          evidenceInputRefs.current[kind]?.click();
                        }}
                        type="button"
                      >
                        {field.isUploading
                          ? "アップロード中..."
                          : field.evidence
                            ? "書類を差し替える"
                            : "書類をアップロード"}
                      </button>
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

                      {field.errorMessage ? (
                        <CreatorRegistrationMessage
                          className="mt-3"
                          kind="error"
                          message={field.errorMessage}
                        />
                      ) : null}
                    </section>
                  );
                })}
              </div>
            </section>

            <section className="rounded-[28px] border border-gray-100 bg-white p-5 shadow-[0_2px_15px_rgba(0,0,0,0.03)]">
              <div className="space-y-3">
                <label className="flex cursor-pointer items-start gap-3 p-1">
                  <input
                    checked={declaresNoProhibitedCategory}
                    className="mt-0.5 size-5 rounded border-gray-300 bg-[#f6f7fb]"
                    disabled={isBusy || isReadOnly}
                    onChange={(event) => setDeclaresNoProhibitedCategory(event.target.checked)}
                    style={{ accentColor: "#4DA8DA" }}
                    type="checkbox"
                  />
                  <span className="text-[13px] font-medium leading-snug text-gray-700">
                    禁止されている内容を扱わないことを確認しました。
                  </span>
                </label>

                <div className="h-px w-full bg-gray-100" />

                <label className="flex cursor-pointer items-start gap-3 p-1">
                  <input
                    checked={acceptsConsentResponsibility}
                    className="mt-0.5 size-5 rounded border-gray-300 bg-[#f6f7fb]"
                    disabled={isBusy || isReadOnly}
                    onChange={(event) => setAcceptsConsentResponsibility(event.target.checked)}
                    style={{ accentColor: "#4DA8DA" }}
                    type="checkbox"
                  />
                  <span className="text-[13px] font-medium leading-snug text-gray-700">
                    出演者の同意と権利確認の責任を自分で負うことを確認しました。
                  </span>
                </label>
              </div>
            </section>
          </form>
        </div>
      </div>

      <div className="absolute bottom-0 left-0 z-30 w-full border-t border-gray-100 bg-white/95 px-4 pb-8 pt-4 shadow-[0_-10px_20px_rgba(0,0,0,0.03)] backdrop-blur-md">
        {successMessage ? (
          <CreatorRegistrationMessage className="mb-3" kind="success" message={successMessage} />
        ) : null}
        {errorMessage ? (
          <CreatorRegistrationMessage className="mb-3" kind="error" message={errorMessage} />
        ) : null}

        {showFormActions ? (
          <>
            <button
              className={`w-full rounded-full border border-gray-200 bg-white py-4 text-[16px] font-bold text-gray-900 shadow-sm transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60 ${creatorRegistrationFocusRingClassName}`}
              disabled={isBusy || isReadOnly}
              onClick={() => {
                void saveDraft();
              }}
              type="button"
            >
              {isSaving ? "保存中..." : panelCopy.saveLabel}
            </button>

            <button
              className={`mt-3 w-full rounded-full bg-[#4DA8DA] py-4 text-[16px] font-bold text-white shadow-lg shadow-[#4DA8DA]/20 transition-transform active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60 ${creatorRegistrationFocusRingClassName}`}
              disabled={submitDisabled}
              form="creator-registration-form"
              type="submit"
            >
              {isSubmitting ? "送信中..." : panelCopy.submitLabel}
            </button>
          </>
        ) : null}

        <Link
          className={`mt-2 block w-full rounded-full py-3 text-center text-[14px] font-bold text-gray-500 transition-colors hover:text-gray-800 ${creatorRegistrationFocusRingClassName}`}
          href="/fan"
        >
          閉じる
        </Link>
      </div>
    </main>
  );
}
