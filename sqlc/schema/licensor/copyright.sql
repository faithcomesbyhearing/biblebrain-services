CREATE TABLE `bible_fileset_copyrights` (
  `hash_id` char(12) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `copyright_date` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `copyright` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `copyright_description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `open_access` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`hash_id`),
  CONSTRAINT `FK_bible_filesets_bible_fileset_copyrights` FOREIGN KEY (`hash_id`) REFERENCES `bible_filesets` (`hash_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE `bible_fileset_copyright_organizations` (
  `hash_id` char(12) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `organization_id` int unsigned NOT NULL,
  `organization_role` int NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`hash_id`,`organization_role`,`organization_id`),
  KEY `FK_org_id` (`organization_id`),
  KEY `FK_org_role` (`organization_role`),
  CONSTRAINT `FK_bible_fileset_copyright_roles_bible_fileset_copyright_organiz` FOREIGN KEY (`organization_role`) REFERENCES `bible_fileset_copyright_roles` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `FK_bible_filesets_bible_fileset_copyright_organizations` FOREIGN KEY (`hash_id`) REFERENCES `bible_filesets` (`hash_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `FK_organizations_bible_fileset_copyright_organizations` FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
