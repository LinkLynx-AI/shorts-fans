ALTER TABLE app.creator_registration_intakes
    ADD COLUMN legal_address TEXT NOT NULL DEFAULT '',
    ADD COLUMN identity_document_type TEXT,
    ADD COLUMN target_audience_category TEXT,
    ADD COLUMN has_co_performers BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN accepts_appearance_verification BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN accepts_co_performer_consent_responsibility BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN accepts_adult_business_compliance BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN confirms_information_matches_documents BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE app.creator_registration_intakes
    ADD CONSTRAINT creator_registration_intakes_legal_address_length_check CHECK (
        char_length(legal_address) <= 500
    ) NOT VALID,
    ADD CONSTRAINT creator_registration_intakes_identity_document_type_check CHECK (
        identity_document_type IS NULL
        OR identity_document_type IN (
            'driver_license',
            'my_number_card',
            'residence_card',
            'basic_resident_register_card',
            'passport',
            'student_or_employee_id',
            'disability_certificate',
            'other_government_photo_id'
        )
    ) NOT VALID,
    ADD CONSTRAINT creator_registration_intakes_target_audience_category_check CHECK (
        target_audience_category IS NULL
        OR target_audience_category IN (
            'all_ages',
            'general_adult',
            'gay_bl'
        )
    ) NOT VALID;

ALTER TABLE app.creator_registration_intakes
    VALIDATE CONSTRAINT creator_registration_intakes_legal_address_length_check,
    VALIDATE CONSTRAINT creator_registration_intakes_identity_document_type_check,
    VALIDATE CONSTRAINT creator_registration_intakes_target_audience_category_check;

ALTER TABLE app.creator_registration_evidences
    DROP CONSTRAINT creator_registration_evidences_kind_check,
    ADD CONSTRAINT creator_registration_evidences_kind_check CHECK (
        kind IN (
            'government_id',
            'payout_proof',
            'identity_selfie',
            'address_proof',
            'business_registration',
            'co_performer_consent'
        )
    ) NOT VALID;

ALTER TABLE app.creator_registration_evidences
    VALIDATE CONSTRAINT creator_registration_evidences_kind_check;
