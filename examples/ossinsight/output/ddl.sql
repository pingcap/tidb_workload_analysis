CREATE INDEX idx_type_action_actor_id ON gharchive_dev.github_events (type, action, actor_id);
CREATE INDEX idx_type_action_created_at ON gharchive_dev.github_events (type, action, created_at);
CREATE INDEX idx_type_repo_id_action ON gharchive_dev.github_events (type, repo_id, action);
CREATE INDEX idx_type_repo_id_repo_name ON gharchive_dev.github_events (type, repo_id, repo_name);
CREATE INDEX idx_stars_primary_language_pushed_at ON gharchive_dev.github_repos (stars, primary_language, pushed_at)