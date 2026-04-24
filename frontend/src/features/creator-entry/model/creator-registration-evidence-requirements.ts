import type { CreatorRegistrationEvidenceKind } from "../api/contracts";

export function getRequiredCreatorRegistrationEvidenceKinds({
  hasCoPerformers,
  payoutRecipientType,
}: {
  hasCoPerformers: boolean;
  payoutRecipientType: string;
}): CreatorRegistrationEvidenceKind[] {
  const kinds: CreatorRegistrationEvidenceKind[] = [
    "government_id",
    "identity_selfie",
    "address_proof",
    "payout_proof",
  ];

  if (payoutRecipientType === "business") {
    kinds.push("business_registration");
  }
  if (hasCoPerformers) {
    kinds.push("co_performer_consent");
  }

  return kinds;
}
