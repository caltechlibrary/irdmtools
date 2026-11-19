-- eprints_access_report.sql
--
-- A sixth counts-only report for the CaltechTHESIS EPrints database: how the
-- record status (eprint_status) lines up with the review status, the full text
-- status and the metadata visibility, for thesis eprints only. It answers one
-- question for the gap table: are the theses whose review is not approved the
-- same ones that are not in the archive? Same runner as the others:
--
--   SHAPE_SQL=eprints_access_report.sql ./eprints_shape_report.bash

-- SECTION: eprint_status by review_status, thesis eprints
SELECT eprint_status, COALESCE(review_status, '(null)') AS review_status, COUNT(*) AS n
FROM eprint WHERE type = 'thesis'
GROUP BY eprint_status, review_status
ORDER BY eprint_status, n DESC;

-- SECTION: eprint_status by full_text_status, thesis eprints
SELECT eprint_status, COALESCE(full_text_status, '(null)') AS full_text_status, COUNT(*) AS n
FROM eprint WHERE type = 'thesis'
GROUP BY eprint_status, full_text_status
ORDER BY eprint_status, n DESC;

-- SECTION: eprint_status by metadata_visibility, thesis eprints
SELECT eprint_status, COALESCE(metadata_visibility, '(null)') AS metadata_visibility, COUNT(*) AS n
FROM eprint WHERE type = 'thesis'
GROUP BY eprint_status, metadata_visibility
ORDER BY eprint_status, n DESC;

-- SECTION: archive thesis eprints that are not approved, by the kinds of document they hold
SELECT review_status, public_documents, COUNT(*) AS n_eprints
FROM (
  SELECT COALESCE(e.review_status, '(null)') AS review_status,
         CASE WHEN COALESCE(SUM(d.security = 'public'), 0) > 0 THEN 'has a public document' ELSE 'no public document' END AS public_documents
  FROM eprint e LEFT JOIN document d ON d.eprintid = e.eprintid
  WHERE e.type = 'thesis' AND e.eprint_status = 'archive' AND (e.review_status IS NULL OR e.review_status <> 'approved')
  GROUP BY e.eprintid, e.review_status
) per_eprint
GROUP BY review_status, public_documents
ORDER BY review_status, public_documents;
