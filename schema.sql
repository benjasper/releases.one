-- Create "repositories" table
CREATE TABLE `repositories` (
  `id` int NOT NULL AUTO_INCREMENT,
  `github_id` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `url` varchar(255) NOT NULL,
  `private` bool NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `last_synced_at` datetime NOT NULL,
  `image_url` varchar(255) NOT NULL,
  `image_size` int NOT NULL,
  `hash` bigint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `github_id` (`github_id`)
);

-- Create "releases" table
CREATE TABLE `releases` (
  `github_id` varchar(255) NOT NULL,
  `id` int NOT NULL AUTO_INCREMENT,
  `repository_id` int NOT NULL,
  `name` varchar(255) NOT NULL,
  `url` varchar(255) NOT NULL,
  `tag_name` varchar(255) NOT NULL,
  `description` longtext NOT NULL,
  `description_short` text NOT NULL,
  `author` varchar(255) NULL,
  `is_prerelease` bool NOT NULL,
  `released_at` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `hash` bigint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `repository_id` (`repository_id`),
  CONSTRAINT `releases_ibfk_1` FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);

-- Create "users" table
CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `github_id` bigint unsigned NOT NULL,
  `github_token` json NOT NULL,
  `last_synced_at` datetime NOT NULL,
  `public_id` varchar(255) NOT NULL,
  `is_onboarded` bool NOT NULL,
  `is_public` bool NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `github_id` (`github_id`),
  INDEX `public_id` (`public_id`)
);

-- Create "repository_stars" table
CREATE TABLE `repository_stars` (
  `repository_id` int NOT NULL,
  `user_id` int NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `type` tinyint NOT NULL,
  PRIMARY KEY (`repository_id`, `user_id`),
  INDEX `user_id` (`user_id`),
  CONSTRAINT `repository_stars_ibfk_1` FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `repository_stars_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
