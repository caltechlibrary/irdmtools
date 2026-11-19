-- eprints_shape_report.sql
--
-- Read-only, counts-only shape report for the CaltechTHESIS EPrints database.
-- It is run by eprints_shape_report.bash, one "-- SECTION:" at a time, in a
-- read-only MySQL session.
--
-- Rules this file keeps (eprints_shape_report.bash refuses a file that breaks
-- them, and eprints_shape_report_test.bash tests that it does):
--   * only SELECT and WITH statements;
--   * no column that names or contacts a person, or names a file;
--   * a count under 5 in a grouped result is shown as '<5' (small cells);
--   * no title, abstract, note or other free text is ever selected, only
--     counted, measured or tested for a character.
-- Values that ARE shown are controlled vocabulary (status, security, type,
-- degree, option and division codes, roles, mime types).
-- Dates are stored by EPrints as _year, _month and _day columns.

-- SECTION: Server, date and totals
SELECT VERSION() AS mysql_version,
       CURDATE() AS report_date,
       (SELECT COUNT(*) FROM eprint) AS eprints,
       (SELECT COUNT(*) FROM document) AS documents,
       (SELECT COUNT(*) FROM file WHERE datasetid = 'document') AS document_files,
       (SELECT COUNT(*) FROM file WHERE datasetid = 'history') AS history_files;

-- SECTION: Record states
SELECT eprint_status, metadata_visibility, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint
GROUP BY eprint_status, metadata_visibility
ORDER BY COUNT(*) DESC;

SELECT eprint_status, COALESCE(full_text_status, '(null)') AS full_text_status,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint
GROUP BY eprint_status, COALESCE(full_text_status, '(null)')
ORDER BY COUNT(*) DESC;

SELECT COALESCE(review_status, '(null)') AS review_status,
       COALESCE(ispublished, '(null)') AS ispublished,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint
GROUP BY COALESCE(review_status, '(null)'), COALESCE(ispublished, '(null)')
ORDER BY COUNT(*) DESC;

-- Versioning links between eprints (counts of eprints, not ids).
SELECT SUM(succeeds IS NOT NULL AND succeeds > 0) AS eprints_that_succeed_another,
       SUM(replacedby IS NOT NULL AND replacedby > 0) AS eprints_replaced_by_another
FROM eprint;

