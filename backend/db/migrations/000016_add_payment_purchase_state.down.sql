DROP INDEX IF EXISTS app.idx_main_purchase_attempts_user_main_inflight;
DROP INDEX IF EXISTS app.idx_main_purchase_attempts_provider_transaction_ref;
DROP INDEX IF EXISTS app.idx_main_purchase_attempts_provider_purchase_ref;
DROP INDEX IF EXISTS app.idx_main_purchase_attempts_user_main_created_at;
DROP TABLE IF EXISTS app.main_purchase_attempts;

DROP INDEX IF EXISTS app.idx_user_payment_methods_user_last_used_at;
DROP TABLE IF EXISTS app.user_payment_methods;
