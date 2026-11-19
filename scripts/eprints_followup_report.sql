-- eprints_followup_report.sql
--
-- A third counts-only report for the CaltechTHESIS EPrints database, answering
-- two questions left by the first two: (1) what the ep3StaticSite file rule
-- (src/files/visibility.ts) selects, and (2) how the advisor and committee rows
-- with no role are arranged. Same rules and runner as the others:
--
--   SHAPE_SQL=eprints_followup_report.sql ./eprints_shape_report.bash
--
-- The ep3StaticSite rule serves a document only if the eprint is in the archive,
-- the document's security is public, it has no active embargo, and it is not an
-- EPrints thumbnail or index document. A thumbnail or index document is one whose
-- main file name, without its last extension and in lower case, is lightbox,
-- preview, medium, small or indexcodes. That name test is applied here to the
-- main column inside an expression only: the name itself is never shown.

-- SECTION: The ep3StaticSite file rule applied to the documents of thesis eprints
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         d.date_embargo_year AS y, d.date_embargo_month AS m, d.date_embargo_day AS dd,
         LOWER(IF(LOCATE('.', d.main) > 0,
                  SUBSTRING(d.main, 1, CHAR_LENGTH(d.main) - CHAR_LENGTH(SUBSTRING_INDEX(d.main, '.', -1)) - 1),
                  d.main)) AS base
  FROM document d
),
cls AS (
  SELECT doc.docid, doc.content,
         CASE WHEN doc.base IN ('lightbox', 'preview', 'medium', 'small', 'indexcodes') THEN '1 excluded: EPrints thumbnail or index document'
              WHEN e.eprint_status <> 'archive' THEN '2 not served: eprint not in the archive'
              WHEN doc.security IS NULL THEN '3 restricted: no security value'
              WHEN doc.security = 'internal' THEN '3 restricted: internal'
              WHEN doc.security = 'validuser' THEN '3 restricted: campus login'
              WHEN doc.security <> 'public' THEN '3 restricted: other security'
              WHEN doc.y IS NOT NULL AND (doc.y >= 2100
                   OR STR_TO_DATE(CONCAT(doc.y, '-', COALESCE(doc.m, 1), '-', COALESCE(doc.dd, 1)), '%Y-%m-%d') > CURDATE())
                   THEN '4 restricted: active embargo, security public'
              ELSE '5 public: served' END AS rule_class
  FROM doc JOIN eprint e ON e.eprintid = doc.eprintid
  WHERE e.type = 'thesis'
)
SELECT cls.rule_class,
       IF(COUNT(DISTINCT cls.docid) < 5, '<5', COUNT(DISTINCT cls.docid)) AS n_documents,
       COUNT(f.fileid) AS n_files,
       ROUND(SUM(f.filesize) / 1073741824, 2) AS total_gb
FROM cls LEFT JOIN file f ON f.datasetid = 'document' AND f.objectid = cls.docid
GROUP BY cls.rule_class
ORDER BY cls.rule_class;

-- What each class holds, by content.
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         d.date_embargo_year AS y, d.date_embargo_month AS m, d.date_embargo_day AS dd,
         LOWER(IF(LOCATE('.', d.main) > 0,
                  SUBSTRING(d.main, 1, CHAR_LENGTH(d.main) - CHAR_LENGTH(SUBSTRING_INDEX(d.main, '.', -1)) - 1),
                  d.main)) AS base
  FROM document d
),
cls AS (
  SELECT doc.docid, doc.content,
         CASE WHEN doc.base IN ('lightbox', 'preview', 'medium', 'small', 'indexcodes') THEN '1 excluded: EPrints thumbnail or index document'
              WHEN e.eprint_status <> 'archive' THEN '2 not served: eprint not in the archive'
              WHEN doc.security IS NULL THEN '3 restricted: no security value'
              WHEN doc.security = 'internal' THEN '3 restricted: internal'
              WHEN doc.security = 'validuser' THEN '3 restricted: campus login'
              WHEN doc.security <> 'public' THEN '3 restricted: other security'
              WHEN doc.y IS NOT NULL AND (doc.y >= 2100
                   OR STR_TO_DATE(CONCAT(doc.y, '-', COALESCE(doc.m, 1), '-', COALESCE(doc.dd, 1)), '%Y-%m-%d') > CURDATE())
                   THEN '4 restricted: active embargo, security public'
              ELSE '5 public: served' END AS rule_class
  FROM doc JOIN eprint e ON e.eprintid = doc.eprintid
  WHERE e.type = 'thesis'
)
SELECT cls.rule_class, COALESCE(cls.content, '(null)') AS content,
       IF(COUNT(DISTINCT cls.docid) < 5, '<5', COUNT(DISTINCT cls.docid)) AS n_documents,
       ROUND(SUM(f.filesize) / 1073741824, 2) AS total_gb
