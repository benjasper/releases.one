CREATE TABLE `users` (
	`id` INT NOT NULL AUTO_INCREMENT,
	`username` VARCHAR(255) NOT NULL,
	`refresh_token` VARCHAR(255) NOT NULL,
	PRIMARY KEY (`id`),
	UNIQUE INDEX `username` (`username`)
);

CREATE TABLE `repositories` (
	`id` INT NOT NULL AUTO_INCREMENT,
	`name` VARCHAR(255) NOT NULL,
	`url` VARCHAR(255) NOT NULL,
	`private` BOOLEAN NOT NULL,
	`created_at` DATETIME NOT NULL,
	`updated_at` DATETIME NOT NULL,
	`last_synced_at` DATETIME NOT NULL,
	PRIMARY KEY (`id`)
);

CREATE TABLE `repository_stars` (
	`repository_id` INT NOT NULL,
	`user_id` INT NOT NULL,
	`created_at` DATETIME NOT NULL,
	`updated_at` DATETIME NOT NULL,
	PRIMARY KEY (`repository_id`, `user_id`),
	FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`) ON DELETE CASCADE,
	FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
);

CREATE TABLE `releases` (
	`id` INT NOT NULL AUTO_INCREMENT,
	`repository_id` INT NOT NULL,
	`tag_name` VARCHAR(255) NOT NULL,
	`description` TEXT NOT NULL,
	`is_prerelease` BOOLEAN NOT NULL,
	`released_at` DATETIME NOT NULL,
	`created_at` DATETIME NOT NULL,
	`updated_at` DATETIME NOT NULL,
	PRIMARY KEY (`id`),
	FOREIGN KEY (`repository_id`) REFERENCES `repositories` (`id`) ON DELETE CASCADE
);
