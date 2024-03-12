CREATE INDEX idx_type_actor_id_actor_login ON gharchive_dev.github_events (type, actor_id, actor_login);
CREATE INDEX idx_type_actor_id_repo_name ON gharchive_dev.github_events (type, actor_id, repo_name);
CREATE INDEX idx_type_repo_id_action ON gharchive_dev.github_events (type, repo_id, action);
CREATE INDEX idx_primary_language_created_at_repo_name ON gharchive_dev.github_repos (primary_language, created_at, repo_name);
CREATE INDEX idx_stars_primary_language_pushed_at ON gharchive_dev.github_repos (stars, primary_language, pushed_at)