set sql_mode='';
create database IF NOT EXISTS gharchive_dev;
use gharchive_dev;


CREATE TABLE `blacklist_repos` (
    `name` varchar(255) DEFAULT NULL
);

CREATE TABLE `blacklist_users` ( 
    `login` varchar(255) NOT NULL,
    UNIQUE KEY `blacklist_users_login_uindex` (`login`),
    PRIMARY KEY (`login`) /*T![clustered_index] NONCLUSTERED */
);

CREATE TABLE `collection_items` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `collection_id` bigint(20) DEFAULT NULL,
    `repo_name` varchar(255) NOT NULL,
    `repo_id` bigint(20) NOT NULL,
    `last_month_rank` int(11) DEFAULT NULL,
    `last_2nd_month_rank` int(11) DEFAULT NULL,
    PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
    KEY `index_collection_items_on_collection_id` (`collection_id`)
);

CREATE TABLE `collections` (
      `id` bigint(20) NOT NULL AUTO_INCREMENT,
      `name` varchar(255) NOT NULL,
      `public` tinyint(1) DEFAULT '1',
      PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
      UNIQUE KEY `index_collections_on_name` (`name`)
);

CREATE TABLE `github_repos` (
`repo_id` int(11) NOT NULL,
`repo_name` varchar(150) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
`owner_id` int(11) DEFAULT NULL,
`owner_login` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
`owner_is_org` tinyint(1) DEFAULT NULL,
`primary_language` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
`license` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
`size` bigint(20) DEFAULT NULL,
`stars` int(11) DEFAULT NULL,
`forks` int(11) DEFAULT NULL,
`parent_repo_id` int(11) DEFAULT NULL,
`is_fork` tinyint(1) NOT NULL DEFAULT '0',
`is_archived` tinyint(1) NOT NULL DEFAULT '0',
`is_deleted` tinyint(1) NOT NULL DEFAULT '0',
`latest_released_at` timestamp NULL DEFAULT NULL,
`pushed_at` timestamp NULL DEFAULT NULL,
`created_at` timestamp NULL DEFAULT NULL,
`updated_at` timestamp NULL DEFAULT NULL,
`refreshed_at` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
`description` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
PRIMARY KEY (`repo_id`) /*T![clustered_index] CLUSTERED */,
KEY `index_owner_on_github_repos` (`owner_login`),
KEY `index_fullname_on_github_repos` (`repo_name`)
);


CREATE TABLE `github_users` (
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
`followers` int(11) NOT NULL DEFAULT '0',
`followings` int(11) NOT NULL DEFAULT '0',
`created_at` timestamp NOT NULL,
`updated_at` timestamp NOT NULL,
`is_deleted` tinyint(1) NOT NULL DEFAULT '0',
`refreshed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`id`) /*T![clustered_index] CLUSTERED */,
KEY `index_gu_on_login_is_bot_organization_country_code` (`login`,`is_bot`,`organization_formatted`,`country_code`),
KEY `index_gu_on_address` (`address`),
KEY `index_gu_on_organization` (`organization`)
);

CREATE TABLE `github_events` (
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
 KEY `index_github_events_on_id` (`id`) /*!80000 INVISIBLE */,
 KEY `index_github_events_on_actor_login` (`actor_login`),
 KEY `index_github_events_on_created_at` (`created_at`),
 KEY `index_github_events_on_repo_name` (`repo_name`),
 KEY `index_github_events_on_repo_id_type_action_month_actor_login` (`repo_id`,`type`,`action`,`event_month`,`actor_login`),
 KEY `index_ge_on_repo_id_type_action_pr_merged_created_at_add_del` (`repo_id`,`type`,`action`,`pr_merged`,`created_at`,`additions`,`deletions`),
 KEY `index_ge_on_creator_id_type_action_merged_created_at_add_del` (`creator_user_id`,`type`,`action`,`pr_merged`,`created_at`,`additions`,`deletions`),
 KEY `index_ge_on_actor_id_type_action_created_at_repo_id_commits` (`actor_id`,`type`,`action`,`created_at`,`repo_id`,`push_distinct_size`),
 KEY `index_ge_on_repo_id_type_action_created_at_number_pdsize_psize` (`repo_id`,`type`,`action`,`created_at`,`number`,`push_distinct_size`,`push_size`),
 KEY `index_ge_on_repo_id_type_action_created_at_actor_login` (`repo_id`,`type`,`action`,`created_at`,`actor_login`)
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

-- alter table blacklist_repos set tiflash replica 1;
-- alter table blacklist_users set tiflash replica 1;
-- alter table collection_items set tiflash replica 1;
-- alter table collections set tiflash replica 1;
-- alter table github_repos set tiflash replica 1;
-- alter table github_users set tiflash replica 1;
-- alter table github_events set tiflash replica 1;