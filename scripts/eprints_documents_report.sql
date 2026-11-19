-- eprints_documents_report.sql
--
-- The second shape report for the CaltechTHESIS EPrints database: documents,
-- embargoes, roles and divisions. Same rules as eprints_shape_report.sql
-- (read-only, counts only, no free text, small cells shown as '<5', only SELECT
-- and WITH), and run the same way:
--
--   SHAPE_SQL=eprints_documents_report.sql ./eprints_shape_report.bash
--
-- A "derived" document is one that carries an is...VersionOf relation of the
-- kinds EPrints generates: isVolatileVersionOf, the preview, small, medium and
-- lightbox thumbnail versions, and isIndexCodesVersionOf. The PARENT documents
-- carry the matching has...Version relations, so "has any relation" is NOT the
-- test (the first run of this report used it and was wrong). Section 2 lists the
-- relation types. A "primary" document is any document that is not derived. The single exception to small-cell suppression is section 1's
-- status totals for a thesis-only recount: a count of records per status
-- identifies nobody, and the sample allocation needs it exact.

-- SECTION: Record states for thesis eprints only, and the other types (exact status totals)
SELECT eprint_status, COUNT(*) AS thesis_eprints
FROM eprint WHERE type = 'thesis'
GROUP BY eprint_status
ORDER BY COUNT(*) DESC;

SELECT type, eprint_status, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint WHERE type <> 'thesis'
GROUP BY type, eprint_status
ORDER BY COUNT(*) DESC;

-- SECTION: Document relations and permission groups
SELECT 'relation_type' AS field, COALESCE(relation_type, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document_relation_type
GROUP BY COALESCE(relation_type, '(null)')
ORDER BY COUNT(*) DESC;

SELECT 'permission_group' AS field, COALESCE(permission_group, '(null)') AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM document_permission_group
GROUP BY COALESCE(permission_group, '(null)')
ORDER BY COUNT(*) DESC
LIMIT 40;

SELECT COALESCE(d.security, '(null)') AS security, IF(COUNT(*) < 5, '<5', COUNT(*)) AS documents_with_a_permission_group
FROM document d
WHERE EXISTS (SELECT 1 FROM document_permission_group g WHERE g.docid = d.docid)
GROUP BY COALESCE(d.security, '(null)')
ORDER BY COUNT(*) DESC;

SELECT COUNT(*) AS documents,
       SUM(EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid)) AS with_any_relation,
       SUM(EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf'))) AS derived_documents
FROM document d;

-- SECTION: Primary and derived documents (derived means it is a generated version of another)
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security, d.format,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
)
SELECT IF(derived = 1, 'derived (is a generated version)', 'primary (not a generated version)') AS kind,
       COALESCE(content, '(null)') AS content, COALESCE(security, '(null)') AS security,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents
FROM doc
GROUP BY derived, COALESCE(content, '(null)'), COALESCE(security, '(null)')
ORDER BY derived, COUNT(*) DESC;

WITH doc AS (
  SELECT d.docid, d.eprintid,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
)
SELECT IF(doc.derived = 1, 'derived (is a generated version)', 'primary (not a generated version)') AS kind,
       COUNT(DISTINCT doc.docid) AS n_documents,
       COUNT(f.fileid) AS n_files,
       ROUND(SUM(f.filesize) / 1073741824, 2) AS total_gb
FROM doc LEFT JOIN file f ON f.datasetid = 'document' AND f.objectid = doc.docid
GROUP BY doc.derived
ORDER BY doc.derived;

-- SECTION: Primary documents of archive thesis eprints, by content, security and format
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security, d.format,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
)
SELECT COALESCE(doc.content, '(null)') AS content,
       COALESCE(doc.security, '(null)') AS security,
       COALESCE(doc.format, '(null)') AS format,
       IF(COUNT(DISTINCT doc.docid) < 5, '<5', COUNT(DISTINCT doc.docid)) AS n_documents,
       ROUND(SUM(f.filesize) / 1073741824, 2) AS total_gb