FROM cls LEFT JOIN file f ON f.datasetid = 'document' AND f.objectid = cls.docid
GROUP BY cls.rule_class, COALESCE(cls.content, '(null)')
ORDER BY cls.rule_class, COUNT(DISTINCT cls.docid) DESC;

-- SECTION: Archive thesis eprints by the kinds of document they hold (public, restricted, administrative form)
-- Public and restricted follow the ep3StaticSite classes above. A thesis that holds both would need
-- a public version and a non-public version, because RDM keeps file visibility per record version.
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         d.date_embargo_year AS y, d.date_embargo_month AS m, d.date_embargo_day AS dd,
         LOWER(IF(LOCATE('.', d.main) > 0,
                  SUBSTRING(d.main, 1, CHAR_LENGTH(d.main) - CHAR_LENGTH(SUBSTRING_INDEX(d.main, '.', -1)) - 1),
                  d.main)) AS base
  FROM document d
),
cls AS (
  SELECT doc.docid, doc.eprintid, doc.content,
         CASE WHEN doc.base IN ('lightbox', 'preview', 'medium', 'small', 'indexcodes') THEN 'excluded'
              WHEN doc.security IS NULL OR doc.security <> 'public' THEN 'restricted'
              WHEN doc.y IS NOT NULL AND (doc.y >= 2100
                   OR STR_TO_DATE(CONCAT(doc.y, '-', COALESCE(doc.m, 1), '-', COALESCE(doc.dd, 1)), '%Y-%m-%d') > CURDATE())
                   THEN 'restricted'
              ELSE 'public' END AS kind
  FROM doc
),
per AS (
  SELECT e.eprintid,
         COALESCE(SUM(cls.kind = 'public'), 0) AS public_docs,
         COALESCE(SUM(cls.kind = 'restricted'), 0) AS restricted_docs,
         COALESCE(SUM(cls.kind = 'restricted' AND cls.content IN ('approvalform', 'embargoform', 'copyright', 'umiform')), 0) AS form_docs,
         COALESCE(SUM(cls.kind = 'restricted' AND (cls.content IS NULL OR cls.content NOT IN ('approvalform', 'embargoform', 'copyright', 'umiform'))), 0) AS other_restricted_docs
  FROM eprint e LEFT JOIN cls ON cls.eprintid = e.eprintid
  WHERE e.type = 'thesis' AND e.eprint_status = 'archive'
  GROUP BY e.eprintid
)
SELECT CASE WHEN public_docs > 0 AND restricted_docs > 0 THEN 'public and restricted documents (two versions)'
            WHEN public_docs > 0 THEN 'public documents only (one public version)'
            WHEN restricted_docs > 0 THEN 'restricted documents only (no public files)'
            ELSE 'no document to migrate' END AS document_mix,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM per
GROUP BY 1
ORDER BY 1;

