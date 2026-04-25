-- +goose Up
-- +goose StatementBegin

-- Remove repository_id from manifests: manifests are content-addressed globally
ALTER TABLE manifests DROP COLUMN repository_id;

-- Fix tag unique constraint to allow re-creating deleted tags
ALTER TABLE tags DROP CONSTRAINT IF EXISTS tags_repository_id_name_key;
CREATE UNIQUE INDEX idx_tags_active ON tags(repository_id, name) WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_tags_active;
ALTER TABLE tags ADD CONSTRAINT tags_repository_id_name_key UNIQUE(repository_id, name);
ALTER TABLE manifests ADD COLUMN repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE;

-- +goose StatementEnd
