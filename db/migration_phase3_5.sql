ALTER TABLE messages ADD COLUMN is_approved BOOLEAN DEFAULT TRUE;
-- Initially set is_approved to TRUE for existing messages (backward compatibility)
UPDATE messages SET is_approved = TRUE;
