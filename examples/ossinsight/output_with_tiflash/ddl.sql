CREATE INDEX idx_event_year_action ON gharchive_dev.github_events (event_year, action);
CREATE INDEX idx_primary_language_repo_name_created_at ON gharchive_dev.github_repos (primary_language, repo_name, created_at);
CREATE INDEX idx_primary_language_stars_repo_name ON gharchive_dev.github_repos (primary_language, stars, repo_name);
CREATE INDEX idx_stars_created_at ON gharchive_dev.github_repos (stars, created_at);
CREATE INDEX idx_updated_at ON gharchive_dev.github_repos (updated_at)