FROM doc
JOIN eprint e ON e.eprintid = doc.eprintid
LEFT JOIN file f ON f.datasetid = 'document' AND f.objectid = doc.docid
WHERE doc.derived = 0 AND e.type = 'thesis' AND e.eprint_status = 'archive'
GROUP BY COALESCE(doc.content, '(null)'), COALESCE(doc.security, '(null)'), COALESCE(doc.format, '(null)')
ORDER BY COUNT(DISTINCT doc.docid) DESC
LIMIT 70;

-- SECTION: Which document is the thesis, per archive thesis eprint
-- Each eprint is placed in the first pattern that fits, and crossed with EPrints' own
-- full_text_status so the two can be compared.
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
),
per AS (
  SELECT e.eprintid, e.full_text_status AS fts,
         COALESCE(SUM(doc.derived = 0), 0) AS primary_docs,
         COALESCE(SUM(doc.derived = 0 AND doc.content = 'final' AND doc.security = 'public'), 0) AS final_public,
         COALESCE(SUM(doc.derived = 0 AND doc.content = 'final' AND doc.security = 'validuser'), 0) AS final_campus,
         COALESCE(SUM(doc.derived = 0 AND doc.content = 'final' AND doc.security = 'internal'), 0) AS final_internal,
         COALESCE(SUM(doc.derived = 0 AND doc.content = 'final'), 0) AS final_any
  FROM eprint e LEFT JOIN doc ON doc.eprintid = e.eprintid
  WHERE e.type = 'thesis' AND e.eprint_status = 'archive'
  GROUP BY e.eprintid, e.full_text_status
)
SELECT CASE WHEN final_public > 0 THEN 'has a public final'
            WHEN final_campus > 0 THEN 'final for campus login only'
            WHEN final_internal > 0 THEN 'final internal only'
            WHEN final_any > 0 THEN 'final with another security value'
            WHEN primary_docs > 0 THEN 'no final, other primary documents'
            ELSE 'no primary documents' END AS thesis_file_pattern,
       COALESCE(fts, '(null)') AS full_text_status,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM per
GROUP BY 1, 2
ORDER BY 1, COUNT(*) DESC;

WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
),
per AS (
  SELECT e.eprintid,
         COALESCE(SUM(doc.derived = 0 AND doc.content = 'final' AND doc.security = 'public'), 0) AS final_public,
         COALESCE(SUM(doc.derived = 0), 0) AS primary_docs
  FROM eprint e LEFT JOIN doc ON doc.eprintid = e.eprintid
  WHERE e.type = 'thesis' AND e.eprint_status = 'archive'
  GROUP BY e.eprintid
)
SELECT 'public finals per eprint (4 means 4 or more)' AS what, LEAST(final_public, 4) AS value, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM per GROUP BY LEAST(final_public, 4)
UNION ALL
SELECT 'primary documents per eprint (10 means 10 or more)', LEAST(primary_docs, 10), IF(COUNT(*) < 5, '<5', COUNT(*))
FROM per GROUP BY LEAST(primary_docs, 10);

-- SECTION: Administrative forms (primary documents of archive thesis eprints)
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
)
SELECT doc.content, COALESCE(doc.security, '(null)') AS security,
       IF(COUNT(DISTINCT doc.eprintid) < 5, '<5', COUNT(DISTINCT doc.eprintid)) AS n_eprints,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents
FROM doc JOIN eprint e ON e.eprintid = doc.eprintid
WHERE doc.derived = 0
  AND doc.content IN ('approvalform', 'embargoform', 'copyright', 'umiform')
  AND e.type = 'thesis' AND e.eprint_status = 'archive'
GROUP BY doc.content, COALESCE(doc.security, '(null)')
ORDER BY COUNT(*) DESC;

