-- Drop foreign key constraint if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'embedding_documents_tenant_id_fk'
    ) THEN
        ALTER TABLE embedding_documents DROP CONSTRAINT embedding_documents_tenant_id_fk;
    END IF;
END $$;

DROP TABLE IF EXISTS embedding_documents;
DROP EXTENSION IF EXISTS vector;

