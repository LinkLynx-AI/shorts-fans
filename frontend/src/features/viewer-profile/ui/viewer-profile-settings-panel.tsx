"use client";

import { startTransition, useState } from "react";
import { useRouter } from "next/navigation";

import { updateViewerProfile } from "../api/update-viewer-profile";
import { useViewerProfileDraft } from "../model/use-viewer-profile-draft";
import {
  getViewerProfileErrorCode,
  getViewerProfileSaveErrorMessage,
  type ViewerProfileInitialValues,
} from "../model/viewer-profile";
import { ViewerProfileEditorForm } from "./viewer-profile-editor-form";

type ViewerProfileSettingsPanelProps = {
  initialValues: ViewerProfileInitialValues;
};

/**
 * fan settings から shared viewer profile を参考デザインに沿って編集する。
 */
export function ViewerProfileSettingsPanel({ initialValues }: ViewerProfileSettingsPanelProps) {
  const router = useRouter();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const draft = useViewerProfileDraft({
    initialValues,
    mode: "edit",
    onDirty: () => {
      if (errorMessage !== null) {
        setErrorMessage(null);
      }
    },
  });

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

      await updateViewerProfile({
        ...(avatarUploadToken ? { avatarUploadToken } : {}),
        displayName: draft.displayName,
        handle: draft.handle,
      });

      startTransition(() => {
        router.replace("/fan");
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
      backHref="/fan"
      backLabel="fan hub に戻る"
      displayName={draft.displayName}
      errorMessage={errorMessage}
      handle={draft.handle}
      isSubmitting={isSubmitting}
      onAvatarClear={draft.clearAvatarSelection}
      onAvatarSelect={draft.selectAvatarFile}
      onDisplayNameChange={draft.setDisplayName}
      onHandleChange={draft.setHandle}
      onSubmit={submit}
    />
  );
}
