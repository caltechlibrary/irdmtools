DROP TABLE IF EXISTS `document`;
CREATE TABLE `document` (
  `docid` int NOT NULL DEFAULT '0',
  `rev_number` int DEFAULT NULL,
  `eprintid` int DEFAULT NULL,
  `pos` int DEFAULT NULL,
  `format` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `formatdesc` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `language` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `security` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `license` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `main` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `date_embargo_year` smallint DEFAULT NULL,
  `date_embargo_month` smallint DEFAULT NULL,
  `date_embargo_day` smallint DEFAULT NULL,
  `content` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `placement` int DEFAULT NULL,
  `mime_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `media_duration` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `media_audio_codec` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `media_video_codec` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `media_width` int DEFAULT NULL,
  `media_height` int DEFAULT NULL,
  `media_aspect_ratio` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `media_sample_start` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `media_sample_stop` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`docid`),
  KEY `document_rev_number` (`rev_number`),
  KEY `document_eprintid` (`eprintid`),
  KEY `document_pos` (`pos`),
  KEY `document_format` (`format`),
  KEY `document_language` (`language`),
  KEY `document_security` (`security`),
  KEY `document_license` (`license`),
  KEY `document_main` (`main`),
  KEY `document_date_eate_embargo_day` (`date_embargo_year`,`date_embargo_month`,`date_embargo_day`),
  KEY `document_content` (`content`),
  KEY `document_placement_1` (`placement`),
  KEY `document_mime_type_1` (`mime_type`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `document_permission_group`;
CREATE TABLE `document_permission_group` (
  `docid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `permission_group` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`docid`,`pos`),
  KEY `document_permission_group_pos` (`pos`),
  KEY `document_permisermission_group` (`permission_group`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;
DROP TABLE IF EXISTS `document_relation_type`;
CREATE TABLE `document_relation_type` (
  `docid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `relation_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`docid`,`pos`),
  KEY `document_relation_type_pos` (`pos`),
  KEY `document_relatie_relation_type` (`relation_type`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `document_relation_uri`;
CREATE TABLE `document_relation_uri` (
  `docid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `relation_uri` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`docid`,`pos`),
  KEY `document_relation_uri_pos` (`pos`),
  KEY `document_relation_uri_relation_uri_1` (`relation_uri`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint`;
CREATE TABLE `eprint` (
  `eprintid` int NOT NULL DEFAULT '0',
  `rev_number` int DEFAULT NULL,
  `eprint_status` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `userid` int DEFAULT NULL,
  `importid` int DEFAULT NULL,
  `source` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `dir` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `datestamp_year` smallint DEFAULT NULL,
  `datestamp_month` smallint DEFAULT NULL,
  `datestamp_day` smallint DEFAULT NULL,
  `datestamp_hour` smallint DEFAULT NULL,
  `datestamp_minute` smallint DEFAULT NULL,
  `datestamp_second` smallint DEFAULT NULL,
  `lastmod_year` smallint DEFAULT NULL,
  `lastmod_month` smallint DEFAULT NULL,
  `lastmod_day` smallint DEFAULT NULL,
  `lastmod_hour` smallint DEFAULT NULL,
  `lastmod_minute` smallint DEFAULT NULL,
  `lastmod_second` smallint DEFAULT NULL,
  `status_changed_year` smallint DEFAULT NULL,
  `status_changed_month` smallint DEFAULT NULL,
  `status_changed_day` smallint DEFAULT NULL,
  `status_changed_hour` smallint DEFAULT NULL,
  `status_changed_minute` smallint DEFAULT NULL,
  `status_changed_second` smallint DEFAULT NULL,
  `type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `succeeds` int DEFAULT NULL,
  `commentary` int DEFAULT NULL,
  `replacedby` int DEFAULT NULL,
  `metadata_visibility` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `contact_email` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `fileinfo` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `latitude` float DEFAULT NULL,
  `longitude` float DEFAULT NULL,
  `item_issues_count` int DEFAULT NULL,
  `sword_depositor` int DEFAULT NULL,
  `sword_slug` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `title` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `ispublished` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `full_text_status` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `monograph_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `pres_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `keywords` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `note` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `suggestions` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `abstract` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `date_year` smallint DEFAULT NULL,
  `date_month` smallint DEFAULT NULL,
  `date_day` smallint DEFAULT NULL,
  `date_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `series` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `publication` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `volume` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `number` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `publisher` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `place_of_pub` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `edition` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `pagerange` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `pages` int DEFAULT NULL,
  `event_title` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `event_location` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `event_dates` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `event_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `id_number` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `patent_applicant` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `institution` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `department` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `refereed` varchar(5) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `isbn` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `issn` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `book_title` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `official_url` mediumtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `output_media` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `num_pieces` int DEFAULT NULL,
  `composition_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `data_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `pedagogic_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `completion_time` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `task_purpose` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `learning_level` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `rights` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `official_cit` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `doi` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `pmc_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `alt_url` mediumtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `collection` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `errata` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `thesis_degree` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_degree_grantor` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_submitted_date_year` smallint DEFAULT NULL,
  `thesis_submitted_date_month` smallint DEFAULT NULL,
  `thesis_submitted_date_day` smallint DEFAULT NULL,
  `thesis_defense_date_year` smallint DEFAULT NULL,
  `thesis_defense_date_month` smallint DEFAULT NULL,
  `thesis_defense_date_day` smallint DEFAULT NULL,
  `thesis_degree_date_year` smallint DEFAULT NULL,
  `thesis_degree_date_month` smallint DEFAULT NULL,
  `thesis_degree_date_day` smallint DEFAULT NULL,
  `thesis_approved_date_year` smallint DEFAULT NULL,
  `thesis_approved_date_month` smallint DEFAULT NULL,
  `thesis_approved_date_day` smallint DEFAULT NULL,
  `thesis_public_date_year` smallint DEFAULT NULL,
  `thesis_public_date_month` smallint DEFAULT NULL,
  `thesis_public_date_day` smallint DEFAULT NULL,
  `nonsubj_keywords` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `reviewer` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `review_status` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `copyright_statement` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `gradofc_approval_date_year` smallint DEFAULT NULL,
  `gradofc_approval_date_month` smallint DEFAULT NULL,
  `gradofc_approval_date_day` smallint DEFAULT NULL,
  `classification_code` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_awards` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  `edit_lock_user` int DEFAULT NULL,
  `edit_lock_since` int DEFAULT NULL,
  `edit_lock_until` int DEFAULT NULL,
  PRIMARY KEY (`eprintid`),
  KEY `eprint_rev_number` (`rev_number`),
  KEY `eprint_eprint_status` (`eprint_status`),
  KEY `eprint_userid` (`userid`),
  KEY `eprint_importid` (`importid`),
  KEY `eprint_datestamatestamp_second` (`datestamp_year`,`datestamp_month`,`datestamp_day`,`datestamp_hour`,`datestamp_minute`,`datestamp_second`),
  KEY `eprint_lastmod__lastmod_second` (`lastmod_year`,`lastmod_month`,`lastmod_day`,`lastmod_hour`,`lastmod_minute`,`lastmod_second`),
  KEY `eprint_status_c_changed_second` (`status_changed_year`,`status_changed_month`,`status_changed_day`,`status_changed_hour`,`status_changed_minute`,`status_changed_second`),
  KEY `eprint_type` (`type`),
  KEY `eprint_succeeds` (`succeeds`),
  KEY `eprint_replacedby` (`replacedby`),
  KEY `eprint_metadata_visibility` (`metadata_visibility`),
  KEY `eprint_latitude` (`latitude`),
  KEY `eprint_longitude` (`longitude`),
  KEY `eprint_item_issues_count` (`item_issues_count`),
  KEY `eprint_sword_depositor` (`sword_depositor`),
  KEY `eprint_ispublished` (`ispublished`),
  KEY `eprint_full_text_status` (`full_text_status`),
  KEY `eprint_monograph_type` (`monograph_type`),
  KEY `eprint_pres_type` (`pres_type`),
  KEY `eprint_date_yea_month_date_day` (`date_year`,`date_month`,`date_day`),
  KEY `eprint_date_type` (`date_type`),
  KEY `eprint_thesis_type` (`thesis_type`),
  KEY `eprint_pedagogic_type` (`pedagogic_type`),
  KEY `eprint_thesis_degree` (`thesis_degree`),
  KEY `eprint_thesis_smitted_date_day` (`thesis_submitted_date_year`,`thesis_submitted_date_month`,`thesis_submitted_date_day`),
  KEY `eprint_thesis_defense_date_day` (`thesis_defense_date_year`,`thesis_defense_date_month`,`thesis_defense_date_day`),
  KEY `eprint_thesis_ddegree_date_day` (`thesis_degree_date_year`,`thesis_degree_date_month`,`thesis_degree_date_day`),
  KEY `eprint_thesis_aproved_date_day` (`thesis_approved_date_year`,`thesis_approved_date_month`,`thesis_approved_date_day`),
  KEY `eprint_thesis_ppublic_date_day` (`thesis_public_date_year`,`thesis_public_date_month`,`thesis_public_date_day`),
  KEY `eprint_review_status` (`review_status`),
  KEY `eprint_gradofc_proval_date_day` (`gradofc_approval_date_year`,`gradofc_approval_date_month`,`gradofc_approval_date_day`),
  KEY `eprint_contact_email_1` (`contact_email`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_accompaniment`;
CREATE TABLE `eprint_accompaniment` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `accompaniment` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_accompaniment_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_alt_title`;
CREATE TABLE `eprint_alt_title` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `alt_title` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `eprint_conductors_id`;
CREATE TABLE `eprint_conductors_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `conductors_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_conductors_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_conductors_name`;
CREATE TABLE `eprint_conductors_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `conductors_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `conductors_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `conductors_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `conductors_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_conductors_name_pos` (`pos`),
  KEY `eprint_conductors_name_conductors_name_family_1` (`conductors_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_contributors_id`;
CREATE TABLE `eprint_contributors_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `contributors_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_contributors_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_contributors_name`;
CREATE TABLE `eprint_contributors_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `contributors_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `contributors_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `contributors_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `contributors_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_contributors_name_pos` (`pos`),
  KEY `eprint_contributors_name_contributors_name_family_1` (`contributors_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_contributors_type`;
CREATE TABLE `eprint_contributors_type` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `contributors_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_contributors_type_pos` (`pos`),
  KEY `eprint_contribuntributors_type` (`contributors_type`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_copyright_holders`;
CREATE TABLE `eprint_copyright_holders` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `copyright_holders` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_copyright_holders_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_creators_email`;
CREATE TABLE `eprint_creators_email` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `creators_email` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_creators_email_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_creators_id`;
CREATE TABLE `eprint_creators_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `creators_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_creators_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_creators_name`;
CREATE TABLE `eprint_creators_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `creators_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `creators_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `creators_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `creators_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_creators_name_pos` (`pos`),
  KEY `eprint_creators_name_creators_name_family_1` (`creators_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_creators_orcid`;
CREATE TABLE `eprint_creators_orcid` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `creators_orcid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `eprint_creators_show_email`;
CREATE TABLE `eprint_creators_show_email` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `creators_show_email` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_creators_show_email_pos` (`pos`),
  KEY `eprint_creatorstors_show_email` (`creators_show_email`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_divisions`;
CREATE TABLE `eprint_divisions` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `divisions` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_divisions_pos` (`pos`),
  KEY `eprint_divisions_divisions` (`divisions`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_editors_id`;
CREATE TABLE `eprint_editors_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `editors_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_editors_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_editors_name`;
CREATE TABLE `eprint_editors_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `editors_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `editors_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `editors_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `editors_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_editors_name_pos` (`pos`),
  KEY `eprint_editors_name_editors_name_family_1` (`editors_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_editors_orcid`;
CREATE TABLE `eprint_editors_orcid` (
  `eprintid` int NOT NULL,
  `pos` int NOT NULL,
  `editors_orcid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `eprint_exhibitors_id`;
CREATE TABLE `eprint_exhibitors_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `exhibitors_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_exhibitors_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_exhibitors_name`;
CREATE TABLE `eprint_exhibitors_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `exhibitors_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `exhibitors_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `exhibitors_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `exhibitors_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_exhibitors_name_pos` (`pos`),
  KEY `eprint_exhibitors_name_exhibitors_name_family_1` (`exhibitors_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_funders_agency`;
CREATE TABLE `eprint_funders_agency` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `funders_agency` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_funders_agency_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_funders_grant_number`;
CREATE TABLE `eprint_funders_grant_number` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `funders_grant_number` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_funders_rant_number_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_item_issues_comment`;
CREATE TABLE `eprint_item_issues_comment` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `item_issues_comment` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_item_issues_comment_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_item_issues_description`;
CREATE TABLE `eprint_item_issues_description` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `item_issues_description` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_item_issdescription_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_item_issues_id`;
CREATE TABLE `eprint_item_issues_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `item_issues_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_item_issues_id_pos` (`pos`),
  KEY `eprint_item_issues_id_item_issues_id_1` (`item_issues_id`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_item_issues_reported_by`;
CREATE TABLE `eprint_item_issues_reported_by` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `item_issues_reported_by` int DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_item_issreported_by_pos` (`pos`),
  KEY `eprint_item_issues_reported_by` (`item_issues_reported_by`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_item_issues_resolved_by`;
CREATE TABLE `eprint_item_issues_resolved_by` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `item_issues_resolved_by` int DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_item_issresolved_by_pos` (`pos`),
  KEY `eprint_item_issues_resolved_by` (`item_issues_resolved_by`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_item_issues_status`;
CREATE TABLE `eprint_item_issues_status` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `item_issues_status` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_item_issues_status_pos` (`pos`),
  KEY `eprint_item_issm_issues_status` (`item_issues_status`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_item_issues_timestamp`;
CREATE TABLE `eprint_item_issues_timestamp` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `item_issues_timestamp_year` smallint DEFAULT NULL,
  `item_issues_timestamp_month` smallint DEFAULT NULL,
  `item_issues_timestamp_day` smallint DEFAULT NULL,
  `item_issues_timestamp_hour` smallint DEFAULT NULL,
  `item_issues_timestamp_minute` smallint DEFAULT NULL,
  `item_issues_timestamp_second` smallint DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_item_isss_timestamp_pos` (`pos`),
  KEY `eprint_item_issimestamp_second` (`item_issues_timestamp_year`,`item_issues_timestamp_month`,`item_issues_timestamp_day`,`item_issues_timestamp_hour`,`item_issues_timestamp_minute`,`item_issues_timestamp_second`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_item_issues_type`;
CREATE TABLE `eprint_item_issues_type` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `item_issues_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_item_issues_type_pos` (`pos`),
  KEY `eprint_item_issues_type_item_issues_type_1` (`item_issues_type`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_local_group`;
CREATE TABLE `eprint_local_group` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `local_group` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `eprint_lyricists_id`;
CREATE TABLE `eprint_lyricists_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `lyricists_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_lyricists_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_lyricists_name`;
CREATE TABLE `eprint_lyricists_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `lyricists_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `lyricists_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `lyricists_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `lyricists_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_lyricists_name_pos` (`pos`),
  KEY `eprint_lyricists_name_lyricists_name_family_1` (`lyricists_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_option_major`;
CREATE TABLE `eprint_option_major` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `option_major` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_option_major_pos` (`pos`),
  KEY `eprint_option_mor_option_major` (`option_major`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_option_minor`;
CREATE TABLE `eprint_option_minor` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `option_minor` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_option_minor_pos` (`pos`),
  KEY `eprint_option_mor_option_minor` (`option_minor`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_other_numbering_system_id`;
CREATE TABLE `eprint_other_numbering_system_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `other_numbering_system_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_other_nutem_id_eprintid` (`eprintid`),
  KEY `eprint_other_nug_system_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_other_numbering_system_name`;
CREATE TABLE `eprint_other_numbering_system_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `other_numbering_system_name` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_other_num_name_eprintid` (`eprintid`),
  KEY `eprint_other_nusystem_name_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_producers_id`;
CREATE TABLE `eprint_producers_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `producers_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_producers_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_producers_name`;
CREATE TABLE `eprint_producers_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `producers_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `producers_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `producers_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `producers_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_producers_name_pos` (`pos`),
  KEY `eprint_producers_name_producers_name_family_1` (`producers_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_projects`;
CREATE TABLE `eprint_projects` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `projects` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_projects_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_referencetext`;
CREATE TABLE `eprint_referencetext` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `referencetext` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_referencetext_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_related_url_description`;
CREATE TABLE `eprint_related_url_description` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `related_url_description` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `eprint_related_url_type`;
CREATE TABLE `eprint_related_url_type` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `related_url_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_related_url_type_pos` (`pos`),
  KEY `eprint_related_elated_url_type` (`related_url_type`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_related_url_url`;
CREATE TABLE `eprint_related_url_url` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `related_url_url` mediumtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_related_url_url_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_relation_type`;
CREATE TABLE `eprint_relation_type` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `relation_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_relation_type_pos` (`pos`),
  KEY `eprint_relatione_relation_type` (`relation_type`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_relation_uri`;
CREATE TABLE `eprint_relation_uri` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `relation_uri` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_relation_uri_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_skill_areas`;
CREATE TABLE `eprint_skill_areas` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `skill_areas` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_skill_areas_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_subjects`;
CREATE TABLE `eprint_subjects` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `subjects` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_subjects_pos` (`pos`),
  KEY `eprint_subjects_subjects` (`subjects`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_thesis_advisor_email`;
CREATE TABLE `eprint_thesis_advisor_email` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `thesis_advisor_email` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_thesis_avisor_email_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_thesis_advisor_id`;
CREATE TABLE `eprint_thesis_advisor_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `thesis_advisor_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_thesis_advisor_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_thesis_advisor_name`;
CREATE TABLE `eprint_thesis_advisor_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `thesis_advisor_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_advisor_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_advisor_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_advisor_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_thesis_advisor_name_pos` (`pos`),
  KEY `eprint_thesis_advisor_name_thesis_advisor_name_family_1` (`thesis_advisor_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_thesis_advisor_orcid`;
CREATE TABLE `eprint_thesis_advisor_orcid` (
  `eprintid` int NOT NULL,
  `pos` int NOT NULL,
  `thesis_advisor_orcid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `eprint_thesis_advisor_role`;
CREATE TABLE `eprint_thesis_advisor_role` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `thesis_advisor_role` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_thesis_advisor_role_pos` (`pos`),
  KEY `eprint_thesis_ais_advisor_role` (`thesis_advisor_role`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_thesis_committee_email`;
CREATE TABLE `eprint_thesis_committee_email` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `thesis_committee_email` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_thesis_cittee_email_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_thesis_committee_id`;
CREATE TABLE `eprint_thesis_committee_id` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `thesis_committee_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_thesis_committee_id_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_thesis_committee_name`;
CREATE TABLE `eprint_thesis_committee_name` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `thesis_committee_name_honourific` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_committee_name_given` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_committee_name_family` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `thesis_committee_name_lineage` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_thesis_cmittee_name_pos` (`pos`),
  KEY `eprint_thesis_committee_name_thesis_committee_name_family_1` (`thesis_committee_name_family`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `eprint_thesis_committee_orcid`;
CREATE TABLE `eprint_thesis_committee_orcid` (
  `eprintid` int NOT NULL,
  `pos` int NOT NULL,
  `thesis_committee_orcid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `eprint_thesis_committee_role`;
CREATE TABLE `eprint_thesis_committee_role` (
  `eprintid` int NOT NULL DEFAULT '0',
  `pos` int NOT NULL DEFAULT '0',
  `thesis_committee_role` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`eprintid`,`pos`),
  KEY `eprint_thesis_cmittee_role_pos` (`pos`),
  KEY `eprint_thesis_c_committee_role` (`thesis_committee_role`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `file`;
CREATE TABLE `file` (
  `fileid` int NOT NULL,
  `datasetid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `objectid` int DEFAULT NULL,
  `filename` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `mime_type` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `hash` varchar(64) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `hash_type` varchar(32) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  `filesize` bigint DEFAULT NULL,
  `mtime_year` smallint DEFAULT NULL,
  `mtime_month` smallint DEFAULT NULL,
  `mtime_day` smallint DEFAULT NULL,
  `mtime_hour` smallint DEFAULT NULL,
  `mtime_minute` smallint DEFAULT NULL,
  `mtime_second` smallint DEFAULT NULL,
  PRIMARY KEY (`fileid`),
  KEY `file_datasetid_1` (`datasetid`),
  KEY `file_objectid_1` (`objectid`),
  KEY `file_filename_1` (`filename`),
  KEY `file_hash_1` (`hash`),
  KEY `file_hash_type_1` (`hash_type`),
  KEY `file_mtime_year_6` (`mtime_year`,`mtime_month`,`mtime_day`,`mtime_hour`,`mtime_minute`,`mtime_second`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `file_copies_pluginid`;
CREATE TABLE `file_copies_pluginid` (
  `fileid` int NOT NULL,
  `pos` int NOT NULL,
  `copies_pluginid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`fileid`,`pos`),
  KEY `file_copies_pluginid_copies_pluginid_1` (`copies_pluginid`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `file_copies_sourceid`;
CREATE TABLE `file_copies_sourceid` (
  `fileid` int NOT NULL,
  `pos` int NOT NULL,
  `copies_sourceid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`fileid`,`pos`),
  KEY `file_copies_sourceid_copies_sourceid_1` (`copies_sourceid`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `subject`;
CREATE TABLE `subject` (
  `subjectid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL DEFAULT '',
  `rev_number` int DEFAULT NULL,
  `depositable` varchar(5) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`subjectid`),
  KEY `subject_rev_number` (`rev_number`),
  KEY `subject_depositable` (`depositable`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `subject_ancestors`;
CREATE TABLE `subject_ancestors` (
  `subjectid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL DEFAULT '',
  `pos` int NOT NULL DEFAULT '0',
  `ancestors` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`subjectid`,`pos`),
  KEY `subject_ancestors_pos` (`pos`),
  KEY `subject_ancestors_ancestors_1` (`ancestors`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `subject_name_lang`;
CREATE TABLE `subject_name_lang` (
  `subjectid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL DEFAULT '',
  `pos` int NOT NULL DEFAULT '0',
  `name_lang` varchar(16) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`subjectid`,`pos`),
  KEY `subject_name_lang_pos` (`pos`),
  KEY `subject_name_lang_name_lang` (`name_lang`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `subject_name_name`;
CREATE TABLE `subject_name_name` (
  `subjectid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL DEFAULT '',
  `pos` int NOT NULL DEFAULT '0',
  `name_name` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`subjectid`,`pos`),
  KEY `subject_name_name_pos` (`pos`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `subject_name_sortvalue`;
CREATE TABLE `subject_name_sortvalue` (
  `subjectid` varchar(128) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL,
  `pos` int NOT NULL,
  `name_sortvalue` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`subjectid`,`pos`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `subject_parents`;
CREATE TABLE `subject_parents` (
  `subjectid` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL DEFAULT '',
  `pos` int NOT NULL DEFAULT '0',
  `parents` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`subjectid`,`pos`),
  KEY `subject_parents_pos` (`pos`),
  KEY `subject_parents_parents_1` (`parents`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_bin;
DROP TABLE IF EXISTS `subject_sortvalue_lang`;
CREATE TABLE `subject_sortvalue_lang` (
  `subjectid` varchar(128) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL,
  `pos` int NOT NULL,
  `sortvalue_lang` varchar(16) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`subjectid`,`pos`),
  KEY `sortvalue_lang_0` (`sortvalue_lang`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
DROP TABLE IF EXISTS `subject_sortvalue_sortvalue`;
CREATE TABLE `subject_sortvalue_sortvalue` (
  `subjectid` varchar(128) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL,
  `pos` int NOT NULL,
  `sortvalue_sortvalue` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin DEFAULT NULL,
  PRIMARY KEY (`subjectid`,`pos`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3;
