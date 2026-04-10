UPDATE app.mains
SET
    price_minor = 1200,
    currency_code = 'JPY'
WHERE price_minor IS NULL
    OR price_minor <= 0
    OR currency_code IS NULL
    OR currency_code !~ '^[A-Z]{3}$';

ALTER TABLE app.mains
    DROP CONSTRAINT mains_check,
    DROP CONSTRAINT mains_price_minor_check,
    ALTER COLUMN price_minor SET NOT NULL,
    ALTER COLUMN currency_code SET NOT NULL,
    ADD CONSTRAINT mains_price_minor_check CHECK (price_minor > 0);
