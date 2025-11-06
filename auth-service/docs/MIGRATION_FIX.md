# Migration Fix: token_family_id Column

## Vấn đề

Khi run application với code mới, gặp lỗi:

```
ERROR: column "token_family_id" of relation "refresh_tokens" contains null values (SQLSTATE 23502)
```

**Nguyên nhân:**
- Code mới thêm field `token_family_id UUID NOT NULL`
- Database đã có data cũ (refresh_tokens không có token_family_id)
- GORM AutoMigrate cố thêm column NOT NULL → lỗi vì existing rows = NULL

## Giải pháp

### Option 1: Manual SQL Migration (Khuyên dùng nếu có production data)

Chạy script `scripts/migrate_token_family.sql`:

```bash
# Connect to database
docker exec -it <postgres-container> psql -U postgres -d auth_db

# Or if postgres local
psql -h localhost -U postgres -d auth_db
```

Paste nội dung:

```sql
-- Step 1: Add column as NULLABLE
ALTER TABLE refresh_tokens 
ADD COLUMN IF NOT EXISTS token_family_id UUID NULL;

-- Step 2: Generate family ID for existing rows
UPDATE refresh_tokens 
SET token_family_id = gen_random_uuid() 
WHERE token_family_id IS NULL;

-- Step 3: Make it NOT NULL (after all rows have values)
ALTER TABLE refresh_tokens 
ALTER COLUMN token_family_id SET NOT NULL;

-- Step 4: Add index
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_family_id 
ON refresh_tokens(token_family_id);

-- Verify
SELECT COUNT(*) as total_tokens, 
       COUNT(token_family_id) as tokens_with_family 
FROM refresh_tokens;
```

### Option 2: Clean Start (Development only)

```bash
# Drop and recreate database
docker exec -it <postgres-container> psql -U postgres -c "DROP DATABASE auth_db;"
docker exec -it <postgres-container> psql -U postgres -c "CREATE DATABASE auth_db;"

# Restart application (AutoMigrate will create fresh schema)
go run cmd/server/main.go
```

### Option 3: Delete existing refresh tokens

```sql
-- Delete all existing refresh tokens (users will need to login again)
DELETE FROM refresh_tokens;

-- Then restart application
```

## Code Changes Made

Để tránh lỗi này trong tương lai, đã update model:

```go
// BEFORE (caused error with existing data)
type RefreshTokenModel struct {
    TokenFamilyID uuid.UUID `gorm:"type:uuid;not null;index"`
}

// AFTER (allows NULL, handles migration gracefully)
type RefreshTokenModel struct {
    TokenFamilyID *uuid.UUID `gorm:"type:uuid;index"`  // Pointer = nullable
}
```

Và handle NULL trong conversion:

```go
func (r *RefreshTokenRepository) toEntity(model *RefreshTokenModel) *entity.RefreshToken {
    var familyID uuid.UUID
    if model.TokenFamilyID != nil {
        familyID = *model.TokenFamilyID
    }
    // Old tokens will have zero UUID as familyID (ok for existing tokens)
    // New tokens always get a proper family ID
    return &entity.RefreshToken{
        TokenFamilyID: familyID,
        // ...
    }
}
```

## Recommended Steps

**For Development:**
```bash
# 1. Stop application
# 2. Choose option 2 or 3 (clean start)
# 3. Restart application
```

**For Production/Staging (with existing users):**
```bash
# 1. Stop application
# 2. Run migration script (Option 1)
# 3. Verify data: SELECT COUNT(*) FROM refresh_tokens WHERE token_family_id IS NULL;
# 4. Should return 0
# 5. Deploy new code
# 6. Restart application
```

## Verification

After migration, verify:

```sql
-- Check all tokens have family ID
SELECT COUNT(*) as tokens_without_family 
FROM refresh_tokens 
WHERE token_family_id IS NULL;
-- Should return: 0

-- Check index exists
SELECT indexname 
FROM pg_indexes 
WHERE tablename = 'refresh_tokens' 
  AND indexname = 'idx_refresh_tokens_token_family_id';
-- Should return: idx_refresh_tokens_token_family_id

-- Sample data
SELECT id, user_id, token_family_id, is_revoked, created_at 
FROM refresh_tokens 
LIMIT 5;
```

## Future Prevention

Lesson learned: When adding NOT NULL columns to existing tables:

1. ✅ Add column as NULLABLE first
2. ✅ Populate data for existing rows
3. ✅ Then make it NOT NULL
4. ✅ Or use pointer types in Go (*uuid.UUID) to allow NULL

**Never add NOT NULL column directly to table with existing data!**
