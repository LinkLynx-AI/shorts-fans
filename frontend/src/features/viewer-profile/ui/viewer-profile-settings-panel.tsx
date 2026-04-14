"use client";

import { startTransition, useState } from "react";
import { useRouter } from "next/navigation";

import { updateViewerProfile } from "../api/update-viewer-profile";
import {
  getViewerProfileErrorCode,
  getViewerProfileSaveErrorMessage,
  type ViewerProfileInitialValues,
} from "../model/viewer-profile";
import { useViewerProfileDraft } from "../model/use-viewer-profile-draft";
import { ProfileEditorPanel } from "./profile-editor-panel";
import { SharedViewerProfileFields } from "./shared-viewer-profile-fields";

type ViewerProfileSettingsPanelProps = {
  initialValues: ViewerProfileInitialValues;
};

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
    <ProfileEditorPanel
      backHref="/fan"
      backLabel="fan hub に戻る"
      description="fan / creator 共通の表示名、handle、avatar を更新できます。workspace からの変更導線はそのまま残ります。"
      errorMessage={errorMessage}
      eyebrow="fan settings"
      isSubmitting={isSubmitting}
      onSubmit={submit}
      submitLabel="保存する"
      submittingLabel="保存中..."
      title="プロフィールを編集"
    >
      <SharedViewerProfileFields
        avatar={draft.avatar}
        avatarInputKey={draft.avatarInputKey}
        displayName={draft.displayName}
        handle={draft.handle}
        isSubmitting={isSubmitting}
        onAvatarClear={draft.clearAvatarSelection}
        onAvatarSelect={draft.selectAvatarFile}
        onDisplayNameChange={draft.setDisplayName}
        onHandleChange={draft.setHandle}
      />
    </ProfileEditorPanel>
  );
}
