"use client";

import { startTransition, useState } from "react";
import { useRouter } from "next/navigation";

import { SurfacePanel } from "@/shared/ui";
import {
  updateCreatorWorkspaceProfile,
  useViewerProfileDraft,
  ViewerProfileEditorForm,
  getViewerProfileErrorCode,
  getViewerProfileSaveErrorMessage,
} from "@/features/viewer-profile";

import { useCreatorWorkspaceSummary } from "../model/use-creator-workspace-summary";
import { CreatorModeShell } from "./creator-mode-shell";

function CreatorProfileSettingsLoading() {
  return (
    <main className="mx-auto flex min-h-full w-full max-w-[408px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">creator settings</p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          プロフィールを編集
        </h1>
        <p className="sr-only" role="status">
          編集内容を読み込んでいます...
        </p>
        <div aria-hidden="true" className="mt-5 h-12 animate-pulse rounded-[18px] bg-[rgba(186,231,255,0.28)]" />
        <div aria-hidden="true" className="mt-3 h-12 animate-pulse rounded-[18px] bg-[rgba(186,231,255,0.24)]" />
        <div aria-hidden="true" className="mt-3 h-[132px] animate-pulse rounded-[22px] bg-[rgba(186,231,255,0.22)]" />
        <div aria-hidden="true" className="mt-3 h-28 animate-pulse rounded-[18px] bg-[rgba(186,231,255,0.2)]" />
      </SurfacePanel>
    </main>
  );
}

function CreatorProfileSettingsError({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <main className="mx-auto flex min-h-full w-full max-w-[408px] flex-col px-4 pb-28 pt-5">
      <SurfacePanel className="w-full px-5 py-6 text-foreground">
        <p className="text-[11px] font-bold uppercase tracking-[0.24em] text-accent">creator settings</p>
        <h1 className="mt-3 font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.04em]">
          プロフィールを編集
        </h1>
        <div
          className="mt-5 rounded-[22px] border border-[rgba(255,184,189,0.84)] bg-[linear-gradient(180deg,rgba(255,247,248,0.98),rgba(255,241,243,0.96))] px-4 py-4 text-foreground"
          role="alert"
        >
          <p className="m-0 text-[13px] leading-6 text-muted">{message}</p>
          <button
            className="mt-3 inline-flex min-h-10 items-center rounded-[12px] bg-[#1082c8] px-4 text-[13px] font-bold text-white transition hover:opacity-90"
            onClick={onRetry}
            type="button"
          >
            再読み込み
          </button>
        </div>
      </SurfacePanel>
    </main>
  );
}

function CreatorProfileBioField({
  disabled,
  onChange,
  value,
}: {
  disabled: boolean;
  onChange: (value: string) => void;
  value: string;
}) {
  return (
    <label className="block" htmlFor="creator-viewer-profile-bio">
      <span className="block pl-1 text-[12px] font-extrabold uppercase tracking-[0.12em] text-[#a9aeb9]">
        Bio
      </span>
      <textarea
        className="mt-2.5 min-h-[132px] w-full rounded-[20px] border border-transparent bg-[#f6f7fa] px-5 py-4 text-[15px] leading-6 text-[#1f2430] outline-none transition placeholder:text-[#b5bbc6] focus:border-[#d7e6f5] focus:bg-white focus:ring-4 focus:ring-[rgba(113,180,234,0.18)]"
        disabled={disabled}
        id="creator-viewer-profile-bio"
        onChange={(event) => onChange(event.target.value)}
        placeholder="quiet rooftop の continuation を中心に投稿します。"
        rows={4}
        value={value}
      />
    </label>
  );
}

function CreatorProfileSettingsForm({
  initialValues,
}: {
  initialValues: {
    avatarUrl: string | null;
    bio: string;
    displayName: string;
    handle: string;
  };
}) {
  const router = useRouter();
  const [bio, setBio] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const draft = useViewerProfileDraft({
    initialValues: {
      avatarUrl: initialValues.avatarUrl,
      displayName: initialValues.displayName,
      handle: initialValues.handle,
    },
    mode: "edit",
    onDirty: () => {
      if (errorMessage !== null) {
        setErrorMessage(null);
      }
    },
  });
  const resolvedBio = bio ?? initialValues.bio;

  const submit = async () => {
    if (isSubmitting) {
      return;
    }

    const profileValidationError = draft.getProfileValidationError();
    if (profileValidationError !== null) {
      setErrorMessage(profileValidationError);
      return;
    }

    const avatarSubmissionError = draft.getAvatarSubmissionError();
    if (avatarSubmissionError !== null) {
      setErrorMessage(avatarSubmissionError);
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      const avatarUploadToken = await draft.uploadAvatarIfNeeded();

      await updateCreatorWorkspaceProfile({
        ...(avatarUploadToken ? { avatarUploadToken } : {}),
        bio: resolvedBio,
        displayName: draft.displayName,
        handle: draft.handle,
      });

      startTransition(() => {
        router.replace("/creator");
      });
    } catch (error) {
      if (getViewerProfileErrorCode(error) === "invalid_avatar_upload_token") {
        draft.resetCompletedAvatarUploadToken();
      }
      setErrorMessage(getViewerProfileSaveErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <ViewerProfileEditorForm
      avatar={draft.avatar}
      avatarInputKey={draft.avatarInputKey}
      backHref="/creator"
      backLabel="ワークスペースへ戻る"
      displayName={draft.displayName}
      errorMessage={errorMessage}
      handle={draft.handle}
      isSubmitting={isSubmitting}
      onAvatarClear={draft.clearAvatarSelection}
      onAvatarSelect={draft.selectAvatarFile}
      onDisplayNameChange={draft.setDisplayName}
      onHandleChange={draft.setHandle}
      onSubmit={submit}
    >
      <CreatorProfileBioField
        disabled={isSubmitting}
        onChange={(nextBio) => {
          setBio(nextBio);
          if (errorMessage !== null) {
            setErrorMessage(null);
          }
        }}
        value={resolvedBio}
      />
    </ViewerProfileEditorForm>
  );
}

/**
 * creator workspace の現在 profile を prefill した編集フォームを表示する。
 */
export function CreatorProfileSettingsShell() {
  const {
    blockedState,
    retry,
    state,
  } = useCreatorWorkspaceSummary();

  if (blockedState !== null) {
    return <CreatorModeShell state={blockedState} />;
  }

  if (state.kind === "loading") {
    return <CreatorProfileSettingsLoading />;
  }

  if (state.kind === "error") {
    return <CreatorProfileSettingsError message={state.message} onRetry={retry} />;
  }

  return (
    <CreatorProfileSettingsForm
      initialValues={{
        avatarUrl: state.summary.creator.avatar?.url ?? null,
        bio: state.summary.creator.bio,
        displayName: state.summary.creator.displayName,
        handle: state.summary.creator.handle,
      }}
    />
  );
}
