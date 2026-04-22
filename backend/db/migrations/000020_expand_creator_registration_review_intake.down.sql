DELETE FROM app.creator_registration_evidences
WHERE kind NOT IN (
    'government_id',
    'payout_proof'
);

ALTER TABLE app.creator_registration_evidences
    DROP CONSTRAINT creator_registration_evidences_kind_check,
    ADD CONSTRAINT creator_registration_evidences_kind_check CHECK (
        kind IN (
            'government_id',
            'payout_proof'
        )
    ) NOT VALID;

ALTER TABLE app.creator_registration_evidences
    VALIDATE CONSTRAINT creator_registration_evidences_kind_check;

ALTER TABLE app.creator_registration_intakes
    DROP CONSTRAINT creator_registration_intakes_target_audience_category_check,
    DROP CONSTRAINT creator_registration_intakes_identity_document_type_check,
    DROP CONSTRAINT creator_registration_intakes_legal_address_length_check,
    DROP COLUMN confirms_information_matches_documents,
    DROP COLUMN accepts_adult_business_compliance,
    DROP COLUMN accepts_co_performer_consent_responsibility,
    DROP COLUMN accepts_appearance_verification,
    DROP COLUMN has_co_performers,
    DROP COLUMN target_audience_category,
    DROP COLUMN identity_document_type,
    DROP COLUMN legal_address;
