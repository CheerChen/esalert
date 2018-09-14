CREATE TABLE `alert_job` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'PK',
  `name` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'name',
  `value` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'config(yaml)',
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'on/off',
  `is_deleted` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'is_deleted',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'updated_at',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPACT;