-- SECTION: Embargo detail (a year of 2100 or later is a sentinel for a permanent embargo)
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         d.date_embargo_year AS y, d.date_embargo_month AS m, d.date_embargo_day AS dd,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
)
SELECT CASE WHEN y >= 2100 THEN 'permanent (sentinel year)'
            WHEN STR_TO_DATE(CONCAT(y, '-', COALESCE(m, 1), '-', COALESCE(dd, 1)), '%Y-%m-%d') IS NULL THEN 'invalid embargo date'
            WHEN STR_TO_DATE(CONCAT(y, '-', COALESCE(m, 1), '-', COALESCE(dd, 1)), '%Y-%m-%d') < CURDATE() THEN 'past'
            WHEN STR_TO_DATE(CONCAT(y, '-', COALESCE(m, 1), '-', COALESCE(dd, 1)), '%Y-%m-%d') = CURDATE() THEN 'today'
            ELSE 'future' END AS embargo_state,
       IF(derived = 1, 'derived', 'primary') AS kind,
       COALESCE(content, '(null)') AS content,
       COALESCE(security, '(null)') AS security,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents
FROM doc
WHERE y IS NOT NULL
GROUP BY 1, 2, 3, 4
ORDER BY 1, 2, COUNT(*) DESC;

-- Per archive thesis eprint: which kinds of embargo its primary documents carry, against EPrints' full_text_status.
WITH doc AS (
  SELECT d.docid, d.eprintid,
         d.date_embargo_year AS y, d.date_embargo_month AS m, d.date_embargo_day AS dd,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
),
per AS (
  SELECT e.eprintid, e.full_text_status AS fts,
         COALESCE(SUM(doc.derived = 0 AND doc.y IS NOT NULL AND doc.y < 2100
                      AND STR_TO_DATE(CONCAT(doc.y, '-', COALESCE(doc.m, 1), '-', COALESCE(doc.dd, 1)), '%Y-%m-%d') >= CURDATE()), 0) AS future_embargo,
         COALESCE(SUM(doc.derived = 0 AND doc.y IS NOT NULL AND doc.y < 2100
                      AND STR_TO_DATE(CONCAT(doc.y, '-', COALESCE(doc.m, 1), '-', COALESCE(doc.dd, 1)), '%Y-%m-%d') < CURDATE()), 0) AS past_embargo,
         COALESCE(SUM(doc.derived = 0 AND doc.y >= 2100), 0) AS permanent_embargo
  FROM eprint e LEFT JOIN doc ON doc.eprintid = e.eprintid
  WHERE e.type = 'thesis' AND e.eprint_status = 'archive'
  GROUP BY e.eprintid, e.full_text_status
)
SELECT IF(future_embargo > 0, 'has a future embargo', 'no future embargo') AS future_kind,
       IF(past_embargo > 0, 'has a past embargo', 'no past embargo') AS past_kind,
       IF(permanent_embargo > 0, 'has a permanent embargo', 'no permanent embargo') AS permanent_kind,
       COALESCE(fts, '(null)') AS full_text_status,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM per
WHERE future_embargo > 0 OR past_embargo > 0 OR permanent_embargo > 0
GROUP BY 1, 2, 3, 4
ORDER BY COUNT(*) DESC;

-- Primary documents whose embargo date has passed but that EPrints never released.
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         d.date_embargo_year AS y, d.date_embargo_month AS m, d.date_embargo_day AS dd,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
)
SELECT COALESCE(doc.content, '(null)') AS content,
       COALESCE(doc.security, '(null)') AS security,
       COALESCE(e.full_text_status, '(null)') AS full_text_status,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents,
       IF(COUNT(DISTINCT e.eprintid) < 5, '<5', COUNT(DISTINCT e.eprintid)) AS n_eprints
FROM doc JOIN eprint e ON e.eprintid = doc.eprintid
WHERE doc.derived = 0 AND doc.y IS NOT NULL AND doc.y < 2100
  AND STR_TO_DATE(CONCAT(doc.y, '-', COALESCE(doc.m, 1), '-', COALESCE(doc.dd, 1)), '%Y-%m-%d') < CURDATE()
  AND doc.security <> 'public'
GROUP BY COALESCE(doc.content, '(null)'), COALESCE(doc.security, '(null)'), COALESCE(e.full_text_status, '(null)')
ORDER BY COUNT(*) DESC;

