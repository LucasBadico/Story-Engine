-- This migration drops old relation tables
-- Rolling back would require recreating them, but we don't have the original data
-- This is a one-way migration for greenfield deployment

-- Note: This down migration is intentionally empty as we cannot recover the old tables
-- without the original data. If you need to rollback, you would need to recreate
-- the tables manually or restore from a backup.

