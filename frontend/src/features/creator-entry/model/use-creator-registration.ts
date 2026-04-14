"use client";

import { useRouter } from "next/navigation";
import { startTransition, useState } from "react";

import {
  getCurrentViewerBootstrap,
  useSetCurrentViewer,
  useSetViewerSession,
} from "@/entities/viewer";

import { registerCreator } from "../api/register-creator";
import { getCreatorRegistrationErrorMessage } from "./creator-entry";

type UseCreatorRegistrationResult = {
  errorMessage: string | null;
  isSubmitting: boolean;
  submit: () => Promise<void>;
};

/**
 * creator registration confirm action の submit 状態を管理する。
 */
export function useCreatorRegistration(): UseCreatorRegistrationResult {
  const router = useRouter();
  const setCurrentViewer = useSetCurrentViewer();
  const setViewerSession = useSetViewerSession();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const submit = async () => {
    if (isSubmitting) {
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      await registerCreator();

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
    errorMessage,
    isSubmitting,
    submit,
  };
}