-- SECTION: Types, degrees and institutions
SELECT 'type' AS field, COALESCE(type, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint GROUP BY COALESCE(type, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'thesis_type' AS field, COALESCE(thesis_type, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint GROUP BY COALESCE(thesis_type, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'thesis_degree' AS field, COALESCE(thesis_degree, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint GROUP BY COALESCE(thesis_degree, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'thesis_degree_grantor' AS field, COALESCE(thesis_degree_grantor, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint GROUP BY COALESCE(thesis_degree_grantor, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'date_type' AS field, COALESCE(date_type, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint GROUP BY COALESCE(date_type, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'institution' AS field, COALESCE(institution, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint GROUP BY COALESCE(institution, '(null)') ORDER BY COUNT(*) DESC LIMIT 40;

SELECT 'department' AS field, COALESCE(department, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint GROUP BY COALESCE(department, '(null)') ORDER BY COUNT(*) DESC LIMIT 60;

-- thesis_type crossed with thesis_degree, both controlled.
SELECT COALESCE(thesis_type, '(null)') AS thesis_type, COALESCE(thesis_degree, '(null)') AS thesis_degree,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint
GROUP BY COALESCE(thesis_type, '(null)'), COALESCE(thesis_degree, '(null)')
ORDER BY COUNT(*) DESC;

-- SECTION: Date precision (counts of eprints; each date is year, month and day columns)
-- with_year: a year is set. year_only: year and nothing else. year_and_month: year
-- and month, no day (year_month is a reserved word in MySQL, so it cannot be a name). full_date: year, month and day. odd_parts: a month or day
-- without the part above it. year_zero: the year is 0.
SELECT 'all' AS scope, 'date' AS field, COUNT(*) AS total,
       SUM(date_year IS NOT NULL) AS with_year,
       SUM(date_year IS NOT NULL AND date_month IS NULL AND date_day IS NULL) AS year_only,
       SUM(date_year IS NOT NULL AND date_month IS NOT NULL AND date_day IS NULL) AS year_and_month,
       SUM(date_year IS NOT NULL AND date_month IS NOT NULL AND date_day IS NOT NULL) AS full_date,
       SUM((date_year IS NULL AND (date_month IS NOT NULL OR date_day IS NOT NULL)) OR (date_month IS NULL AND date_day IS NOT NULL)) AS odd_parts,
       SUM(date_year = 0) AS year_zero
FROM eprint
UNION ALL
SELECT 'all', 'thesis_submitted_date', COUNT(*),
       SUM(thesis_submitted_date_year IS NOT NULL),
       SUM(thesis_submitted_date_year IS NOT NULL AND thesis_submitted_date_month IS NULL AND thesis_submitted_date_day IS NULL),
       SUM(thesis_submitted_date_year IS NOT NULL AND thesis_submitted_date_month IS NOT NULL AND thesis_submitted_date_day IS NULL),
       SUM(thesis_submitted_date_year IS NOT NULL AND thesis_submitted_date_month IS NOT NULL AND thesis_submitted_date_day IS NOT NULL),
       SUM((thesis_submitted_date_year IS NULL AND (thesis_submitted_date_month IS NOT NULL OR thesis_submitted_date_day IS NOT NULL)) OR (thesis_submitted_date_month IS NULL AND thesis_submitted_date_day IS NOT NULL)),
       SUM(thesis_submitted_date_year = 0)
FROM eprint
UNION ALL
SELECT 'all', 'thesis_defense_date', COUNT(*),
       SUM(thesis_defense_date_year IS NOT NULL),
       SUM(thesis_defense_date_year IS NOT NULL AND thesis_defense_date_month IS NULL AND thesis_defense_date_day IS NULL),
       SUM(thesis_defense_date_year IS NOT NULL AND thesis_defense_date_month IS NOT NULL AND thesis_defense_date_day IS NULL),
       SUM(thesis_defense_date_year IS NOT NULL AND thesis_defense_date_month IS NOT NULL AND thesis_defense_date_day IS NOT NULL),
       SUM((thesis_defense_date_year IS NULL AND (thesis_defense_date_month IS NOT NULL OR thesis_defense_date_day IS NOT NULL)) OR (thesis_defense_date_month IS NULL AND thesis_defense_date_day IS NOT NULL)),
       SUM(thesis_defense_date_year = 0)
FROM eprint
UNION ALL
SELECT 'all', 'thesis_degree_date', COUNT(*),
       SUM(thesis_degree_date_year IS NOT NULL),
       SUM(thesis_degree_date_year IS NOT NULL AND thesis_degree_date_month IS NULL AND thesis_degree_date_day IS NULL),
       SUM(thesis_degree_date_year IS NOT NULL AND thesis_degree_date_month IS NOT NULL AND thesis_degree_date_day IS NULL),
       SUM(thesis_degree_date_year IS NOT NULL AND thesis_degree_date_month IS NOT NULL AND thesis_degree_date_day IS NOT NULL),
       SUM((thesis_degree_date_year IS NULL AND (thesis_degree_date_month IS NOT NULL OR thesis_degree_date_day IS NOT NULL)) OR (thesis_degree_date_month IS NULL AND thesis_degree_date_day IS NOT NULL)),
       SUM(thesis_degree_date_year = 0)
FROM eprint
UNION ALL
SELECT 'all', 'thesis_approved_date', COUNT(*),
       SUM(thesis_approved_date_year IS NOT NULL),
       SUM(thesis_approved_date_year IS NOT NULL AND thesis_approved_date_month IS NULL AND thesis_approved_date_day IS NULL),
       SUM(thesis_approved_date_year IS NOT NULL AND thesis_approved_date_month IS NOT NULL AND thesis_approved_date_day IS NULL),
       SUM(thesis_approved_date_year IS NOT NULL AND thesis_approved_date_month IS NOT NULL AND thesis_approved_date_day IS NOT NULL),
       SUM((thesis_approved_date_year IS NULL AND (thesis_approved_date_month IS NOT NULL OR thesis_approved_date_day IS NOT NULL)) OR (thesis_approved_date_month IS NULL AND thesis_approved_date_day IS NOT NULL)),
       SUM(thesis_approved_date_year = 0)
FROM eprint
UNION ALL
SELECT 'all', 'thesis_public_date', COUNT(*),
       SUM(thesis_public_date_year IS NOT NULL),
       SUM(thesis_public_date_year IS NOT NULL AND thesis_public_date_month IS NULL AND thesis_public_date_day IS NULL),
       SUM(thesis_public_date_year IS NOT NULL AND thesis_public_date_month IS NOT NULL AND thesis_public_date_day IS NULL),
       SUM(thesis_public_date_year IS NOT NULL AND thesis_public_date_month IS NOT NULL AND thesis_public_date_day IS NOT NULL),
       SUM((thesis_public_date_year IS NULL AND (thesis_public_date_month IS NOT NULL OR thesis_public_date_day IS NOT NULL)) OR (thesis_public_date_month IS NULL AND thesis_public_date_day IS NOT NULL)),
       SUM(thesis_public_date_year = 0)
FROM eprint
UNION ALL
SELECT 'all', 'gradofc_approval_date', COUNT(*),
       SUM(gradofc_approval_date_year IS NOT NULL),
       SUM(gradofc_approval_date_year IS NOT NULL AND gradofc_approval_date_month IS NULL AND gradofc_approval_date_day IS NULL),
       SUM(gradofc_approval_date_year IS NOT NULL AND gradofc_approval_date_month IS NOT NULL AND gradofc_approval_date_day IS NULL),
       SUM(gradofc_approval_date_year IS NOT NULL AND gradofc_approval_date_month IS NOT NULL AND gradofc_approval_date_day IS NOT NULL),
       SUM((gradofc_approval_date_year IS NULL AND (gradofc_approval_date_month IS NOT NULL OR gradofc_approval_date_day IS NOT NULL)) OR (gradofc_approval_date_month IS NULL AND gradofc_approval_date_day IS NOT NULL)),
       SUM(gradofc_approval_date_year = 0)
FROM eprint
UNION ALL
SELECT 'archive', 'date', COUNT(*),
       SUM(date_year IS NOT NULL),
       SUM(date_year IS NOT NULL AND date_month IS NULL AND date_day IS NULL),
       SUM(date_year IS NOT NULL AND date_month IS NOT NULL AND date_day IS NULL),
       SUM(date_year IS NOT NULL AND date_month IS NOT NULL AND date_day IS NOT NULL),
       SUM((date_year IS NULL AND (date_month IS NOT NULL OR date_day IS NOT NULL)) OR (date_month IS NULL AND date_day IS NOT NULL)),
       SUM(date_year = 0)
FROM eprint WHERE eprint_status = 'archive'
UNION ALL
SELECT 'archive', 'thesis_submitted_date', COUNT(*),
       SUM(thesis_submitted_date_year IS NOT NULL),
       SUM(thesis_submitted_date_year IS NOT NULL AND thesis_submitted_date_month IS NULL AND thesis_submitted_date_day IS NULL),
       SUM(thesis_submitted_date_year IS NOT NULL AND thesis_submitted_date_month IS NOT NULL AND thesis_submitted_date_day IS NULL),
       SUM(thesis_submitted_date_year IS NOT NULL AND thesis_submitted_date_month IS NOT NULL AND thesis_submitted_date_day IS NOT NULL),
       SUM((thesis_submitted_date_year IS NULL AND (thesis_submitted_date_month IS NOT NULL OR thesis_submitted_date_day IS NOT NULL)) OR (thesis_submitted_date_month IS NULL AND thesis_submitted_date_day IS NOT NULL)),
       SUM(thesis_submitted_date_year = 0)
FROM eprint WHERE eprint_status = 'archive'
UNION ALL
SELECT 'archive', 'thesis_defense_date', COUNT(*),
       SUM(thesis_defense_date_year IS NOT NULL),
       SUM(thesis_defense_date_year IS NOT NULL AND thesis_defense_date_month IS NULL AND thesis_defense_date_day IS NULL),
       SUM(thesis_defense_date_year IS NOT NULL AND thesis_defense_date_month IS NOT NULL AND thesis_defense_date_day IS NULL),
       SUM(thesis_defense_date_year IS NOT NULL AND thesis_defense_date_month IS NOT NULL AND thesis_defense_date_day IS NOT NULL),
       SUM((thesis_defense_date_year IS NULL AND (thesis_defense_date_month IS NOT NULL OR thesis_defense_date_day IS NOT NULL)) OR (thesis_defense_date_month IS NULL AND thesis_defense_date_day IS NOT NULL)),
       SUM(thesis_defense_date_year = 0)
FROM eprint WHERE eprint_status = 'archive'
UNION ALL
SELECT 'archive', 'thesis_approved_date', COUNT(*),
       SUM(thesis_approved_date_year IS NOT NULL),
       SUM(thesis_approved_date_year IS NOT NULL AND thesis_approved_date_month IS NULL AND thesis_approved_date_day IS NULL),
       SUM(thesis_approved_date_year IS NOT NULL AND thesis_approved_date_month IS NOT NULL AND thesis_approved_date_day IS NULL),
       SUM(thesis_approved_date_year IS NOT NULL AND thesis_approved_date_month IS NOT NULL AND thesis_approved_date_day IS NOT NULL),
       SUM((thesis_approved_date_year IS NULL AND (thesis_approved_date_month IS NOT NULL OR thesis_approved_date_day IS NOT NULL)) OR (thesis_approved_date_month IS NULL AND thesis_approved_date_day IS NOT NULL)),
       SUM(thesis_approved_date_year = 0)
FROM eprint WHERE eprint_status = 'archive'
UNION ALL
SELECT 'archive', 'gradofc_approval_date', COUNT(*),
       SUM(gradofc_approval_date_year IS NOT NULL),
       SUM(gradofc_approval_date_year IS NOT NULL AND gradofc_approval_date_month IS NULL AND gradofc_approval_date_day IS NULL),
       SUM(gradofc_approval_date_year IS NOT NULL AND gradofc_approval_date_month IS NOT NULL AND gradofc_approval_date_day IS NULL),
       SUM(gradofc_approval_date_year IS NOT NULL AND gradofc_approval_date_month IS NOT NULL AND gradofc_approval_date_day IS NOT NULL),
       SUM((gradofc_approval_date_year IS NULL AND (gradofc_approval_date_month IS NOT NULL OR gradofc_approval_date_day IS NOT NULL)) OR (gradofc_approval_date_month IS NULL AND gradofc_approval_date_day IS NOT NULL)),
       SUM(gradofc_approval_date_year = 0)
FROM eprint WHERE eprint_status = 'archive'
UNION ALL
SELECT 'documents', 'document.date_embargo', COUNT(*),
       SUM(date_embargo_year IS NOT NULL),
       SUM(date_embargo_year IS NOT NULL AND date_embargo_month IS NULL AND date_embargo_day IS NULL),
       SUM(date_embargo_year IS NOT NULL AND date_embargo_month IS NOT NULL AND date_embargo_day IS NULL),
       SUM(date_embargo_year IS NOT NULL AND date_embargo_month IS NOT NULL AND date_embargo_day IS NOT NULL),
       SUM((date_embargo_year IS NULL AND (date_embargo_month IS NOT NULL OR date_embargo_day IS NOT NULL)) OR (date_embargo_month IS NULL AND date_embargo_day IS NOT NULL)),
       SUM(date_embargo_year = 0)
FROM document;

-- SECTION: Items per eprint (0 to 4, and 5 meaning 5 or more)
SELECT * FROM (
  SELECT 'creators' AS item, LEAST(c, 5) AS per_eprint, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_creators_name t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'thesis_advisor', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_thesis_advisor_name t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'thesis_committee', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_thesis_committee_name t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'divisions', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_divisions t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'option_major', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_option_major t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'option_minor', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_option_minor t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'subjects', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_subjects t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'funders', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_funders_agency t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'local_group', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_local_group t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'related_url', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_related_url_url t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'other_numbering', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_other_numbering_system_name t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'alt_title', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_alt_title t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
  UNION ALL
  SELECT 'item_issues', LEAST(c, 5), IF(COUNT(*) < 5, '<5', COUNT(*))
  FROM (SELECT e.eprintid, COUNT(t.eprintid) AS c FROM eprint e LEFT JOIN eprint_item_issues_type t ON t.eprintid = e.eprintid GROUP BY e.eprintid) x
  GROUP BY LEAST(c, 5)
) u
ORDER BY item, per_eprint;

-- SECTION: Roles and identifiers (counts only, never the values)
SELECT 'thesis_advisor_role' AS field, COALESCE(thesis_advisor_role, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint_thesis_advisor_role GROUP BY COALESCE(thesis_advisor_role, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'thesis_committee_role' AS field, COALESCE(thesis_committee_role, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint_thesis_committee_role GROUP BY COALESCE(thesis_committee_role, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'creators_name rows' AS what, COUNT(*) AS n FROM eprint_creators_name
UNION ALL SELECT 'creators_id set', COUNT(*) FROM eprint_creators_id WHERE creators_id IS NOT NULL AND creators_id <> ''
UNION ALL SELECT 'creators_orcid set', COUNT(*) FROM eprint_creators_orcid WHERE creators_orcid IS NOT NULL AND creators_orcid <> ''
UNION ALL SELECT 'thesis_advisor_name rows', COUNT(*) FROM eprint_thesis_advisor_name
UNION ALL SELECT 'thesis_advisor_id set', COUNT(*) FROM eprint_thesis_advisor_id WHERE thesis_advisor_id IS NOT NULL AND thesis_advisor_id <> ''
UNION ALL SELECT 'thesis_advisor_orcid set', COUNT(*) FROM eprint_thesis_advisor_orcid WHERE thesis_advisor_orcid IS NOT NULL AND thesis_advisor_orcid <> ''
UNION ALL SELECT 'thesis_committee_name rows', COUNT(*) FROM eprint_thesis_committee_name
UNION ALL SELECT 'thesis_committee_id set', COUNT(*) FROM eprint_thesis_committee_id WHERE thesis_committee_id IS NOT NULL AND thesis_committee_id <> ''
UNION ALL SELECT 'thesis_committee_orcid set', COUNT(*) FROM eprint_thesis_committee_orcid WHERE thesis_committee_orcid IS NOT NULL AND thesis_committee_orcid <> ''
UNION ALL SELECT 'distinct creators_id', COUNT(DISTINCT creators_id) FROM eprint_creators_id
UNION ALL SELECT 'distinct thesis_advisor_id', COUNT(DISTINCT thesis_advisor_id) FROM eprint_thesis_advisor_id;

-- SECTION: Controlled vocabularies in use (value, then count)
SELECT 'divisions' AS field, divisions AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint_divisions GROUP BY divisions ORDER BY COUNT(*) DESC LIMIT 80;

SELECT 'option_major' AS field, option_major AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint_option_major GROUP BY option_major ORDER BY COUNT(*) DESC LIMIT 120;

SELECT 'option_minor' AS field, option_minor AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint_option_minor GROUP BY option_minor ORDER BY COUNT(*) DESC LIMIT 120;

SELECT 'local_group' AS field, local_group AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint_local_group GROUP BY local_group ORDER BY COUNT(*) DESC LIMIT 80;

SELECT 'divisions' AS field, COUNT(DISTINCT divisions) AS distinct_values FROM eprint_divisions
UNION ALL SELECT 'option_major', COUNT(DISTINCT option_major) FROM eprint_option_major
UNION ALL SELECT 'option_minor', COUNT(DISTINCT option_minor) FROM eprint_option_minor
UNION ALL SELECT 'local_group', COUNT(DISTINCT local_group) FROM eprint_local_group
UNION ALL SELECT 'subjects', COUNT(DISTINCT subjects) FROM eprint_subjects;

-- SECTION: Documents (EPrints keeps derived formats as documents too)
-- Documents per eprint, 0 to 9, and 10 meaning 10 or more.
SELECT LEAST(c, 10) AS documents_per_eprint, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM (SELECT e.eprintid, COUNT(d.docid) AS c FROM eprint e LEFT JOIN document d ON d.eprintid = e.eprintid GROUP BY e.eprintid) x
GROUP BY LEAST(c, 10)
ORDER BY 1;

SELECT 'security' AS field, COALESCE(security, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document GROUP BY COALESCE(security, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'content' AS field, COALESCE(content, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document GROUP BY COALESCE(content, '(null)') ORDER BY COUNT(*) DESC;

SELECT 'format' AS field, COALESCE(format, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document GROUP BY COALESCE(format, '(null)') ORDER BY COUNT(*) DESC LIMIT 40;

SELECT 'mime_type' AS field, COALESCE(mime_type, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document GROUP BY COALESCE(mime_type, '(null)') ORDER BY COUNT(*) DESC LIMIT 40;

SELECT 'language' AS field, COALESCE(language, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document GROUP BY COALESCE(language, '(null)') ORDER BY COUNT(*) DESC LIMIT 20;

SELECT 'license' AS field, COALESCE(license, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document GROUP BY COALESCE(license, '(null)') ORDER BY COUNT(*) DESC LIMIT 20;

-- content crossed with security: which documents are the thesis itself.
SELECT COALESCE(content, '(null)') AS content, COALESCE(security, '(null)') AS security,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document
GROUP BY COALESCE(content, '(null)'), COALESCE(security, '(null)')
ORDER BY COUNT(*) DESC;

-- Is a main file set on the document (the file name itself is never shown).
SELECT SUM(main IS NOT NULL AND main <> '') AS documents_with_a_main_file,
       SUM(main IS NULL OR main = '') AS documents_without
FROM document;

-- SECTION: Embargoes (documents)
-- The embargo date is compared with today; a missing month or day counts as the 1st
-- for the comparison only, and is never shown.
SELECT CASE
         WHEN date_embargo_year IS NULL THEN 'no embargo date'
         WHEN STR_TO_DATE(CONCAT(date_embargo_year, '-', COALESCE(date_embargo_month, 1), '-', COALESCE(date_embargo_day, 1)), '%Y-%m-%d') IS NULL THEN 'invalid embargo date'
         WHEN STR_TO_DATE(CONCAT(date_embargo_year, '-', COALESCE(date_embargo_month, 1), '-', COALESCE(date_embargo_day, 1)), '%Y-%m-%d') < CURDATE() THEN 'embargo date in the past'
         WHEN STR_TO_DATE(CONCAT(date_embargo_year, '-', COALESCE(date_embargo_month, 1), '-', COALESCE(date_embargo_day, 1)), '%Y-%m-%d') = CURDATE() THEN 'embargo date is today'
         ELSE 'embargo date in the future'
       END AS embargo_state,
       COALESCE(security, '(null)') AS security,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents
FROM document
GROUP BY 1, 2
ORDER BY 1, 2;

-- Embargo state by the eprint's own status and by what the document is.
SELECT e.eprint_status,
       CASE WHEN d.date_embargo_year IS NULL THEN 'no embargo date' ELSE 'has embargo date' END AS embargo,
       COALESCE(d.content, '(null)') AS content,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents
FROM document d JOIN eprint e ON e.eprintid = d.eprintid
GROUP BY e.eprint_status, 2, COALESCE(d.content, '(null)')
ORDER BY e.eprint_status, 2, COUNT(*) DESC;

-- Year of the embargo end date for dates today or later.
SELECT date_embargo_year AS embargo_end_year, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents
FROM document
WHERE date_embargo_year IS NOT NULL
  AND STR_TO_DATE(CONCAT(date_embargo_year, '-', COALESCE(date_embargo_month, 1), '-', COALESCE(date_embargo_day, 1)), '%Y-%m-%d') >= CURDATE()
GROUP BY date_embargo_year
ORDER BY date_embargo_year;

-- SECTION: Embargo and security patterns per eprint (the mixed-document question)
-- For each eprint with documents: are none, some or all of its documents embargoed,
-- and does it hold one security value or several.
WITH d AS (
  SELECT eprintid, COUNT(*) AS n, SUM(date_embargo_year IS NOT NULL) AS embargoed, COUNT(DISTINCT security) AS securities
  FROM document GROUP BY eprintid
)
SELECT e.eprint_status,
       CASE WHEN d.embargoed = 0 THEN 'no document embargoed'
            WHEN d.embargoed = d.n THEN 'all documents embargoed'
            ELSE 'some documents embargoed' END AS embargo_pattern,
       IF(d.securities > 1, 'mixed security values', 'one security value') AS security_pattern,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM d JOIN eprint e ON e.eprintid = d.eprintid
GROUP BY e.eprint_status, 2, 3
ORDER BY e.eprint_status, 2, 3;

-- Eprints with and without any document, by status.
SELECT e.eprint_status,
       IF(SUM(d.eprintid IS NULL) < 5, '<5', SUM(d.eprintid IS NULL)) AS eprints_without_documents,
       IF(SUM(d.eprintid IS NOT NULL) < 5, '<5', SUM(d.eprintid IS NOT NULL)) AS eprints_with_documents
FROM eprint e
LEFT JOIN (SELECT DISTINCT eprintid FROM document) d ON d.eprintid = e.eprintid
GROUP BY e.eprint_status
ORDER BY e.eprint_status;

-- SECTION: Files
-- Files per document, 0 to 9, and 10 meaning 10 or more.
SELECT LEAST(c, 10) AS files_per_document, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents
FROM (SELECT d.docid, COUNT(f.fileid) AS c FROM document d LEFT JOIN file f ON f.datasetid = 'document' AND f.objectid = d.docid GROUP BY d.docid) x
GROUP BY LEAST(c, 10)
ORDER BY 1;

-- File sizes: how many files and how many gigabytes in each size band.
SELECT CASE WHEN filesize < 1048576 THEN '1 under 1 MB'
            WHEN filesize < 10485760 THEN '2 1 to 10 MB'
            WHEN filesize < 104857600 THEN '3 10 to 100 MB'
            WHEN filesize < 1073741824 THEN '4 100 MB to 1 GB'
            ELSE '5 1 GB or more' END AS size_band,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_files,
       ROUND(SUM(filesize) / 1073741824, 2) AS total_gb
FROM file WHERE datasetid = 'document'
GROUP BY 1
ORDER BY 1;

SELECT COUNT(*) AS document_files,
       ROUND(SUM(filesize) / 1073741824, 2) AS total_gb,
       ROUND(MAX(filesize) / 1048576, 1) AS largest_mb
FROM file WHERE datasetid = 'document';

SELECT 'mime_type' AS field, COALESCE(mime_type, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM file WHERE datasetid = 'document' GROUP BY COALESCE(mime_type, '(null)') ORDER BY COUNT(*) DESC LIMIT 30;

SELECT 'hash_type' AS field, COALESCE(hash_type, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM file WHERE datasetid = 'document' GROUP BY COALESCE(hash_type, '(null)') ORDER BY COUNT(*) DESC;

SELECT SUM(hash IS NULL OR hash = '') AS files_without_a_hash,
       SUM(filesize IS NULL OR filesize = 0) AS files_with_no_size
FROM file WHERE datasetid = 'document';

-- Document files whose document row is missing.
SELECT COUNT(*) AS files_with_no_document
FROM file f LEFT JOIN document d ON d.docid = f.objectid
WHERE f.datasetid = 'document' AND d.docid IS NULL;

-- SECTION: Text conditions and presence (counts only, never the text)
-- Backslash quote pairs left by PHP magic quotes, dollar signs (LaTeX), non-ASCII
-- text, empty fields and the longest fields. The text itself is never selected.
SELECT COUNT(*) AS eprints,
       SUM(title IS NULL OR title = '') AS empty_title,
       SUM(abstract IS NULL OR abstract = '') AS empty_abstract,
       SUM(LOCATE(CONCAT(CHAR(92), CHAR(34)), title) > 0) AS title_backslash_dquote,
       SUM(LOCATE(CONCAT(CHAR(92), CHAR(39)), title) > 0) AS title_backslash_squote,
       SUM(LOCATE(CONCAT(CHAR(92), CHAR(34)), abstract) > 0) AS abstract_backslash_dquote,
       SUM(LOCATE(CONCAT(CHAR(92), CHAR(39)), abstract) > 0) AS abstract_backslash_squote,
       SUM(LOCATE(CHAR(36), title) > 0) AS title_has_dollar,
       SUM(LOCATE(CHAR(36), abstract) > 0) AS abstract_has_dollar,
       SUM(LENGTH(title) <> CHAR_LENGTH(title)) AS title_non_ascii,
       SUM(LENGTH(abstract) <> CHAR_LENGTH(abstract)) AS abstract_non_ascii,
       MAX(CHAR_LENGTH(title)) AS longest_title_chars,
       MAX(CHAR_LENGTH(abstract)) AS longest_abstract_chars
FROM eprint;

SELECT CASE WHEN abstract IS NULL OR abstract = '' THEN '0 empty'
            WHEN CHAR_LENGTH(abstract) < 1000 THEN '1 under 1,000 chars'
            WHEN CHAR_LENGTH(abstract) < 5000 THEN '2 1,000 to 5,000'
            WHEN CHAR_LENGTH(abstract) < 20000 THEN '3 5,000 to 20,000'
            ELSE '4 20,000 or more' END AS abstract_length,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM eprint GROUP BY 1 ORDER BY 1;

SELECT SUM(doi IS NOT NULL AND doi <> '') AS has_doi,
       SUM(official_url IS NOT NULL AND official_url <> '') AS has_official_url,
       SUM(id_number IS NOT NULL AND id_number <> '') AS has_id_number,
       SUM(institution IS NOT NULL AND institution <> '') AS has_institution,
       SUM(department IS NOT NULL AND department <> '') AS has_department,
       SUM(thesis_awards IS NOT NULL AND thesis_awards <> '') AS has_thesis_awards,
       SUM(rights IS NOT NULL AND rights <> '') AS has_rights,
       SUM(copyright_statement IS NOT NULL AND copyright_statement <> '') AS has_copyright_statement,
       SUM(keywords IS NOT NULL AND keywords <> '') AS has_keywords,
       SUM(note IS NOT NULL AND note <> '') AS has_note,
       SUM(suggestions IS NOT NULL AND suggestions <> '') AS has_suggestions,
       SUM(item_issues_count IS NOT NULL AND item_issues_count > 0) AS has_item_issues
FROM eprint;
