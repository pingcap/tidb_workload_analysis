set sql_mode='';
create database IF NOT EXISTS gharchive_dev;
use gharchive_dev;

CREATE TABLE IF NOT EXISTS `github_events` (
                                 `id` bigint(20) NOT NULL DEFAULT '0',
                                 `type` varchar(29) NOT NULL DEFAULT 'Event',
                                 `created_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
                                 `repo_id` bigint(20) NOT NULL DEFAULT '0',
                                 `repo_name` varchar(140) NOT NULL DEFAULT '',
                                 `actor_id` bigint(20) NOT NULL DEFAULT '0',
                                 `actor_login` varchar(40) NOT NULL DEFAULT '',
                                 `language` varchar(26) NOT NULL DEFAULT '',
                                 `additions` bigint(20) NOT NULL DEFAULT '0',
                                 `deletions` bigint(20) NOT NULL DEFAULT '0',
                                 `action` varchar(11) NOT NULL DEFAULT '',
                                 `number` int(11) NOT NULL DEFAULT '0',
                                 `commit_id` varchar(40) NOT NULL DEFAULT '',
                                 `comment_id` bigint(20) NOT NULL DEFAULT '0',
                                 `org_login` varchar(40) NOT NULL DEFAULT '',
                                 `org_id` bigint(20) NOT NULL DEFAULT '0',
                                 `state` varchar(6) NOT NULL DEFAULT '',
                                 `closed_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
                                 `comments` int(11) NOT NULL DEFAULT '0',
                                 `pr_merged_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
                                 `pr_merged` tinyint(1) NOT NULL DEFAULT '0',
                                 `pr_changed_files` int(11) NOT NULL DEFAULT '0',
                                 `pr_review_comments` int(11) NOT NULL DEFAULT '0',
                                 `pr_or_issue_id` bigint(20) NOT NULL DEFAULT '0',
                                 `event_day` date NOT NULL,
                                 `event_month` date NOT NULL,
                                 `event_year` int(11) NOT NULL,
                                 `push_size` int(11) NOT NULL DEFAULT '0',
                                 `push_distinct_size` int(11) NOT NULL DEFAULT '0',
                                 `creator_user_login` varchar(40) NOT NULL DEFAULT '',
                                 `creator_user_id` bigint(20) NOT NULL DEFAULT '0',
                                 `pr_or_issue_created_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
                                 KEY `index_github_events_on_id` (`id`),
                                 KEY `index_github_events_on_created_at` (`created_at`),
                                 KEY `index_ge_on_creator_id_type_action_merged_created_at_add_del` (`creator_user_id`,`type`,`action`,`pr_merged`,`created_at`,`additions`,`deletions`),
                                 KEY `index_ge_on_actor_id_type_action_created_at_repo_id_commits` (`actor_id`,`type`,`action`,`created_at`,`repo_id`,`push_distinct_size`),
                                 KEY `index_ge_on_repo_name_type` (`repo_name`,`type`),
                                 KEY `index_ge_on_actor_login_type` (`actor_login`,`type`),
                                 KEY `index_ge_on_org_login_type` (`org_login`,`type`),
                                 KEY `index_ge_on_org_id_type` (`org_id`,`type`),
                                 KEY `index_ge_on_repo_name_lower` ((lower(`repo_name`))),
                                 KEY `index_ge_on_repo_id_created_at` (`repo_id`,`created_at`),
                                 KEY `index_ge_on_repo_id_type_action_created_at` (`repo_id`,`type`,`action`,`created_at`,`pr_merged`,`actor_login`,`number`,`push_distinct_size`,`push_size`,`additions`,`deletions`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin
PARTITION BY LIST COLUMNS(`type`)
(PARTITION `push_event` VALUES IN ('PushEvent'),
 PARTITION `create_event` VALUES IN ('CreateEvent'),
 PARTITION `pull_request_event` VALUES IN ('PullRequestEvent'),
 PARTITION `watch_event` VALUES IN ('WatchEvent'),
 PARTITION `issue_comment_event` VALUES IN ('IssueCommentEvent'),
 PARTITION `issues_event` VALUES IN ('IssuesEvent'),
 PARTITION `delete_event` VALUES IN ('DeleteEvent'),
 PARTITION `fork_event` VALUES IN ('ForkEvent'),
 PARTITION `pull_request_review_comment_event` VALUES IN ('PullRequestReviewCommentEvent'),
 PARTITION `pull_request_review_event` VALUES IN ('PullRequestReviewEvent'),
 PARTITION `gollum_event` VALUES IN ('GollumEvent'),
 PARTITION `release_event` VALUES IN ('ReleaseEvent'),
 PARTITION `member_event` VALUES IN ('MemberEvent'),
 PARTITION `commit_comment_event` VALUES IN ('CommitCommentEvent'),
 PARTITION `public_event` VALUES IN ('PublicEvent'),
 PARTITION `gist_event` VALUES IN ('GistEvent'),
 PARTITION `follow_event` VALUES IN ('FollowEvent'),
 PARTITION `event` VALUES IN ('Event'),
 PARTITION `download_event` VALUES IN ('DownloadEvent'),
 PARTITION `team_add_event` VALUES IN ('TeamAddEvent'),
 PARTITION `fork_apply_event` VALUES IN ('ForkApplyEvent'));


CREATE TABLE IF NOT EXISTS  `github_repos` (
                                `repo_id` int(11) NOT NULL,
                                `repo_name` varchar(150) NOT NULL,
                                `owner_id` int(11) NOT NULL,
                                `owner_login` varchar(255) NOT NULL,
                                `owner_is_org` tinyint(1) NOT NULL,
                                `description` varchar(512) NOT NULL DEFAULT '',
                                `primary_language` varchar(32) NOT NULL DEFAULT '',
                                `license` varchar(32) NOT NULL DEFAULT '',
                                `size` bigint(20) NOT NULL DEFAULT '0',
                                `stars` int(11) NOT NULL DEFAULT '0',
                                `forks` int(11) NOT NULL DEFAULT '0',
                                `parent_repo_id` int(11) DEFAULT NULL,
                                `is_fork` tinyint(1) NOT NULL DEFAULT '0',
                                `is_archived` tinyint(1) NOT NULL DEFAULT '0',
                                `is_deleted` tinyint(1) NOT NULL DEFAULT '0',
                                `latest_released_at` timestamp NULL DEFAULT NULL,
                                `pushed_at` timestamp NULL DEFAULT NULL,
                                `created_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
                                `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
                                `last_event_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
                                `refreshed_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
                                PRIMARY KEY (`repo_id`) /*T![clustered_index] CLUSTERED */,
                                KEY `index_gr_on_owner_id` (`owner_id`),
                                KEY `index_gr_on_repo_name` (`repo_name`),
                                KEY `index_gr_on_stars` (`stars`),
                                KEY `index_gr_on_repo_id_repo_name` (`repo_id`,`repo_name`),
                                KEY `index_gr_on_created_at_is_deleted` (`created_at`,`is_deleted`),
                                KEY `index_gr_on_owner_login_owner_id_is_deleted` (`owner_login`,`owner_id`,`is_deleted`),
                                KEY `index_gr_on_name_lower` ((lower(`repo_name`))),
                                KEY `index_gr_on_owner_login_repo_id_created_at` (`owner_login`,`repo_id`,`created_at`)
);

CREATE TABLE IF NOT EXISTS `github_users` (
                                `id` int(11) NOT NULL,
                                `login` varchar(255) NOT NULL,
                                `type` char(3) NOT NULL DEFAULT 'N/A',
                                `is_bot` tinyint(1) NOT NULL DEFAULT '0',
                                `name` varchar(255) NOT NULL DEFAULT '',
                                `email` varchar(255) NOT NULL DEFAULT '',
                                `organization` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
                                `organization_formatted` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
                                `address` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',
                                `country_code` char(3) NOT NULL DEFAULT 'N/A',
                                `region_code` char(3) NOT NULL DEFAULT 'N/A',
                                `state` varchar(255) NOT NULL DEFAULT '',
                                `city` varchar(255) NOT NULL DEFAULT '',
                                `longitude` decimal(11,8) NOT NULL DEFAULT '0',
                                `latitude` decimal(10,8) NOT NULL DEFAULT '0',
                                `public_repos` int(11) NOT NULL DEFAULT '0',
                                `stars_total` int(11) DEFAULT NULL,
                                `participant_total` int(11) DEFAULT NULL,
                                `last_event_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
                                `followers` int(11) NOT NULL DEFAULT '0',
                                `followings` int(11) NOT NULL DEFAULT '0',
                                `created_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
                                `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
                                `is_deleted` tinyint(1) NOT NULL DEFAULT '0',
                                `refreshed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
                                KEY `index_gu_on_login_is_bot_organization_country_code` (`login`,`is_bot`,`organization_formatted`,`country_code`),
                                KEY `index_gu_on_address` (`address`),
                                KEY `index_gu_on_organization` (`organization`),
                                KEY `index_gu_on_login_lower` ((lower(`login`))),
                                KEY `index_gu_on_organization_lower` ((lower(`organization`))),
                                KEY `index_gu_on_country_code_lower` ((lower(`country_code`))),
                                KEY `index_gu_on_country_code_login` (`country_code`,`login`),
                                KEY `idx_gu_on_created_at_type` (`created_at`,`type`)
);

CREATE TABLE IF NOT EXISTS `mv_repo_daily_engagements` (
                                             `repo_id` int(11) NOT NULL,
                                             `day` date NOT NULL,
                                             `user_login` varchar(40) NOT NULL,
                                             `engagements` int(11) DEFAULT NULL,
                                             PRIMARY KEY (`repo_id`,`day`,`user_login`) /*T![clustered_index] CLUSTERED */
);

CREATE TABLE IF NOT EXISTS `coss_repo` (
                             `repo_id` int(11) NOT NULL,
                             `repo_name` varchar(150) NOT NULL,
                             `owner_id` int(11) NOT NULL,
                             `owner_login` varchar(255) NOT NULL,
                             `owner_is_org` tinyint(1) NOT NULL,
                             `description` varchar(512) NOT NULL DEFAULT '',
                             `primary_language` varchar(32) NOT NULL DEFAULT '',
                             `license` varchar(255) DEFAULT NULL,
                             `size` bigint(20) NOT NULL DEFAULT '0',
                             `stars` int(11) NOT NULL DEFAULT '0',
                             `forks` int(11) NOT NULL DEFAULT '0',
                             `watchers` int(11) NOT NULL DEFAULT '0',
                             `parent_repo_id` int(11) DEFAULT NULL,
                             `is_fork` tinyint(1) NOT NULL DEFAULT '0',
                             `is_archived` tinyint(1) NOT NULL DEFAULT '0',
                             `is_deleted` tinyint(1) NOT NULL DEFAULT '0',
                             `latest_released_at` timestamp NULL DEFAULT NULL,
                             `pushed_at` timestamp NULL DEFAULT NULL,
                             `created_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
                             `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
                             PRIMARY KEY (`repo_id`) /*T![clustered_index] CLUSTERED */,
                             KEY `index_gr_on_owner_id` (`owner_id`),
                             KEY `index_gr_on_repo_name` (`repo_name`),
                             KEY `index_gr_on_stars` (`stars`),
                             KEY `index_gr_on_repo_id_repo_name` (`repo_id`,`repo_name`),
                             KEY `index_gr_on_created_at_is_deleted` (`created_at`,`is_deleted`),
                             KEY `index_gr_on_owner_login_owner_id_is_deleted` (`owner_login`,`owner_id`,`is_deleted`)
);

CREATE TABLE IF NOT EXISTS `trending_repos` (
                                  `id` bigint(20) NOT NULL AUTO_INCREMENT,
                                  `repo_name` varchar(255) DEFAULT NULL,
                                  `created_at` datetime DEFAULT NULL,
                                  PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
                                  KEY `index_trending_repos_on_repo_name` (`repo_name`),
                                  KEY `index_trending_repos_on_created_at` (`created_at`),
                                  KEY `index_tr_on_repo_name_lower` ((lower(`repo_name`)))
);

CREATE TABLE IF NOT EXISTS `mv_repo_daily_engagements` (
                                             `repo_id` int(11) NOT NULL,
                                             `day` date NOT NULL,
                                             `user_login` varchar(40) NOT NULL,
                                             `engagements` int(11) DEFAULT NULL,
                                             PRIMARY KEY (`repo_id`,`day`,`user_login`) /*T![clustered_index] CLUSTERED */
);


CREATE TABLE IF NOT EXISTS `github_repo_topics` (
                                      `repo_id` int(11) NOT NULL,
                                      `topic` varchar(50) NOT NULL,
                                      PRIMARY KEY (`repo_id`,`topic`) /*T![clustered_index] CLUSTERED */,
                                      KEY `index_grt_on_topic_lower` ((lower(`topic`))),
                                      KEY `index_ge_on_topic_repo_id` (`topic`,`repo_id`)
);

CREATE TABLE IF NOT EXISTS `stats_api_requests` (
                                      `client_ip` varchar(128) NOT NULL DEFAULT '',
                                      `client_origin` varchar(255) NOT NULL DEFAULT '',
                                      `method` enum('GET','POST','HEAD','PUT','PATCH','DELETE','OPTIONS') NOT NULL,
                                      `path` varchar(255) DEFAULT NULL,
                                      `query` json DEFAULT NULL,
                                      `status_code` int(11) DEFAULT NULL,
                                      `error` tinyint(1) NOT NULL DEFAULT '0',
                                      `is_dev` tinyint(1) NOT NULL DEFAULT '0',
                                      `duration` float DEFAULT NULL,
                                      `finished_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                      KEY `idx_sar_on_finished_at` (`finished_at`),
                                      KEY `idx_sar_on_path` (`path`),
                                      KEY `idx_sar_on_client_ip` (`client_ip`),
                                      KEY `idx_sar_on_client_origin` (`client_origin`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin
PARTITION BY RANGE COLUMNS(`finished_at`)
(PARTITION `P_LT_2023-08-01` VALUES LESS THAN ('2023-08-01'),
 PARTITION `P_LT_2023-09-01` VALUES LESS THAN ('2023-09-01'),
 PARTITION `P_LT_2023-10-01` VALUES LESS THAN ('2023-10-01'),
 PARTITION `P_LT_2023-11-01` VALUES LESS THAN ('2023-11-01'),
 PARTITION `P_LT_2023-12-01` VALUES LESS THAN ('2023-12-01'),
 PARTITION `P_LT_2024-01-01` VALUES LESS THAN ('2024-01-01'),
 PARTITION `P_LT_2024-02-01` VALUES LESS THAN ('2024-02-01'),
 PARTITION `P_LT_2024-03-01` VALUES LESS THAN ('2024-03-01'),
 PARTITION `P_LT_2024-04-01` VALUES LESS THAN ('2024-04-01'),
 PARTITION `P_LT_2024-05-01` VALUES LESS THAN ('2024-05-01'),
 PARTITION `P_LT_2024-06-01` VALUES LESS THAN ('2024-06-01'),
 PARTITION `P_LT_2024-07-01` VALUES LESS THAN ('2024-07-01'),
 PARTITION `P_LT_2024-08-01` VALUES LESS THAN ('2024-08-01'),
 PARTITION `P_LT_2024-09-01` VALUES LESS THAN ('2024-09-01'),
 PARTITION `P_LT_2024-10-01` VALUES LESS THAN ('2024-10-01'),
 PARTITION `P_LT_2024-11-01` VALUES LESS THAN ('2024-11-01'),
 PARTITION `P_LT_2024-12-01` VALUES LESS THAN ('2024-12-01'),
 PARTITION `P_LT_2025-01-01` VALUES LESS THAN ('2025-01-01'));


CREATE TABLE if not exists `archive_access_logs` (
                                       `id` bigint(20) NOT NULL /*T![auto_rand] AUTO_RANDOM(5) */,
                                       `remote_addr` varchar(128) NOT NULL DEFAULT '',
                                       `origin` varchar(128) NOT NULL DEFAULT '',
                                       `status_code` int(11) NOT NULL DEFAULT '0',
                                       `request_path` varchar(1024) DEFAULT NULL,
                                       `request_params` json DEFAULT NULL,
                                       `requested_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                       PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
                                       KEY `index_al_on_requested_at` (`requested_at`)
);


CREATE TABLE if not exists `explorer_questions` (
                                      `id` varbinary(16) NOT NULL,
                                      `hash` varchar(128) NOT NULL,
                                      `user_id` int(11) NOT NULL COMMENT 'The user id of system user',
                                      `status` enum('new','answer_generating','sql_validating','waiting','summarizing','running','success','error','cancel') NOT NULL,
                                      `title` varchar(512) NOT NULL,
                                      `revised_title` varchar(512) DEFAULT NULL,
                                      `not_clear` varchar(512) DEFAULT NULL,
                                      `assumption` text DEFAULT NULL,
                                      `combined_title` varchar(512) DEFAULT NULL,
                                      `sql_can_answer` tinyint(1) DEFAULT NULL,
                                      `query_sql` text DEFAULT NULL,
                                      `query_hash` varchar(128) DEFAULT NULL,
                                      `engines` json DEFAULT NULL,
                                      `plan` json DEFAULT NULL COMMENT 'The execution plan of SQL.',
                                      `queue_name` enum('explorer_high_concurrent_queue','explorer_low_concurrent_queue') DEFAULT NULL,
                                      `queue_job_id` varchar(128) DEFAULT NULL,
                                      `recommended_questions` json DEFAULT NULL,
                                      `result` json DEFAULT NULL,
                                      `chart` json DEFAULT NULL,
                                      `answer` json DEFAULT NULL,
                                      `answer_summary` json DEFAULT NULL,
                                      `batch_job_id` varchar(40) DEFAULT NULL,
                                      `need_review` tinyint(1) NOT NULL DEFAULT '0',
                                      `recommended` tinyint(1) NOT NULL DEFAULT '0',
                                      `hit_cache` tinyint(1) NOT NULL DEFAULT '0',
                                      `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                      `requested_at` datetime DEFAULT CURRENT_TIMESTAMP,
                                      `executed_at` datetime DEFAULT NULL,
                                      `finished_at` datetime DEFAULT NULL,
                                      `spent` float DEFAULT NULL,
                                      `error_type` varchar(30) DEFAULT NULL,
                                      `error` text DEFAULT NULL,
                                      PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
                                      KEY `idx_eq_on_user_id_created_at` (`user_id`,`created_at`)
);

CREATE TABLE IF NOT EXISTS `collections` (
                               `id` bigint(20) NOT NULL AUTO_INCREMENT,
                               `name` varchar(255) NOT NULL,
                               `public` tinyint(1) DEFAULT '1',
                               `past_month_visits` int(11) NOT NULL DEFAULT '0',
                               PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
                               UNIQUE KEY `index_collections_on_name` (`name`)
);

CREATE TABLE IF NOT EXISTS `collection_items` (
                                    `id` bigint(20) NOT NULL AUTO_INCREMENT,
                                    `collection_id` bigint(20) DEFAULT NULL,
                                    `repo_name` varchar(255) NOT NULL,
                                    `repo_id` bigint(20) NOT NULL,
                                    `last_month_rank` int(11) DEFAULT NULL,
                                    `last_2nd_month_rank` int(11) DEFAULT NULL,
                                    `stars_total` int(11) NOT NULL DEFAULT '0',
                                    `pull_requests_total` int(11) NOT NULL DEFAULT '0',
                                    `issues_total` int(11) NOT NULL DEFAULT '0',
                                    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
                                    KEY `index_collection_items_on_collection_id` (`collection_id`),
                                    UNIQUE KEY `index_ci_on_collection_id_repo_id` (`collection_id`,`repo_id`)
);

CREATE TABLE IF NOT EXISTS `cached_table_cache` (
                                      `cache_key` varchar(512) NOT NULL,
                                      `cache_value` json NOT NULL,
                                      `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                      `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                      `expires` int(11) DEFAULT '-1' COMMENT 'cache will expire after n seconds',
                                      `expired_at` datetime GENERATED ALWAYS AS (if(`expires` > 0, date_add(`updated_at`, interval `expires` second), date_add(`updated_at`, interval 99 year))) VIRTUAL,
                                      PRIMARY KEY (`cache_key`) /*T![clustered_index] CLUSTERED */,
                                      KEY `idx_ctc_on_created_at` (`created_at`)
);

CREATE TABLE IF NOT EXISTS `blacklist_users` (
                                   `login` varchar(255) NOT NULL,
                                   UNIQUE KEY `blacklist_users_login_uindex` (`login`),
                                   PRIMARY KEY (`login`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE IF NOT EXISTS `cache` (
                         `cache_key` varchar(512) NOT NULL,
                         `cache_value` json NOT NULL,
                         `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                         `expires` int(11) DEFAULT '-1' COMMENT 'cache will expire after n seconds',
                         `expired_at` datetime GENERATED ALWAYS AS (if(`expires` > 0, date_add(`updated_at`, interval `expires` second), date_add(`updated_at`, interval 99 year))) VIRTUAL,
                         PRIMARY KEY (`cache_key`) /*T![clustered_index] CLUSTERED */
);

CREATE TABLE IF NOT EXISTS `event_logs` (
                              `id` bigint(20) NOT NULL AUTO_INCREMENT,
                              `created_at` datetime NOT NULL,
                              PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
                              KEY `index_event_logs_on_created_at` (`created_at`)
);