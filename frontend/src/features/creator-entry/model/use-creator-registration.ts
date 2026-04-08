"use client";

import { useRouter } from "next/navigation";
import {
  startTransition,
  useState,
} from "react";

import {
  getCurrentViewerBootstrap,
  useSetCurrentViewer,
  useSetViewerSession,
} from "@/entities/viewer";

import { registerCreator } from "../api/register-creator";
import { getCreatorRegistrationErrorMessage } from "./creator-entry";

type UseCreatorRegistrationResult = {
  bio: string;
  displayName: string;
  errorMessage: string | null;
  handle: string;
  isSubmitting: boolean;
  setBio: (bio: string) => void;
  setDisplayName: (displayName: string) => void;
  setHandle: (handle: string) => void;
  submit: () => Promise<void>;
};

/**
 * creator registration form の入力状態と submit を管理する。
 */
export function useCreatorRegistration(): UseCreatorRegistrationResult {
  const router = useRouter();
  const setCurrentViewer = useSetCurrentViewer();
  const setViewerSession = useSetViewerSession();
  const [bio, setBioState] = useState("");
  const [displayName, setDisplayNameState] = useState("");
  const [handle, setHandleState] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const setDisplayName = (nextDisplayName: string) => {
    setDisplayNameState(nextDisplayName);
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const setBio = (nextBio: string) => {
    setBioState(nextBio);
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const setHandle = (nextHandle: string) => {
    setHandleState(nextHandle);
    if (errorMessage !== null) {
      setErrorMessage(null);
    }
  };

  const submit = async () => {
    if (isSubmitting) {
      return;
    }

    if (displayName.trim() === "") {
      setErrorMessage("表示名を入力してください。");
      return;
    }
    if (handle.trim() === "") {
      setErrorMessage("handleを入力してください。");
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await registerCreator({
        bio,
        displayName,
        handle,
      });

      const currentViewer = await getCurrentViewerBootstrap({
        credentials: "include",
      }).catch(() => null);

      if (currentViewer === null) {
        setErrorMessage("登録後の状態反映を確認できませんでした。画面を更新して確認してください。");
        return;
      }

      setCurrentViewer(currentViewer);
      setViewerSession(true);

      startTransition(() => {
        router.push("/fan/creator/success");
      });
    } catch (error) {
      setErrorMessage(getCreatorRegistrationErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    bio,
    displayName,
    errorMessage,
    handle,
    isSubmitting,
    setBio,
    setDisplayName,
    setHandle,
    submit,
  };
}
