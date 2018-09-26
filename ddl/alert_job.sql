CREATE TABLE `alert_job` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'PK',
  `user_id` varchar(11) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '0' COMMENT 'user_id',
  `name` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'name',
  `value` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'yaml',
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'status',
  `is_deleted` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'is_deleted',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'updated_at',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPACT;