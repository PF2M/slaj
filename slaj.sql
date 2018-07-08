-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Server version:               8.0.11 - MySQL Community Server - GPL
-- Server OS:                    Win64
-- HeidiSQL Version:             9.5.0.5196
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;

-- Dumping structure for table slaj.comments
CREATE TABLE IF NOT EXISTS `comments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created_by` int(11) NOT NULL,
  `community_id` int(11) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `post` int(11) NOT NULL,
  `body` varchar(2000) NOT NULL,
  `image` tinytext,
  `is_spoiler` tinyint(1) NOT NULL DEFAULT '0',
  `is_rm` tinyint(1) NOT NULL DEFAULT '0',
  `is_rm_by_admin` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table slaj.comments: ~0 rows (approximately)
/*!40000 ALTER TABLE `comments` DISABLE KEYS */;
/*!40000 ALTER TABLE `comments` ENABLE KEYS */;

-- Dumping structure for table slaj.communities
CREATE TABLE IF NOT EXISTS `communities` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(64) NOT NULL,
  `description` varchar(2000) NOT NULL,
  `icon` tinytext NOT NULL,
  `banner` tinytext NOT NULL,
  `is_featured` tinyint(1) NOT NULL,
  `developer_only` tinyint(1) NOT NULL DEFAULT '0',
  `staff_only` tinyint(1) NOT NULL DEFAULT '0',
  `rm` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table slaj.communities: ~0 rows (approximately)
/*!40000 ALTER TABLE `communities` DISABLE KEYS */;
INSERT INTO `communities` (`id`, `title`, `description`, `icon`, `banner`, `is_featured`, `developer_only`, `staff_only`, `rm`) VALUES
	(1, 'General Discussion Community', 'who cares lmao', 'https://kek.gg/i/5k4jmK.png', 'https://kek.gg/i/3qs7b9.png', 1, 0, 0, 0);
/*!40000 ALTER TABLE `communities` ENABLE KEYS */;

-- Dumping structure for table slaj.posts
CREATE TABLE IF NOT EXISTS `posts` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `created_by` int(11) NOT NULL,
  `community_id` int(11) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `body` varchar(2000) NOT NULL,
  `image` tinytext,
  `url` tinytext,
  `is_spoiler` tinyint(1) NOT NULL DEFAULT '0',
  `is_rm` tinyint(1) NOT NULL DEFAULT '0',
  `is_rm_by_admin` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table slaj.posts: ~19 rows (approximately)
/*!40000 ALTER TABLE `posts` DISABLE KEYS */;
INSERT INTO `posts` (`id`, `created_by`, `community_id`, `created_at`, `body`, `image`, `url`, `is_spoiler`, `is_rm`, `is_rm_by_admin`) VALUES
	(1, 1, 1, '2018-07-07 21:17:39', 'this is a post ', NULL, NULL, 0, 0, 0);
/*!40000 ALTER TABLE `posts` ENABLE KEYS */;

-- Dumping structure for table slaj.profiles
CREATE TABLE IF NOT EXISTS `profiles` (
  `user` int(11) NOT NULL AUTO_INCREMENT,
  `comment` text COLLATE utf8mb4_bin NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `nnid` varchar(16) COLLATE utf8mb4_bin NOT NULL,
  `region` varchar(64) COLLATE utf8mb4_bin NOT NULL,
  `gender` int(1) NOT NULL,
  `nnid_visibility` tinyint(1) NOT NULL,
  `yeah_visibility` tinyint(1) NOT NULL,
  `reply_visibility` tinyint(1) NOT NULL,
  PRIMARY KEY (`user`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Dumping data for table slaj.profiles: ~0 rows (approximately)
/*!40000 ALTER TABLE `profiles` DISABLE KEYS */;
/*!40000 ALTER TABLE `profiles` ENABLE KEYS */;

-- Dumping structure for table slaj.users
CREATE TABLE IF NOT EXISTS `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(32) COLLATE utf8mb4_bin NOT NULL,
  `nickname` varchar(32) COLLATE utf8mb4_bin NOT NULL,
  `avatar` tinytext COLLATE utf8mb4_bin NOT NULL,
  `email` tinytext COLLATE utf8mb4_bin,
  `password` varchar(75) COLLATE utf8mb4_bin NOT NULL,
  `ip` varchar(39) COLLATE utf8mb4_bin NOT NULL,
  `level` int(2) NOT NULL,
  `role` int(11) NOT NULL,
  `last_seen` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `color` varchar(7) COLLATE utf8mb4_bin NOT NULL,
  `yeah_notifications` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- Dumping data for table slaj.users: 2 rows
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` (`id`, `username`, `nickname`, `avatar`, `email`, `password`, `ip`, `level`, `role`, `last_seen`, `color`, `yeah_notifications`) VALUES
	(1, 'cedar', 'ok', 'https://cdn.discordapp.com/attachments/462129189152096277/464848128886046740/JPEG_20180706_193945.jpg', 'cedar', '$2a$10$6SSZNsK6aCAspCx.0tfgXO9uYn3YbD2GSaYId4ilwLDmHq1a70tYu', '', 0, 0, '2018-07-08 00:49:30', '', 1);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;

/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IF(@OLD_FOREIGN_KEY_CHECKS IS NULL, 1, @OLD_FOREIGN_KEY_CHECKS) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
