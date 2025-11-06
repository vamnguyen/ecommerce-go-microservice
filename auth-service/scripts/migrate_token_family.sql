-- Migration: Add token_family_id to refresh_tokens
-- This script handles existing data safely

-- Step 1: Add column as NULLABLE first (to allow existing rows)
ALTER TABLE refresh_tokens 
ADD COLUMN IF NOT EXISTS token_family_id UUID NULL;

-- Step 2: Generate token_family_id for existing rows
-- Each existing token gets its own unique family
UPDATE refresh_tokens 
SET token_family_id = gen_random_uuid() 
WHERE token_family_id IS NULL;

-- Step 3: Now make it NOT NULL (after all rows have values)
ALTER TABLE refresh_tokens 
ALTER COLUMN token_family_id SET NOT NULL;

-- Step 4: Add index for performance
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_family_id 
ON refresh_tokens(token_family_id);

-- Verify
SELECT COUNT(*) as total_tokens, 
       COUNT(token_family_id) as tokens_with_family 
FROM refresh_tokens;
