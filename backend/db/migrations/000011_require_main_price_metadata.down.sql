ALTER TABLE app.mains
    DROP CONSTRAINT mains_price_minor_check,
    ALTER COLUMN price_minor DROP NOT NULL,
    ALTER COLUMN currency_code DROP NOT NULL,
    ADD CONSTRAINT mains_check CHECK (
        (price_minor IS NULL AND currency_code IS NULL)
        OR (price_minor IS NOT NULL AND currency_code IS NOT NULL)
    ),
    ADD CONSTRAINT mains_price_minor_check CHECK (
        price_minor IS NULL
        OR price_minor >= 0
    );