-- SECTION: Size of each candidate set of files
WITH doc AS (
  SELECT d.docid, d.content, d.security,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
),
sized AS (
  SELECT doc.docid, doc.derived, doc.content, doc.security, f.filesize
  FROM doc LEFT JOIN file f ON f.datasetid = 'document' AND f.objectid = doc.docid
)
SELECT 'all document files' AS scope, COUNT(filesize) AS n_files, ROUND(SUM(filesize) / 1073741824, 2) AS total_gb FROM sized
UNION ALL SELECT 'derived (is a generated version)', COUNT(filesize), ROUND(SUM(filesize) / 1073741824, 2) FROM sized WHERE derived = 1
UNION ALL SELECT 'primary (not a generated version)', COUNT(filesize), ROUND(SUM(filesize) / 1073741824, 2) FROM sized WHERE derived = 0
UNION ALL SELECT 'primary, content final, any security', COUNT(filesize), ROUND(SUM(filesize) / 1073741824, 2) FROM sized WHERE derived = 0 AND content = 'final'
UNION ALL SELECT 'primary, content final, public', COUNT(filesize), ROUND(SUM(filesize) / 1073741824, 2) FROM sized WHERE derived = 0 AND content = 'final' AND security = 'public'
UNION ALL SELECT 'primary, security public, any content', COUNT(filesize), ROUND(SUM(filesize) / 1073741824, 2) FROM sized WHERE derived = 0 AND security = 'public'
UNION ALL SELECT 'primary, supplemental, any security', COUNT(filesize), ROUND(SUM(filesize) / 1073741824, 2) FROM sized WHERE derived = 0 AND content = 'supplemental'
UNION ALL SELECT 'primary, administrative forms', COUNT(filesize), ROUND(SUM(filesize) / 1073741824, 2) FROM sized WHERE derived = 0 AND content IN ('approvalform', 'embargoform', 'copyright', 'umiform')
UNION ALL SELECT 'primary, archival or release', COUNT(filesize), ROUND(SUM(filesize) / 1073741824, 2) FROM sized WHERE derived = 0 AND content IN ('archival', 'release');

-- SECTION: Advisor and committee rows with no role
SELECT 'advisor name rows' AS what, COUNT(*) AS n FROM eprint_thesis_advisor_name
UNION ALL
SELECT 'advisor name rows with no role row', COUNT(*)
FROM eprint_thesis_advisor_name a
LEFT JOIN eprint_thesis_advisor_role r ON r.eprintid = a.eprintid AND r.pos = a.pos
WHERE r.eprintid IS NULL
UNION ALL
SELECT 'advisor name rows with an empty role', COUNT(*)
FROM eprint_thesis_advisor_name a
JOIN eprint_thesis_advisor_role r ON r.eprintid = a.eprintid AND r.pos = a.pos
WHERE r.thesis_advisor_role IS NULL OR r.thesis_advisor_role = ''
UNION ALL
SELECT 'committee name rows', COUNT(*) FROM eprint_thesis_committee_name
UNION ALL
SELECT 'committee name rows with no role row', COUNT(*)
FROM eprint_thesis_committee_name c
LEFT JOIN eprint_thesis_committee_role r ON r.eprintid = c.eprintid AND r.pos = c.pos
WHERE r.eprintid IS NULL
UNION ALL
SELECT 'committee name rows with an empty role', COUNT(*)
FROM eprint_thesis_committee_name c
JOIN eprint_thesis_committee_role r ON r.eprintid = c.eprintid AND r.pos = c.pos
WHERE r.thesis_committee_role IS NULL OR r.thesis_committee_role = '';

-- Eprints with at least one advisor row that has no usable role, by type and status.
SELECT 'advisor' AS person, e.type, e.eprint_status,
       IF(COUNT(DISTINCT e.eprintid) < 5, '<5', COUNT(DISTINCT e.eprintid)) AS n_eprints_missing_a_role