-- Of the eprints that hold a restricted document: what kind, and whether they also hold a public one.
WITH doc AS (
  SELECT d.docid, d.eprintid, d.content, d.security,
         d.date_embargo_year AS y, d.date_embargo_month AS m, d.date_embargo_day AS dd,
         LOWER(IF(LOCATE('.', d.main) > 0,
                  SUBSTRING(d.main, 1, CHAR_LENGTH(d.main) - CHAR_LENGTH(SUBSTRING_INDEX(d.main, '.', -1)) - 1),
                  d.main)) AS base
  FROM document d
),
cls AS (
  SELECT doc.docid, doc.eprintid, doc.content,
         CASE WHEN doc.base IN ('lightbox', 'preview', 'medium', 'small', 'indexcodes') THEN 'excluded'
              WHEN doc.security IS NULL OR doc.security <> 'public' THEN 'restricted'
              WHEN doc.y IS NOT NULL AND (doc.y >= 2100
                   OR STR_TO_DATE(CONCAT(doc.y, '-', COALESCE(doc.m, 1), '-', COALESCE(doc.dd, 1)), '%Y-%m-%d') > CURDATE())
                   THEN 'restricted'
              ELSE 'public' END AS kind
  FROM doc
),
per AS (
  SELECT e.eprintid,
         COALESCE(SUM(cls.kind = 'public'), 0) AS public_docs,
         COALESCE(SUM(cls.kind = 'restricted'), 0) AS restricted_docs,
         COALESCE(SUM(cls.kind = 'restricted' AND cls.content IN ('approvalform', 'embargoform', 'copyright', 'umiform')), 0) AS form_docs,
         COALESCE(SUM(cls.kind = 'restricted' AND (cls.content IS NULL OR cls.content NOT IN ('approvalform', 'embargoform', 'copyright', 'umiform'))), 0) AS other_restricted_docs
  FROM eprint e LEFT JOIN cls ON cls.eprintid = e.eprintid
  WHERE e.type = 'thesis' AND e.eprint_status = 'archive'
  GROUP BY e.eprintid
)
SELECT IF(form_docs > 0, 'has an administrative form', 'no administrative form') AS forms,
       IF(other_restricted_docs > 0, 'has another restricted document (archival, release, internal final, ...)', 'no other restricted document') AS others,
       IF(public_docs > 0, 'has public documents', 'no public documents') AS public_mix,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM per
WHERE restricted_docs > 0
GROUP BY 1, 2, 3
ORDER BY COUNT(*) DESC;

-- SECTION: The ep3StaticSite thumbnail test against the relation-based test
-- by_name: the main file name is one of the five. by_relation: the document carries an
-- is...VersionOf relation of the kinds EPrints generates.
WITH doc AS (
  SELECT d.docid,
         LOWER(IF(LOCATE('.', d.main) > 0,
                  SUBSTRING(d.main, 1, CHAR_LENGTH(d.main) - CHAR_LENGTH(SUBSTRING_INDEX(d.main, '.', -1)) - 1),
                  d.main)) AS base,
         EXISTS (SELECT 1 FROM document_relation_type r WHERE r.docid = d.docid AND r.relation_type IN (
           'http://eprints.org/relation/isVolatileVersionOf',
           'http://eprints.org/relation/ispreviewThumbnailVersionOf',
           'http://eprints.org/relation/issmallThumbnailVersionOf',
           'http://eprints.org/relation/ismediumThumbnailVersionOf',
           'http://eprints.org/relation/islightboxThumbnailVersionOf',
           'http://eprints.org/relation/isIndexCodesVersionOf')) AS by_relation
  FROM document d
)
SELECT IF(base IN ('lightbox', 'preview', 'medium', 'small', 'indexcodes'), 'by name: yes', 'by name: no') AS thumbnail_by_name,
       IF(by_relation = 1, 'by relation: yes', 'by relation: no') AS thumbnail_by_relation,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_documents
FROM doc
GROUP BY 1, 2
ORDER BY 1, 2;

