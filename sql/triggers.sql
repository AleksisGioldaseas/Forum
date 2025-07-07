-- PREVENT DUPLICATE MOD REQUEST
CREATE TRIGGER prevent_monthly_duplicate_modrequest
BEFORE INSERT ON ModRequest
FOR EACH ROW
BEGIN
    SELECT 
        RAISE(ABORT, 'Sender has already submitted a mod request in the past month.')
    WHERE EXISTS (
        SELECT 1 FROM ModRequest
        WHERE SenderId = NEW.SenderId
          AND Created >= datetime('now', '-1 month')
    );
END;
