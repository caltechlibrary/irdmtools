-- eprints_population_report.sql
--
-- A fourth counts-only report for the CaltechTHESIS EPrints database, generated
-- from the table structure in testdata/eprints_thesis_schema.sql. It answers one
-- question for the gap table: which EPrints columns and list tables are actually
-- populated for thesis eprints. A column counts as populated when it is neither
-- NULL nor an empty string. Only counts are selected, never a value. Columns and
-- tables whose names the runner's linter forbids (email addresses, family and
-- given names, file names) are left out on purpose; the gap table treats them as
-- identifying data. Same runner as the others:
--
--   SHAPE_SQL=eprints_population_report.sql ./eprints_shape_report.bash

-- SECTION: Populated columns of the eprint table, thesis eprints only
SELECT 'eprintid' AS eprint_column, SUM(`eprintid` IS NOT NULL AND CAST(`eprintid` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'rev_number' AS eprint_column, SUM(`rev_number` IS NOT NULL AND CAST(`rev_number` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'eprint_status' AS eprint_column, SUM(`eprint_status` IS NOT NULL AND CAST(`eprint_status` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'userid' AS eprint_column, SUM(`userid` IS NOT NULL AND CAST(`userid` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'importid' AS eprint_column, SUM(`importid` IS NOT NULL AND CAST(`importid` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'source' AS eprint_column, SUM(`source` IS NOT NULL AND CAST(`source` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'dir' AS eprint_column, SUM(`dir` IS NOT NULL AND CAST(`dir` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'datestamp (date, by its year column)' AS eprint_column, SUM(`datestamp_year` IS NOT NULL AND CAST(`datestamp_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'lastmod (date, by its year column)' AS eprint_column, SUM(`lastmod_year` IS NOT NULL AND CAST(`lastmod_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'status_changed (date, by its year column)' AS eprint_column, SUM(`status_changed_year` IS NOT NULL AND CAST(`status_changed_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'type' AS eprint_column, SUM(`type` IS NOT NULL AND CAST(`type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'succeeds' AS eprint_column, SUM(`succeeds` IS NOT NULL AND CAST(`succeeds` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'commentary' AS eprint_column, SUM(`commentary` IS NOT NULL AND CAST(`commentary` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'replacedby' AS eprint_column, SUM(`replacedby` IS NOT NULL AND CAST(`replacedby` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'metadata_visibility' AS eprint_column, SUM(`metadata_visibility` IS NOT NULL AND CAST(`metadata_visibility` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'fileinfo' AS eprint_column, SUM(`fileinfo` IS NOT NULL AND CAST(`fileinfo` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'latitude' AS eprint_column, SUM(`latitude` IS NOT NULL AND CAST(`latitude` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'longitude' AS eprint_column, SUM(`longitude` IS NOT NULL AND CAST(`longitude` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'item_issues_count' AS eprint_column, SUM(`item_issues_count` IS NOT NULL AND CAST(`item_issues_count` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'sword_depositor' AS eprint_column, SUM(`sword_depositor` IS NOT NULL AND CAST(`sword_depositor` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'sword_slug' AS eprint_column, SUM(`sword_slug` IS NOT NULL AND CAST(`sword_slug` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'title' AS eprint_column, SUM(`title` IS NOT NULL AND CAST(`title` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'ispublished' AS eprint_column, SUM(`ispublished` IS NOT NULL AND CAST(`ispublished` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'full_text_status' AS eprint_column, SUM(`full_text_status` IS NOT NULL AND CAST(`full_text_status` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'monograph_type' AS eprint_column, SUM(`monograph_type` IS NOT NULL AND CAST(`monograph_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'pres_type' AS eprint_column, SUM(`pres_type` IS NOT NULL AND CAST(`pres_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'keywords' AS eprint_column, SUM(`keywords` IS NOT NULL AND CAST(`keywords` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'note' AS eprint_column, SUM(`note` IS NOT NULL AND CAST(`note` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'suggestions' AS eprint_column, SUM(`suggestions` IS NOT NULL AND CAST(`suggestions` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'abstract' AS eprint_column, SUM(`abstract` IS NOT NULL AND CAST(`abstract` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'date (date, by its year column)' AS eprint_column, SUM(`date_year` IS NOT NULL AND CAST(`date_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'date_type' AS eprint_column, SUM(`date_type` IS NOT NULL AND CAST(`date_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'series' AS eprint_column, SUM(`series` IS NOT NULL AND CAST(`series` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'publication' AS eprint_column, SUM(`publication` IS NOT NULL AND CAST(`publication` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'volume' AS eprint_column, SUM(`volume` IS NOT NULL AND CAST(`volume` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'number' AS eprint_column, SUM(`number` IS NOT NULL AND CAST(`number` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'publisher' AS eprint_column, SUM(`publisher` IS NOT NULL AND CAST(`publisher` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'place_of_pub' AS eprint_column, SUM(`place_of_pub` IS NOT NULL AND CAST(`place_of_pub` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'edition' AS eprint_column, SUM(`edition` IS NOT NULL AND CAST(`edition` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'pagerange' AS eprint_column, SUM(`pagerange` IS NOT NULL AND CAST(`pagerange` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'pages' AS eprint_column, SUM(`pages` IS NOT NULL AND CAST(`pages` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'event_title' AS eprint_column, SUM(`event_title` IS NOT NULL AND CAST(`event_title` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'event_location' AS eprint_column, SUM(`event_location` IS NOT NULL AND CAST(`event_location` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'event_dates' AS eprint_column, SUM(`event_dates` IS NOT NULL AND CAST(`event_dates` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'event_type' AS eprint_column, SUM(`event_type` IS NOT NULL AND CAST(`event_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'id_number' AS eprint_column, SUM(`id_number` IS NOT NULL AND CAST(`id_number` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'patent_applicant' AS eprint_column, SUM(`patent_applicant` IS NOT NULL AND CAST(`patent_applicant` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'institution' AS eprint_column, SUM(`institution` IS NOT NULL AND CAST(`institution` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'department' AS eprint_column, SUM(`department` IS NOT NULL AND CAST(`department` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_type' AS eprint_column, SUM(`thesis_type` IS NOT NULL AND CAST(`thesis_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'refereed' AS eprint_column, SUM(`refereed` IS NOT NULL AND CAST(`refereed` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'isbn' AS eprint_column, SUM(`isbn` IS NOT NULL AND CAST(`isbn` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'issn' AS eprint_column, SUM(`issn` IS NOT NULL AND CAST(`issn` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'book_title' AS eprint_column, SUM(`book_title` IS NOT NULL AND CAST(`book_title` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'official_url' AS eprint_column, SUM(`official_url` IS NOT NULL AND CAST(`official_url` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'output_media' AS eprint_column, SUM(`output_media` IS NOT NULL AND CAST(`output_media` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'num_pieces' AS eprint_column, SUM(`num_pieces` IS NOT NULL AND CAST(`num_pieces` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'composition_type' AS eprint_column, SUM(`composition_type` IS NOT NULL AND CAST(`composition_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'data_type' AS eprint_column, SUM(`data_type` IS NOT NULL AND CAST(`data_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'pedagogic_type' AS eprint_column, SUM(`pedagogic_type` IS NOT NULL AND CAST(`pedagogic_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'completion_time' AS eprint_column, SUM(`completion_time` IS NOT NULL AND CAST(`completion_time` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'task_purpose' AS eprint_column, SUM(`task_purpose` IS NOT NULL AND CAST(`task_purpose` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'learning_level' AS eprint_column, SUM(`learning_level` IS NOT NULL AND CAST(`learning_level` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'rights' AS eprint_column, SUM(`rights` IS NOT NULL AND CAST(`rights` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'official_cit' AS eprint_column, SUM(`official_cit` IS NOT NULL AND CAST(`official_cit` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'doi' AS eprint_column, SUM(`doi` IS NOT NULL AND CAST(`doi` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'pmc_id' AS eprint_column, SUM(`pmc_id` IS NOT NULL AND CAST(`pmc_id` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'alt_url' AS eprint_column, SUM(`alt_url` IS NOT NULL AND CAST(`alt_url` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'collection' AS eprint_column, SUM(`collection` IS NOT NULL AND CAST(`collection` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'errata' AS eprint_column, SUM(`errata` IS NOT NULL AND CAST(`errata` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_degree' AS eprint_column, SUM(`thesis_degree` IS NOT NULL AND CAST(`thesis_degree` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_degree_grantor' AS eprint_column, SUM(`thesis_degree_grantor` IS NOT NULL AND CAST(`thesis_degree_grantor` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_submitted_date (date, by its year column)' AS eprint_column, SUM(`thesis_submitted_date_year` IS NOT NULL AND CAST(`thesis_submitted_date_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_defense_date (date, by its year column)' AS eprint_column, SUM(`thesis_defense_date_year` IS NOT NULL AND CAST(`thesis_defense_date_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_degree_date (date, by its year column)' AS eprint_column, SUM(`thesis_degree_date_year` IS NOT NULL AND CAST(`thesis_degree_date_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_approved_date (date, by its year column)' AS eprint_column, SUM(`thesis_approved_date_year` IS NOT NULL AND CAST(`thesis_approved_date_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_public_date (date, by its year column)' AS eprint_column, SUM(`thesis_public_date_year` IS NOT NULL AND CAST(`thesis_public_date_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'nonsubj_keywords' AS eprint_column, SUM(`nonsubj_keywords` IS NOT NULL AND CAST(`nonsubj_keywords` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'reviewer' AS eprint_column, SUM(`reviewer` IS NOT NULL AND CAST(`reviewer` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'review_status' AS eprint_column, SUM(`review_status` IS NOT NULL AND CAST(`review_status` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'copyright_statement' AS eprint_column, SUM(`copyright_statement` IS NOT NULL AND CAST(`copyright_statement` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'gradofc_approval_date (date, by its year column)' AS eprint_column, SUM(`gradofc_approval_date_year` IS NOT NULL AND CAST(`gradofc_approval_date_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'classification_code' AS eprint_column, SUM(`classification_code` IS NOT NULL AND CAST(`classification_code` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'thesis_awards' AS eprint_column, SUM(`thesis_awards` IS NOT NULL AND CAST(`thesis_awards` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'edit_lock_user' AS eprint_column, SUM(`edit_lock_user` IS NOT NULL AND CAST(`edit_lock_user` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'edit_lock_since' AS eprint_column, SUM(`edit_lock_since` IS NOT NULL AND CAST(`edit_lock_since` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis'
UNION ALL
SELECT 'edit_lock_until' AS eprint_column, SUM(`edit_lock_until` IS NOT NULL AND CAST(`edit_lock_until` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_thesis_eprints FROM eprint WHERE type = 'thesis';

-- SECTION: List tables of eprint, thesis eprints only (rows, eprints with a row, rows that are not blank)
SELECT 'eprint_accompaniment' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`accompaniment` IS NOT NULL AND CAST(t.`accompaniment` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_accompaniment t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_alt_title' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`alt_title` IS NOT NULL AND CAST(t.`alt_title` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_alt_title t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_conductors_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`conductors_id` IS NOT NULL AND CAST(t.`conductors_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_conductors_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_conductors_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_conductors_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_contributors_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`contributors_id` IS NOT NULL AND CAST(t.`contributors_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_contributors_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_contributors_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_contributors_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_contributors_type' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`contributors_type` IS NOT NULL AND CAST(t.`contributors_type` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_contributors_type t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_copyright_holders' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`copyright_holders` IS NOT NULL AND CAST(t.`copyright_holders` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_copyright_holders t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_creators_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`creators_id` IS NOT NULL AND CAST(t.`creators_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_creators_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_creators_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_creators_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_creators_orcid' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`creators_orcid` IS NOT NULL AND CAST(t.`creators_orcid` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_creators_orcid t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_divisions' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`divisions` IS NOT NULL AND CAST(t.`divisions` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_divisions t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_editors_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`editors_id` IS NOT NULL AND CAST(t.`editors_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_editors_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_editors_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_editors_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_editors_orcid' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`editors_orcid` IS NOT NULL AND CAST(t.`editors_orcid` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_editors_orcid t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_exhibitors_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`exhibitors_id` IS NOT NULL AND CAST(t.`exhibitors_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_exhibitors_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_exhibitors_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_exhibitors_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_funders_agency' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`funders_agency` IS NOT NULL AND CAST(t.`funders_agency` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_funders_agency t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_funders_grant_number' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`funders_grant_number` IS NOT NULL AND CAST(t.`funders_grant_number` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_funders_grant_number t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_item_issues_comment' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`item_issues_comment` IS NOT NULL AND CAST(t.`item_issues_comment` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_item_issues_comment t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_item_issues_description' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`item_issues_description` IS NOT NULL AND CAST(t.`item_issues_description` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_item_issues_description t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_item_issues_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`item_issues_id` IS NOT NULL AND CAST(t.`item_issues_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_item_issues_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_item_issues_reported_by' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`item_issues_reported_by` IS NOT NULL AND CAST(t.`item_issues_reported_by` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_item_issues_reported_by t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_item_issues_resolved_by' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`item_issues_resolved_by` IS NOT NULL AND CAST(t.`item_issues_resolved_by` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_item_issues_resolved_by t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_item_issues_status' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`item_issues_status` IS NOT NULL AND CAST(t.`item_issues_status` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_item_issues_status t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_item_issues_timestamp' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_item_issues_timestamp t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_item_issues_type' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`item_issues_type` IS NOT NULL AND CAST(t.`item_issues_type` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_item_issues_type t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_local_group' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`local_group` IS NOT NULL AND CAST(t.`local_group` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_local_group t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_lyricists_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`lyricists_id` IS NOT NULL AND CAST(t.`lyricists_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_lyricists_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_lyricists_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_lyricists_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_option_major' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`option_major` IS NOT NULL AND CAST(t.`option_major` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_option_major t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_option_minor' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`option_minor` IS NOT NULL AND CAST(t.`option_minor` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_option_minor t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_other_numbering_system_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`other_numbering_system_id` IS NOT NULL AND CAST(t.`other_numbering_system_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_other_numbering_system_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_other_numbering_system_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`other_numbering_system_name` IS NOT NULL AND CAST(t.`other_numbering_system_name` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_other_numbering_system_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_producers_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`producers_id` IS NOT NULL AND CAST(t.`producers_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_producers_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_producers_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_producers_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_projects' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`projects` IS NOT NULL AND CAST(t.`projects` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_projects t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_referencetext' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`referencetext` IS NOT NULL AND CAST(t.`referencetext` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_referencetext t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_related_url_description' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`related_url_description` IS NOT NULL AND CAST(t.`related_url_description` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_related_url_description t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_related_url_type' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`related_url_type` IS NOT NULL AND CAST(t.`related_url_type` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_related_url_type t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_related_url_url' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`related_url_url` IS NOT NULL AND CAST(t.`related_url_url` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_related_url_url t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_relation_type' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`relation_type` IS NOT NULL AND CAST(t.`relation_type` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_relation_type t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_relation_uri' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`relation_uri` IS NOT NULL AND CAST(t.`relation_uri` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_relation_uri t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_skill_areas' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`skill_areas` IS NOT NULL AND CAST(t.`skill_areas` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_skill_areas t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_subjects' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`subjects` IS NOT NULL AND CAST(t.`subjects` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_subjects t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_thesis_advisor_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`thesis_advisor_id` IS NOT NULL AND CAST(t.`thesis_advisor_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_thesis_advisor_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_thesis_advisor_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_thesis_advisor_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_thesis_advisor_orcid' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`thesis_advisor_orcid` IS NOT NULL AND CAST(t.`thesis_advisor_orcid` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_thesis_advisor_orcid t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_thesis_advisor_role' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`thesis_advisor_role` IS NOT NULL AND CAST(t.`thesis_advisor_role` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_thesis_advisor_role t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_thesis_committee_id' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`thesis_committee_id` IS NOT NULL AND CAST(t.`thesis_committee_id` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_thesis_committee_id t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_thesis_committee_name' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, NULL AS n_nonblank_rows FROM eprint_thesis_committee_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_thesis_committee_orcid' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`thesis_committee_orcid` IS NOT NULL AND CAST(t.`thesis_committee_orcid` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_thesis_committee_orcid t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprint_thesis_committee_role' AS list_table, COUNT(*) AS n_rows, COUNT(DISTINCT t.eprintid) AS n_eprints, SUM(t.`thesis_committee_role` IS NOT NULL AND CAST(t.`thesis_committee_role` AS CHAR) <> '') AS n_nonblank_rows FROM eprint_thesis_committee_role t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis';

-- SECTION: Populated columns of the document table, documents of thesis eprints
SELECT 'docid' AS document_column, SUM(d.`docid` IS NOT NULL AND CAST(d.`docid` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'rev_number' AS document_column, SUM(d.`rev_number` IS NOT NULL AND CAST(d.`rev_number` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'eprintid' AS document_column, SUM(d.`eprintid` IS NOT NULL AND CAST(d.`eprintid` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'pos' AS document_column, SUM(d.`pos` IS NOT NULL AND CAST(d.`pos` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'format' AS document_column, SUM(d.`format` IS NOT NULL AND CAST(d.`format` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'formatdesc' AS document_column, SUM(d.`formatdesc` IS NOT NULL AND CAST(d.`formatdesc` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'language' AS document_column, SUM(d.`language` IS NOT NULL AND CAST(d.`language` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'security' AS document_column, SUM(d.`security` IS NOT NULL AND CAST(d.`security` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'license' AS document_column, SUM(d.`license` IS NOT NULL AND CAST(d.`license` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'main' AS document_column, SUM(d.`main` IS NOT NULL AND CAST(d.`main` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'date_embargo (date, by its year column)' AS document_column, SUM(d.`date_embargo_year` IS NOT NULL AND CAST(d.`date_embargo_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'content' AS document_column, SUM(d.`content` IS NOT NULL AND CAST(d.`content` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'placement' AS document_column, SUM(d.`placement` IS NOT NULL AND CAST(d.`placement` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'mime_type' AS document_column, SUM(d.`mime_type` IS NOT NULL AND CAST(d.`mime_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'media_duration' AS document_column, SUM(d.`media_duration` IS NOT NULL AND CAST(d.`media_duration` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'media_audio_codec' AS document_column, SUM(d.`media_audio_codec` IS NOT NULL AND CAST(d.`media_audio_codec` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'media_video_codec' AS document_column, SUM(d.`media_video_codec` IS NOT NULL AND CAST(d.`media_video_codec` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'media_width' AS document_column, SUM(d.`media_width` IS NOT NULL AND CAST(d.`media_width` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'media_height' AS document_column, SUM(d.`media_height` IS NOT NULL AND CAST(d.`media_height` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'media_aspect_ratio' AS document_column, SUM(d.`media_aspect_ratio` IS NOT NULL AND CAST(d.`media_aspect_ratio` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'media_sample_start' AS document_column, SUM(d.`media_sample_start` IS NOT NULL AND CAST(d.`media_sample_start` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'media_sample_stop' AS document_column, SUM(d.`media_sample_stop` IS NOT NULL AND CAST(d.`media_sample_stop` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_documents FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis';

-- SECTION: Populated columns of the file table, files of documents of thesis eprints
SELECT 'fileid' AS file_column, SUM(f.`fileid` IS NOT NULL AND CAST(f.`fileid` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_files FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'datasetid' AS file_column, SUM(f.`datasetid` IS NOT NULL AND CAST(f.`datasetid` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_files FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'objectid' AS file_column, SUM(f.`objectid` IS NOT NULL AND CAST(f.`objectid` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_files FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'mime_type' AS file_column, SUM(f.`mime_type` IS NOT NULL AND CAST(f.`mime_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_files FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'hash' AS file_column, SUM(f.`hash` IS NOT NULL AND CAST(f.`hash` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_files FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'hash_type' AS file_column, SUM(f.`hash_type` IS NOT NULL AND CAST(f.`hash_type` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_files FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'filesize' AS file_column, SUM(f.`filesize` IS NOT NULL AND CAST(f.`filesize` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_files FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL
SELECT 'mtime (date, by its year column)' AS file_column, SUM(f.`mtime_year` IS NOT NULL AND CAST(f.`mtime_year` AS CHAR) <> '') AS n_populated, COUNT(*) AS n_files FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis';