-- SECTION: Advisor roles by position, for eprints with two or more advisor rows
WITH adv AS (
  SELECT n.eprintid,
         ROW_NUMBER() OVER (PARTITION BY n.eprintid ORDER BY n.pos) AS rn,
         COALESCE(NULLIF(r.thesis_advisor_role, ''), '(missing)') AS role_value
  FROM eprint_thesis_advisor_name n
  LEFT JOIN eprint_thesis_advisor_role r ON r.eprintid = n.eprintid AND r.pos = n.pos
),
multi AS (SELECT eprintid FROM adv GROUP BY eprintid HAVING COUNT(*) >= 2)
SELECT LEAST(a.rn, 4) AS position_in_the_list_4_means_4_or_later, a.role_value,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_rows
FROM adv a JOIN multi m ON m.eprintid = a.eprintid
GROUP BY LEAST(a.rn, 4), a.role_value
ORDER BY 1, COUNT(*) DESC;

-- How the advisor rows with no role are arranged.
WITH adv AS (
  SELECT n.eprintid,
         COALESCE(NULLIF(r.thesis_advisor_role, ''), '(missing)') AS role_value
  FROM eprint_thesis_advisor_name n
  LEFT JOIN eprint_thesis_advisor_role r ON r.eprintid = n.eprintid AND r.pos = n.pos
),
per AS (
  SELECT eprintid, COUNT(*) AS n_rows,
         SUM(role_value = '(missing)') AS n_missing,
         SUM(role_value = 'advisor') AS n_advisor
  FROM adv GROUP BY eprintid
)
SELECT CASE WHEN n_missing = 0 THEN 'no row is missing a role'
            WHEN n_rows = 1 THEN 'one advisor row, and its role is missing'
            WHEN n_advisor > 0 THEN 'several rows, an advisor role exists, another row is missing one'
            ELSE 'several rows, no advisor role among them' END AS advisor_pattern,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM per
GROUP BY 1
ORDER BY 1;

-- SECTION: Committee rows with no role: where they sit in the list
WITH com AS (
  SELECT n.eprintid,
         ROW_NUMBER() OVER (PARTITION BY n.eprintid ORDER BY n.pos) AS rn,
         COUNT(*) OVER (PARTITION BY n.eprintid) AS n_rows,
         COALESCE(NULLIF(r.thesis_committee_role, ''), '(missing)') AS role_value
  FROM eprint_thesis_committee_name n
  LEFT JOIN eprint_thesis_committee_role r ON r.eprintid = n.eprintid AND r.pos = n.pos
),
chairs AS (SELECT eprintid, SUM(role_value = 'chair') AS n_chair FROM com GROUP BY eprintid)
SELECT LEAST(m.rn, 5) AS position_of_the_row_5_means_5_or_later,
       IF(c.n_chair > 0, 'the eprint has a chair row', 'the eprint has no chair row') AS chair_row,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_rows_with_no_role
FROM com m JOIN chairs c ON c.eprintid = m.eprintid
WHERE m.role_value = '(missing)'
GROUP BY LEAST(m.rn, 5), 2
ORDER BY 1, COUNT(*) DESC;

-- How many committee rows each affected eprint has, and how many are missing a role.
WITH com AS (
  SELECT n.eprintid,
         COALESCE(NULLIF(r.thesis_committee_role, ''), '(missing)') AS role_value
  FROM eprint_thesis_committee_name n
  LEFT JOIN eprint_thesis_committee_role r ON r.eprintid = n.eprintid AND r.pos = n.pos
),
per AS (
  SELECT eprintid, COUNT(*) AS n_rows, SUM(role_value = '(missing)') AS n_missing
  FROM com GROUP BY eprintid
)
SELECT LEAST(n_rows, 6) AS committee_rows_6_means_6_or_more,
       LEAST(n_missing, 3) AS rows_missing_a_role_3_means_3_or_more,
       IF(COUNT(*) < 5, '<5', COUNT(*)) AS n_eprints
FROM per
WHERE n_missing > 0
GROUP BY LEAST(n_rows, 6), LEAST(n_missing, 3)
ORDER BY 1, 2;
