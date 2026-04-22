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
import type {
  CreatorRegistrationEvidence,
  CreatorRegistrationEvidenceKind,
  CreatorRegistrationIntake,
  CreatorRegistrationStatus,
} from "../api/contracts";
import { getRequiredCreatorRegistrationEvidenceKinds } from "./creator-registration-evidence-requirements";
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
  acceptsAdultBusinessCompliance: boolean;
  acceptsAppearanceVerification: boolean;
  acceptsConsentResponsibility: boolean;
  acceptsCoPerformerConsentResponsibility: boolean;
  canSubmit: boolean;
  confirmsInformationMatchesDocuments: boolean;
  creatorBio: string;
  declaresNoProhibitedCategory: boolean;
  evidences: Record<CreatorRegistrationEvidenceKind, EvidenceFieldState>;
  hasCoPerformers: boolean;
  identityDocumentType: string;
  isReadOnly: boolean;
  legalAddress: string;
  legalName: string;
  payoutRecipientName: string;
  payoutRecipientType: string;
  registrationState: string | null;
  sharedProfile: CreatorRegistrationIntake["sharedProfile"];
  targetAudienceCategory: string;
  birthDate: string;
};

type UseCreatorRegistrationResult = {
  acceptsAdultBusinessCompliance: boolean;
  acceptsAppearanceVerification: boolean;
  acceptsConsentResponsibility: boolean;
  acceptsCoPerformerConsentResponsibility: boolean;
  creatorBio: string;
  confirmsInformationMatchesDocuments: boolean;
  declaresNoProhibitedCategory: boolean;
  errorMessage: string | null;
  evidences: Record<CreatorRegistrationEvidenceKind, EvidenceFieldState>;
  hasCoPerformers: boolean;
  hasLoaded: boolean;
  identityDocumentType: string;
  isBusy: boolean;
  isLoading: boolean;
  isReadOnly: boolean;
  isSaving: boolean;
  isSubmitting: boolean;
  legalAddress: string;
  legalName: string;
  payoutRecipientName: string;
  payoutRecipientType: string;
  registration: CreatorRegistrationStatus | null;
  registrationState: string | null;
  requiredEvidenceKinds: CreatorRegistrationEvidenceKind[];
  saveDraft: () => Promise<void>;
  setAcceptsAdultBusinessCompliance: (value: boolean) => void;
  setAcceptsAppearanceVerification: (value: boolean) => void;
  setAcceptsConsentResponsibility: (value: boolean) => void;
  setAcceptsCoPerformerConsentResponsibility: (value: boolean) => void;
  setConfirmsInformationMatchesDocuments: (value: boolean) => void;
  setCreatorBio: (value: string) => void;
  setDeclaresNoProhibitedCategory: (value: boolean) => void;
  setHasCoPerformers: (value: boolean) => void;
  setIdentityDocumentType: (value: string) => void;
  setLegalAddress: (value: string) => void;
  setLegalName: (value: string) => void;
  setPayoutRecipientName: (value: string) => void;
  setPayoutRecipientType: (value: string) => void;
  setTargetAudienceCategory: (value: string) => void;
  setBirthDate: (value: string) => void;
  sharedProfile: CreatorRegistrationIntake["sharedProfile"] | null;
  submit: () => Promise<void>;
  submitDisabled: boolean;
  successMessage: string | null;
  targetAudienceCategory: string;
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
    identity_selfie: buildEvidenceFieldState(),
    address_proof: buildEvidenceFieldState(),
    payout_proof: buildEvidenceFieldState(),
    business_registration: buildEvidenceFieldState(),
    co_performer_consent: buildEvidenceFieldState(),
  } satisfies Record<CreatorRegistrationEvidenceKind, EvidenceFieldState>;

  for (const evidence of evidences) {
    record[evidence.kind] = buildEvidenceFieldState(evidence);
  }

  return record;
}