FROM eprint_thesis_advisor_name a
JOIN eprint e ON e.eprintid = a.eprintid
LEFT JOIN eprint_thesis_advisor_role r ON r.eprintid = a.eprintid AND r.pos = a.pos
WHERE r.eprintid IS NULL OR r.thesis_advisor_role IS NULL OR r.thesis_advisor_role = ''
GROUP BY e.type, e.eprint_status
UNION ALL
SELECT 'committee', e.type, e.eprint_status,
       IF(COUNT(DISTINCT e.eprintid) < 5, '<5', COUNT(DISTINCT e.eprintid))
FROM eprint_thesis_committee_name c
JOIN eprint e ON e.eprintid = c.eprintid
LEFT JOIN eprint_thesis_committee_role r ON r.eprintid = c.eprintid AND r.pos = c.pos
WHERE r.eprintid IS NULL OR r.thesis_committee_role IS NULL OR r.thesis_committee_role = ''
GROUP BY e.type, e.eprint_status;

-- The same, by decade of the degree date, for committee rows (is it the older theses).
SELECT FLOOR(e.date_year / 10) * 10 AS decade,
       IF(COUNT(DISTINCT e.eprintid) < 5, '<5', COUNT(DISTINCT e.eprintid)) AS n_eprints_with_a_committee_row_missing_a_role
FROM eprint_thesis_committee_name c
JOIN eprint e ON e.eprintid = c.eprintid
LEFT JOIN eprint_thesis_committee_role r ON r.eprintid = c.eprintid AND r.pos = c.pos
WHERE e.type = 'thesis' AND (r.eprintid IS NULL OR r.thesis_committee_role IS NULL OR r.thesis_committee_role = '')
GROUP BY FLOOR(e.date_year / 10) * 10
ORDER BY 1;

-- Chairs per archive thesis eprint (3 means 3 or more).
WITH ch AS (
  SELECT e.eprintid, COALESCE(SUM(r.thesis_committee_role = 'chair'), 0) AS chairs
  FROM eprint e LEFT JOIN eprint_thesis_committee_role r ON r.eprintid = e.eprintid
  WHERE e.type = 'thesis' AND e.eprint_status = 'archive'
  GROUP BY e.eprintid
)
SELECT LEAST(chairs, 3) AS chairs_per_eprint, IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM ch GROUP BY LEAST(chairs, 3) ORDER BY 1;

-- SECTION: Divisions, versions and odd file types
SELECT e.type, e.eprint_status, IF(COUNT(*) < 5, '<5', COUNT(*)) AS eprints_without_a_division
FROM eprint e LEFT JOIN eprint_divisions dv ON dv.eprintid = e.eprintid
WHERE dv.eprintid IS NULL
GROUP BY e.type, e.eprint_status
ORDER BY COUNT(*) DESC;

-- Eprints that succeed another: the status of each side.
SELECT e.eprint_status AS successor_status, p.eprint_status AS predecessor_status, e.type,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n
FROM eprint e JOIN eprint p ON p.eprintid = e.succeeds
WHERE e.succeeds > 0
GROUP BY e.eprint_status, p.eprint_status, e.type;

-- The odd rfc822 mime type, by kind, content and security.
WITH doc AS (
  SELECT d.docid, d.content, d.security,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN ('http://eprints.org/relation/isVolatileVersionOf', 'http://eprints.org/relation/ispreviewThumbnailVersionOf', 'http://eprints.org/relation/issmallThumbnailVersionOf', 'http://eprints.org/relation/ismediumThumbnailVersionOf', 'http://eprints.org/relation/islightboxThumbnailVersionOf', 'http://eprints.org/relation/isIndexCodesVersionOf')) AS derived
  FROM document d
)
SELECT IF(doc.derived = 1, 'derived', 'primary') AS kind,
       COALESCE(doc.content, '(null)') AS content,
       COALESCE(doc.security, '(null)') AS security,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_files
FROM doc JOIN file f ON f.datasetid = 'document' AND f.objectid = doc.docid
WHERE f.mime_type LIKE 'message/rfc822%'
GROUP BY doc.derived, COALESCE(doc.content, '(null)'), COALESCE(doc.security, '(null)')
ORDER BY COUNT(*) DESC;
