import { getRequiredCreatorRegistrationEvidenceKinds } from "./creator-registration-evidence-requirements";

describe("creator registration evidence requirements", () => {
  it("returns the base evidence set for self payout without co-performers", () => {
    expect(getRequiredCreatorRegistrationEvidenceKinds({
      hasCoPerformers: false,
      payoutRecipientType: "self",
    })).toEqual([
      "government_id",
      "identity_selfie",
      "address_proof",
      "payout_proof",
    ]);
  });

  it("adds business registration and co-performer consent when required", () => {
    expect(getRequiredCreatorRegistrationEvidenceKinds({
      hasCoPerformers: true,
      payoutRecipientType: "business",
    })).toEqual([
      "government_id",
      "identity_selfie",
      "address_proof",
      "payout_proof",
      "business_registration",
      "co_performer_consent",
    ]);
  });
});
