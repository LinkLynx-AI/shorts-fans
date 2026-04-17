"use client";

import { useRouter } from "next/navigation";
import { startTransition, useEffect, useState } from "react";

import {
  completeCreatorRegistrationEvidenceUpload,
  createCreatorRegistrationEvidenceUpload,
  fetchCreatorRegistrationIntake,
  registerCreator,
  saveCreatorRegistrationIntake,
  uploadCreatorRegistrationEvidenceTarget,
} from "../api";
import {
  creatorRegistrationEvidenceKinds,
  type CreatorRegistrationEvidence,
  type CreatorRegistrationEvidenceKind,
  type CreatorRegistrationIntake,
} from "../api/contracts";
import { getCreatorRegistrationErrorMessage } from "./creator-entry";

type EvidenceFieldState = {
  errorMessage: string | null;
  evidence: CreatorRegistrationEvidence | null;
  inputKey: number;
  isUploading: boolean;
};

type CreatorRegistrationDraft = {
  acceptsConsentResponsibility: boolean;
  creatorBio: string;
  declaresNoProhibitedCategory: boolean;
  evidences: Record<CreatorRegistrationEvidenceKind, EvidenceFieldState>;
  isReadOnly: boolean;
  legalName: string;
  payoutRecipientName: string;
  payoutRecipientType: string;
  registrationState: string | null;
  sharedProfile: CreatorRegistrationIntake["sharedProfile"];
  birthDate: string;
};

type UseCreatorRegistrationResult = {
  acceptsConsentResponsibility: boolean;
  creatorBio: string;
  declaresNoProhibitedCategory: boolean;
  errorMessage: string | null;
  evidences: Record<CreatorRegistrationEvidenceKind, EvidenceFieldState>;
  hasLoaded: boolean;
  isBusy: boolean;
  isLoading: boolean;
  isReadOnly: boolean;
  isSaving: boolean;
  isSubmitting: boolean;
  legalName: string;
  payoutRecipientName: string;
  payoutRecipientType: string;
  registrationState: string | null;
  saveDraft: () => Promise<void>;
  setAcceptsConsentResponsibility: (value: boolean) => void;
  setCreatorBio: (value: string) => void;
  setDeclaresNoProhibitedCategory: (value: boolean) => void;
  setLegalName: (value: string) => void;
  setPayoutRecipientName: (value: string) => void;
  setPayoutRecipientType: (value: string) => void;
  setBirthDate: (value: string) => void;
  sharedProfile: CreatorRegistrationIntake["sharedProfile"] | null;
  submit: () => Promise<void>;
  submitDisabled: boolean;
  successMessage: string | null;
  uploadEvidence: (kind: CreatorRegistrationEvidenceKind, file: File | null) => Promise<void>;
  birthDate: string;
};

function buildEvidenceFieldState(
  evidence: CreatorRegistrationEvidence | null = null,
): EvidenceFieldState {
  return {
    errorMessage: null,
    evidence,
    inputKey: 0,
    isUploading: false,
  };
}

function buildEvidenceRecord(
  evidences: CreatorRegistrationEvidence[],
): Record<CreatorRegistrationEvidenceKind, EvidenceFieldState> {
  const record = {
    government_id: buildEvidenceFieldState(),
    payout_proof: buildEvidenceFieldState(),
  } satisfies Record<CreatorRegistrationEvidenceKind, EvidenceFieldState>;

  for (const evidence of evidences) {
    record[evidence.kind] = buildEvidenceFieldState(evidence);
  }

  return record;
}

function buildDraftFromIntake(intake: CreatorRegistrationIntake): CreatorRegistrationDraft {
  return {
    acceptsConsentResponsibility: intake.acceptsConsentResponsibility,
    birthDate: intake.birthDate ?? "",
    creatorBio: intake.creatorBio,
    declaresNoProhibitedCategory: intake.declaresNoProhibitedCategory,
    evidences: buildEvidenceRecord(intake.evidences),
    isReadOnly: intake.isReadOnly,
    legalName: intake.legalName,
    payoutRecipientName: intake.payoutRecipientName,
    payoutRecipientType: intake.payoutRecipientType ?? "",
    registrationState: intake.registrationState,
    sharedProfile: intake.sharedProfile,
  };
}

function computeSubmitDisabled(draft: CreatorRegistrationDraft | null, isBusy: boolean): boolean {
  if (isBusy || draft === null || draft.isReadOnly) {
    return true;
  }

  if (
    draft.creatorBio.trim() === "" ||
    draft.legalName.trim() === "" ||
    draft.birthDate.trim() === "" ||
    draft.payoutRecipientType.trim() === "" ||
    draft.payoutRecipientName.trim() === "" ||
    !draft.declaresNoProhibitedCategory ||
    !draft.acceptsConsentResponsibility
  ) {
    return true;
  }

  return creatorRegistrationEvidenceKinds.some((kind) => {
    const evidence = draft.evidences[kind];
    return evidence.isUploading || evidence.evidence === null;
  });
}

