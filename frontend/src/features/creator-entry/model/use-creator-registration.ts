"use client";

import { useRouter } from "next/navigation";
import { startTransition, useEffect, useState } from "react";

import {
  completeCreatorRegistrationEvidenceUpload,
  createCreatorRegistrationEvidenceUpload,
  fetchCreatorRegistration,
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
  type CreatorRegistrationStatus,
} from "../api/contracts";
import {
  getCreatorEntryErrorCode,
  getCreatorRegistrationErrorMessage,
} from "./creator-entry";

type EvidenceFieldState = {
  errorMessage: string | null;
  evidence: CreatorRegistrationEvidence | null;
  inputKey: number;
  isUploading: boolean;
};

type CreatorRegistrationDraft = {
  acceptsConsentResponsibility: boolean;
  canSubmit: boolean;
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
  registration: CreatorRegistrationStatus | null;
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

type LoadedRegistrationState =
  | {
      redirectTo: string;
    }
  | {
      draft: CreatorRegistrationDraft;
      registration: CreatorRegistrationStatus | null;
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
    canSubmit: intake.canSubmit,
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

function reconcileRegistration(
  initialRegistration: CreatorRegistrationStatus | null,
  intake: CreatorRegistrationIntake,
): CreatorRegistrationStatus | null {
  if (initialRegistration === null || initialRegistration.state !== intake.registrationState) {
    return null;
  }

  if (
    initialRegistration.state === "rejected" &&
    initialRegistration.actions.canResubmit !== !intake.isReadOnly
  ) {
    return null;
  }

  return initialRegistration;
}

function shouldRefreshRejectedRegistration(
  reconciledRegistration: CreatorRegistrationStatus | null,
  intake: CreatorRegistrationIntake,
): boolean {
  if (intake.registrationState !== "rejected") {
    return false;
  }

  return intake.isReadOnly || reconciledRegistration === null;
}

async function resolveRegistrationForIntake(
  initialRegistration: CreatorRegistrationStatus | null,
  intake: CreatorRegistrationIntake,
): Promise<CreatorRegistrationStatus | null> {
  const reconciledRegistration = reconcileRegistration(initialRegistration, intake);

  if (!shouldRefreshRejectedRegistration(reconciledRegistration, intake)) {
    return reconciledRegistration;
  }

  try {
    const refreshedRegistration = await fetchCreatorRegistration();
    return reconcileRegistration(refreshedRegistration ?? null, intake);
  } catch {
    return null;
  }
}

function computeSubmitDisabled(draft: CreatorRegistrationDraft | null, isBusy: boolean): boolean {
  if (isBusy || draft === null || draft.isReadOnly) {
    return true;
  }

  if (!draft.canSubmit) {
    return true;
  }

  return creatorRegistrationEvidenceKinds.some((kind) => {
    const evidence = draft.evidences[kind];
    return evidence.isUploading || evidence.evidence === null;
  });
}

export function useCreatorRegistration(
  initialRegistration: CreatorRegistrationStatus | null,
): UseCreatorRegistrationResult {
  const router = useRouter();
  const [draft, setDraft] = useState<CreatorRegistrationDraft | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [hasLoaded, setHasLoaded] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [registration, setRegistration] = useState<CreatorRegistrationStatus | null>(initialRegistration);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const loadDraftState = async (
    registrationSeed: CreatorRegistrationStatus | null,
  ): Promise<LoadedRegistrationState> => {
    const intake = await fetchCreatorRegistrationIntake();

    if (registrationSeed?.state === "submitted" || intake.registrationState === "submitted") {
      return { redirectTo: "/fan/creator/success" };
    }

    if (registrationSeed?.actions.canEnterCreatorMode || intake.registrationState === "approved") {
      return { redirectTo: "/fan" };
    }

    const resolvedRegistration = await resolveRegistrationForIntake(registrationSeed, intake);

    if (resolvedRegistration?.actions.canEnterCreatorMode) {
      return { redirectTo: "/fan" };
    }

    return {
      draft: buildDraftFromIntake(intake),
      registration: resolvedRegistration,
    };
  };

  const applyLoadedState = (loadedState: LoadedRegistrationState) => {
    if ("redirectTo" in loadedState) {
      startTransition(() => {
        router.replace(loadedState.redirectTo);
      });
      return;
    }

    setDraft(loadedState.draft);
    setRegistration(loadedState.registration);
  };

  const refreshStateAfterConflict = async (error: unknown) => {
    if (getCreatorEntryErrorCode(error) !== "registration_state_conflict") {
      return false;
    }

    setIsLoading(true);
    try {
      const loadedState = await loadDraftState(registration);
      applyLoadedState(loadedState);
    } catch {
      return false;
    } finally {
      setHasLoaded(true);
      setIsLoading(false);
    }

    return true;
  };

  useEffect(() => {
    let cancelled = false;

    async function loadIntake() {
      setIsLoading(true);
      setErrorMessage(null);

      try {
        const loadedState = await loadDraftState(initialRegistration);
        if (cancelled) {
          return;
        }

        if ("redirectTo" in loadedState) {
          startTransition(() => {
            router.replace(loadedState.redirectTo);
          });
          return;
        }

        setDraft(loadedState.draft);
        setRegistration(loadedState.registration);
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
  }, [initialRegistration, router]);

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
      if (await refreshStateAfterConflict(error)) {
        setErrorMessage(getCreatorRegistrationErrorMessage(error));
        return;
      }
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
      if (await refreshStateAfterConflict(error)) {
        setErrorMessage(getCreatorRegistrationErrorMessage(error));
        return;
      }
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
      if (await refreshStateAfterConflict(error)) {
        setErrorMessage(getCreatorRegistrationErrorMessage(error));
        return;
      }
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
    registration,
    registrationState: draft?.registrationState ?? null,
    saveDraft,
    setAcceptsConsentResponsibility: (value) => {
      updateDraft((current) => ({
        ...current,
        acceptsConsentResponsibility: value,
        canSubmit: false,
      }));
      setSuccessMessage(null);
    },
    setBirthDate: (value) => {
      updateDraft((current) => ({
        ...current,
        birthDate: value,
        canSubmit: false,
      }));
      setSuccessMessage(null);
    },
    setCreatorBio: (value) => {
      updateDraft((current) => ({
        ...current,
        canSubmit: false,
        creatorBio: value,
      }));
      setSuccessMessage(null);
    },
    setDeclaresNoProhibitedCategory: (value) => {
      updateDraft((current) => ({
        ...current,
        canSubmit: false,
        declaresNoProhibitedCategory: value,
      }));
      setSuccessMessage(null);
    },
    setLegalName: (value) => {
      updateDraft((current) => ({
        ...current,
        canSubmit: false,
        legalName: value,
      }));
      setSuccessMessage(null);
    },
    setPayoutRecipientName: (value) => {
      updateDraft((current) => ({
        ...current,
        canSubmit: false,
        payoutRecipientName: value,
      }));
      setSuccessMessage(null);
    },
    setPayoutRecipientType: (value) => {
      updateDraft((current) => ({
        ...current,
        canSubmit: false,
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