function buildDraftFromIntake(intake: CreatorRegistrationIntake): CreatorRegistrationDraft {
  return {
    acceptsAdultBusinessCompliance: intake.acceptsAdultBusinessCompliance ?? false,
    acceptsAppearanceVerification: intake.acceptsAppearanceVerification ?? false,
    acceptsConsentResponsibility: intake.acceptsConsentResponsibility,
    acceptsCoPerformerConsentResponsibility: intake.acceptsCoPerformerConsentResponsibility ?? false,
    birthDate: intake.birthDate ?? "",
    canSubmit: intake.canSubmit,
    confirmsInformationMatchesDocuments: intake.confirmsInformationMatchesDocuments ?? false,
    creatorBio: intake.creatorBio,
    declaresNoProhibitedCategory: intake.declaresNoProhibitedCategory,
    evidences: buildEvidenceRecord(intake.evidences),
    hasCoPerformers: intake.hasCoPerformers ?? false,
    identityDocumentType: intake.identityDocumentType ?? "",
    isReadOnly: intake.isReadOnly,
    legalAddress: intake.legalAddress ?? "",
    legalName: intake.legalName,
    payoutRecipientName: intake.payoutRecipientName,
    payoutRecipientType: intake.payoutRecipientType ?? "",
    registrationState: intake.registrationState,
    sharedProfile: intake.sharedProfile,
    targetAudienceCategory: intake.targetAudienceCategory ?? "",
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

  const requiredEvidenceKinds = getRequiredCreatorRegistrationEvidenceKinds({
    hasCoPerformers: draft.hasCoPerformers,
    payoutRecipientType: draft.payoutRecipientType,
  });

  return requiredEvidenceKinds.some((kind) => {
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
      acceptsAdultBusinessCompliance: draft.acceptsAdultBusinessCompliance,
      acceptsAppearanceVerification: draft.acceptsAppearanceVerification,
      acceptsConsentResponsibility: draft.acceptsConsentResponsibility,
      acceptsCoPerformerConsentResponsibility: draft.acceptsCoPerformerConsentResponsibility,
      birthDate: draft.birthDate,
      confirmsInformationMatchesDocuments: draft.confirmsInformationMatchesDocuments,
      creatorBio: draft.creatorBio,
      declaresNoProhibitedCategory: draft.declaresNoProhibitedCategory,
      hasCoPerformers: draft.hasCoPerformers,
      identityDocumentType: draft.identityDocumentType,
      legalAddress: draft.legalAddress,
      legalName: draft.legalName,
      payoutRecipientName: draft.payoutRecipientName,
      payoutRecipientType: draft.payoutRecipientType,
      targetAudienceCategory: draft.targetAudienceCategory,
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

  const requiredEvidenceKinds = draft
    ? getRequiredCreatorRegistrationEvidenceKinds({
        hasCoPerformers: draft.hasCoPerformers,
        payoutRecipientType: draft.payoutRecipientType,
      })
    : getRequiredCreatorRegistrationEvidenceKinds({
        hasCoPerformers: false,
        payoutRecipientType: "",
      });

  return {
    acceptsAdultBusinessCompliance: draft?.acceptsAdultBusinessCompliance ?? false,
    acceptsAppearanceVerification: draft?.acceptsAppearanceVerification ?? false,
    acceptsConsentResponsibility: draft?.acceptsConsentResponsibility ?? false,
    acceptsCoPerformerConsentResponsibility: draft?.acceptsCoPerformerConsentResponsibility ?? false,
    birthDate: draft?.birthDate ?? "",
    confirmsInformationMatchesDocuments: draft?.confirmsInformationMatchesDocuments ?? false,
    creatorBio: draft?.creatorBio ?? "",
    declaresNoProhibitedCategory: draft?.declaresNoProhibitedCategory ?? false,
    errorMessage,
    evidences:
      draft?.evidences ??
      buildEvidenceRecord([]),
    hasCoPerformers: draft?.hasCoPerformers ?? false,
    hasLoaded,
    identityDocumentType: draft?.identityDocumentType ?? "",
    isBusy,
    isLoading,
    isReadOnly: draft?.isReadOnly ?? true,
    isSaving,
    isSubmitting,
    legalAddress: draft?.legalAddress ?? "",
    legalName: draft?.legalName ?? "",
    payoutRecipientName: draft?.payoutRecipientName ?? "",
    payoutRecipientType: draft?.payoutRecipientType ?? "",
    registration,
    registrationState: draft?.registrationState ?? null,
    requiredEvidenceKinds,
    saveDraft,
    setAcceptsAdultBusinessCompliance: (value) => {
      updateDraft((current) => ({
        ...current,
        acceptsAdultBusinessCompliance: value,
        canSubmit: false,
      }));
      setSuccessMessage(null);
    },
    setAcceptsAppearanceVerification: (value) => {
      updateDraft((current) => ({
        ...current,
        acceptsAppearanceVerification: value,
        canSubmit: false,
      }));
      setSuccessMessage(null);
    },
    setAcceptsConsentResponsibility: (value) => {
      updateDraft((current) => ({
        ...current,
        acceptsConsentResponsibility: value,
        canSubmit: false,
      }));
      setSuccessMessage(null);
    },
    setAcceptsCoPerformerConsentResponsibility: (value) => {
      updateDraft((current) => ({
        ...current,
        acceptsCoPerformerConsentResponsibility: value,
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
    setConfirmsInformationMatchesDocuments: (value) => {
      updateDraft((current) => ({
        ...current,
        confirmsInformationMatchesDocuments: value,
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
    setHasCoPerformers: (value) => {
      updateDraft((current) => ({
        ...current,
        acceptsCoPerformerConsentResponsibility: value
          ? current.acceptsCoPerformerConsentResponsibility
          : false,
        canSubmit: false,
        hasCoPerformers: value,
      }));
      setSuccessMessage(null);
    },
    setIdentityDocumentType: (value) => {
      updateDraft((current) => ({
        ...current,
        canSubmit: false,
        identityDocumentType: value,
      }));
      setSuccessMessage(null);
    },
    setLegalAddress: (value) => {
      updateDraft((current) => ({
        ...current,
        canSubmit: false,
        legalAddress: value,
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
    setTargetAudienceCategory: (value) => {
      updateDraft((current) => ({
        ...current,
        canSubmit: false,
        targetAudienceCategory: value,
      }));
      setSuccessMessage(null);
    },
    sharedProfile: draft?.sharedProfile ?? null,
    submit,
    submitDisabled,
    successMessage,
    targetAudienceCategory: draft?.targetAudienceCategory ?? "",
    uploadEvidence,
  };
}