export function useCreatorRegistration(): UseCreatorRegistrationResult {
  const router = useRouter();
  const [draft, setDraft] = useState<CreatorRegistrationDraft | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [hasLoaded, setHasLoaded] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadIntake() {
      setIsLoading(true);
      setErrorMessage(null);

      try {
        const intake = await fetchCreatorRegistrationIntake();
        if (cancelled) {
          return;
        }

        if (intake.registrationState === "submitted") {
          startTransition(() => {
            router.replace("/fan/creator/success");
          });
          return;
        }

        setDraft(buildDraftFromIntake(intake));
      } catch (error) {
        if (!cancelled) {
          setErrorMessage(getCreatorRegistrationErrorMessage(error));
        }
      } finally {
        if (!cancelled) {
          setHasLoaded(true);
          setIsLoading(false);
        }
      }
    }

    void loadIntake();

    return () => {
      cancelled = true;
    };
  }, [router]);

  const isBusy = isLoading || isSaving || isSubmitting;
  const submitDisabled = computeSubmitDisabled(draft, isBusy);

  const updateDraft = (updater: (current: CreatorRegistrationDraft) => CreatorRegistrationDraft) => {
    setDraft((current) => {
      if (current === null) {
        return current;
      }

      return updater(current);
    });
  };

  const persistDraft = async (): Promise<CreatorRegistrationIntake | null> => {
    if (draft === null) {
      return null;
    }

    const intake = await saveCreatorRegistrationIntake({
      acceptsConsentResponsibility: draft.acceptsConsentResponsibility,
      birthDate: draft.birthDate,
      creatorBio: draft.creatorBio,
      declaresNoProhibitedCategory: draft.declaresNoProhibitedCategory,
      legalName: draft.legalName,
      payoutRecipientName: draft.payoutRecipientName,
      payoutRecipientType: draft.payoutRecipientType,
    });

    setDraft(buildDraftFromIntake(intake));
    return intake;
  };

  const saveDraft = async () => {
    if (draft === null || draft.isReadOnly || isBusy) {
      return;
    }

    setIsSaving(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      await persistDraft();
      setSuccessMessage("下書きを保存しました。");
    } catch (error) {
      setErrorMessage(getCreatorRegistrationErrorMessage(error));
    } finally {
      setIsSaving(false);
    }
  };

  const submit = async () => {
    if (draft === null || draft.isReadOnly || isBusy) {
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      await persistDraft();
      await registerCreator();

      startTransition(() => {
        router.push("/fan/creator/success");
      });
    } catch (error) {
      setErrorMessage(getCreatorRegistrationErrorMessage(error));
    } finally {
      setIsSubmitting(false);
    }
  };

  const uploadEvidence = async (kind: CreatorRegistrationEvidenceKind, file: File | null) => {
    if (draft === null || draft.isReadOnly || isBusy || file === null) {
      return;
    }

    updateDraft((current) => ({
      ...current,
      evidences: {
        ...current.evidences,
        [kind]: {
          ...current.evidences[kind],
          errorMessage: null,
          isUploading: true,
        },
      },
    }));
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      const created = await createCreatorRegistrationEvidenceUpload(kind, file);
      await uploadCreatorRegistrationEvidenceTarget({
        file,
        target: created.uploadTarget,
      });
      const completed = await completeCreatorRegistrationEvidenceUpload(created.evidenceUploadToken);

      updateDraft((current) => ({
        ...current,
        evidences: {
          ...current.evidences,
          [kind]: {
            errorMessage: null,
            evidence: completed.evidence,
            inputKey: current.evidences[kind].inputKey + 1,
            isUploading: false,
          },
        },
      }));
    } catch (error) {
      updateDraft((current) => ({
        ...current,
        evidences: {
          ...current.evidences,
          [kind]: {
            ...current.evidences[kind],
            errorMessage: getCreatorRegistrationErrorMessage(error),
            inputKey: current.evidences[kind].inputKey + 1,
            isUploading: false,
          },
        },
      }));
    }
  };

  return {
    acceptsConsentResponsibility: draft?.acceptsConsentResponsibility ?? false,
    birthDate: draft?.birthDate ?? "",
    creatorBio: draft?.creatorBio ?? "",
    declaresNoProhibitedCategory: draft?.declaresNoProhibitedCategory ?? false,
    errorMessage,
    evidences:
      draft?.evidences ??
      buildEvidenceRecord([]),
    hasLoaded,
    isBusy,
    isLoading,
    isReadOnly: draft?.isReadOnly ?? true,
    isSaving,
    isSubmitting,
    legalName: draft?.legalName ?? "",
    payoutRecipientName: draft?.payoutRecipientName ?? "",
    payoutRecipientType: draft?.payoutRecipientType ?? "",
    registrationState: draft?.registrationState ?? null,
    saveDraft,
    setAcceptsConsentResponsibility: (value) => {
      updateDraft((current) => ({
        ...current,
        acceptsConsentResponsibility: value,
      }));
      setSuccessMessage(null);
    },
    setBirthDate: (value) => {
      updateDraft((current) => ({
        ...current,
        birthDate: value,
      }));
      setSuccessMessage(null);
    },
    setCreatorBio: (value) => {
      updateDraft((current) => ({
        ...current,
        creatorBio: value,
      }));
      setSuccessMessage(null);
    },
    setDeclaresNoProhibitedCategory: (value) => {
      updateDraft((current) => ({
        ...current,
        declaresNoProhibitedCategory: value,
      }));
      setSuccessMessage(null);
    },
    setLegalName: (value) => {
      updateDraft((current) => ({
        ...current,
        legalName: value,
      }));
      setSuccessMessage(null);
    },
    setPayoutRecipientName: (value) => {
      updateDraft((current) => ({
        ...current,
        payoutRecipientName: value,
      }));
      setSuccessMessage(null);
    },
    setPayoutRecipientType: (value) => {
      updateDraft((current) => ({
        ...current,
        payoutRecipientType: value,
      }));
      setSuccessMessage(null);
    },
    sharedProfile: draft?.sharedProfile ?? null,
    submit,
    submitDisabled,
    successMessage,
    uploadEvidence,
  };
}
