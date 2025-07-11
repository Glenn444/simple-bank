-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION validate_transfer_balance()
RETURNS TRIGGER AS $$
DECLARE
    current_balance numeric(9,2);
BEGIN
    -- Get the current balance of the from_account
    SELECT balance INTO current_balance 
    FROM accounts 
    WHERE id = NEW.from_account_id;
    
    -- Check if sufficient balance exists
    IF current_balance < NEW.amount THEN
        RAISE EXCEPTION 'Insufficient balance. Available: %, Required: %', 
            current_balance, NEW.amount;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_validate_transfer_balance
    BEFORE INSERT ON transfers
    FOR EACH ROW
    EXECUTE FUNCTION validate_transfer_balance();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_validate_transfer_balance ON transfers;
DROP FUNCTION IF EXISTS validate_transfer_balance();
-- +goose StatementEnd
