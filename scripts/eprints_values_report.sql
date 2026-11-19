-- eprints_values_report.sql
--
-- A fifth counts-only report for the CaltechTHESIS EPrints database, for the
-- gap table: the value counts of the small controlled columns the crosswalk has
-- to interpret (review status, metadata visibility, published and full text
-- status, document licence and language), and for the free-text columns only the
-- number of distinct values and a few shape tests, never a value. Thesis eprints
-- only. Same runner as the others:
--
--   SHAPE_SQL=eprints_values_report.sql ./eprints_shape_report.bash

-- SECTION: Controlled columns of thesis eprints (value, then count)
SELECT 'review_status' AS field, COALESCE(review_status, '(null)') AS value, COUNT(*) AS n FROM eprint WHERE type = 'thesis' GROUP BY review_status
UNION ALL SELECT 'metadata_visibility', COALESCE(metadata_visibility, '(null)'), COUNT(*) FROM eprint WHERE type = 'thesis' GROUP BY metadata_visibility
UNION ALL SELECT 'ispublished', COALESCE(ispublished, '(null)'), COUNT(*) FROM eprint WHERE type = 'thesis' GROUP BY ispublished
UNION ALL SELECT 'full_text_status', COALESCE(full_text_status, '(null)'), COUNT(*) FROM eprint WHERE type = 'thesis' GROUP BY full_text_status
UNION ALL SELECT 'refereed', COALESCE(refereed, '(null)'), COUNT(*) FROM eprint WHERE type = 'thesis' GROUP BY refereed
UNION ALL SELECT 'collection', COALESCE(collection, '(null)'), COUNT(*) FROM eprint WHERE type = 'thesis' GROUP BY collection
UNION ALL SELECT 'errata is set', IF(errata IS NULL OR errata = '', 'no', 'yes'), COUNT(*) FROM eprint WHERE type = 'thesis' GROUP BY 2
ORDER BY field, n DESC;

-- SECTION: Free-text columns of thesis eprints: distinct values and shape tests (counts only)
SELECT 'rights' AS what, COUNT(DISTINCT rights) AS n_distinct, COUNT(*) AS n_eprints FROM eprint WHERE type = 'thesis' AND rights <> ''
UNION ALL SELECT 'copyright_statement', COUNT(DISTINCT copyright_statement), COUNT(*) FROM eprint WHERE type = 'thesis' AND copyright_statement <> ''
UNION ALL SELECT 'rights equals copyright_statement', NULL, SUM(rights = copyright_statement) FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'id_number', COUNT(DISTINCT id_number), COUNT(*) FROM eprint WHERE type = 'thesis' AND id_number <> ''
UNION ALL SELECT 'id_number is all digits', NULL, SUM(id_number REGEXP '^[0-9]+$') FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'id_number starts CaltechTHESIS', NULL, SUM(id_number LIKE 'CaltechTHESIS:%') FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'official_url', COUNT(DISTINCT official_url), COUNT(*) FROM eprint WHERE type = 'thesis' AND official_url <> ''
UNION ALL SELECT 'official_url starts https://thesis.library.caltech.edu/', NULL, SUM(official_url LIKE 'https://thesis.library.caltech.edu/%') FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'official_url mentions the eprintid', NULL, SUM(official_url LIKE CONCAT('%/', eprintid, '%')) FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'doi', COUNT(DISTINCT doi), COUNT(*) FROM eprint WHERE type = 'thesis' AND doi <> ''
UNION ALL SELECT 'doi starts 10.7907/', NULL, SUM(doi LIKE '10.7907/%') FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'doi starts 10. but not 10.7907/', NULL, SUM(doi LIKE '10.%' AND doi NOT LIKE '10.7907/%') FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'doi does not start 10.', NULL, SUM(doi <> '' AND doi NOT LIKE '10.%') FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'doi shared by more than one eprint', NULL, (SELECT COUNT(*) FROM (SELECT doi FROM eprint WHERE type = 'thesis' AND doi <> '' GROUP BY doi HAVING COUNT(*) > 1) d)
UNION ALL SELECT 'reviewer', COUNT(DISTINCT reviewer), COUNT(*) FROM eprint WHERE type = 'thesis' AND reviewer <> ''
UNION ALL SELECT 'userid', COUNT(DISTINCT userid), COUNT(*) FROM eprint WHERE type = 'thesis' AND userid IS NOT NULL
UNION ALL SELECT 'classification_code', COUNT(DISTINCT classification_code), COUNT(*) FROM eprint WHERE type = 'thesis' AND classification_code <> ''
UNION ALL SELECT 'thesis_awards', COUNT(DISTINCT thesis_awards), COUNT(*) FROM eprint WHERE type = 'thesis' AND thesis_awards <> ''
UNION ALL SELECT 'note', COUNT(DISTINCT note), COUNT(*) FROM eprint WHERE type = 'thesis' AND note <> ''
UNION ALL SELECT 'suggestions', COUNT(DISTINCT suggestions), COUNT(*) FROM eprint WHERE type = 'thesis' AND suggestions <> ''
UNION ALL SELECT 'keywords', COUNT(DISTINCT keywords), COUNT(*) FROM eprint WHERE type = 'thesis' AND keywords <> ''
UNION ALL SELECT 'keywords with a semicolon', NULL, SUM(keywords LIKE CONCAT('%', CHAR(59), '%')) FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'keywords with a comma and no semicolon', NULL, SUM(keywords LIKE '%,%' AND keywords NOT LIKE CONCAT('%', CHAR(59), '%')) FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'keywords equal to cls', NULL, SUM(TRIM(keywords) = 'cls') FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'succeeds set', NULL, SUM(succeeds IS NOT NULL) FROM eprint WHERE type = 'thesis'
UNION ALL SELECT 'title or abstract holds an HTML entity', NULL, SUM(CONCAT(title, ' ', abstract) REGEXP CONCAT('&#?[a-zA-Z0-9]+', CHAR(59))) FROM eprint WHERE type = 'thesis';

-- SECTION: List tables: the value kinds the crosswalk has to interpret (value, then count of rows)
SELECT 'related_url_type' AS field, COALESCE(t.related_url_type, '(null)') AS value, COUNT(*) AS n FROM eprint_related_url_type t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis' GROUP BY t.related_url_type
UNION ALL SELECT 'other_numbering_system_name', COALESCE(t.other_numbering_system_name, '(null)'), COUNT(*) FROM eprint_other_numbering_system_name t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis' GROUP BY t.other_numbering_system_name
UNION ALL SELECT 'funders_agency distinct', CAST(COUNT(DISTINCT t.funders_agency) AS CHAR), COUNT(*) FROM eprint_funders_agency t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL SELECT 'projects distinct', CAST(COUNT(DISTINCT t.projects) AS CHAR), COUNT(*) FROM eprint_projects t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL SELECT 'referencetext distinct', CAST(COUNT(DISTINCT t.referencetext) AS CHAR), COUNT(*) FROM eprint_referencetext t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
UNION ALL SELECT 'alt_title distinct', CAST(COUNT(DISTINCT t.alt_title) AS CHAR), COUNT(*) FROM eprint_alt_title t JOIN eprint e ON e.eprintid = t.eprintid WHERE e.type = 'thesis'
ORDER BY field, n DESC;

-- SECTION: Document columns the crosswalk has to interpret (value, then count of documents of thesis eprints)
SELECT 'license' AS field, COALESCE(d.license, '(null)') AS value, COUNT(*) AS n FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis' GROUP BY d.license
UNION ALL SELECT 'language', COALESCE(d.language, '(null)'), COUNT(*) FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis' GROUP BY d.language
UNION ALL SELECT 'format', COALESCE(d.format, '(null)'), COUNT(*) FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis' GROUP BY d.format
UNION ALL SELECT 'formatdesc distinct', CAST(COUNT(DISTINCT d.formatdesc) AS CHAR), COUNT(*) FROM document d JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis'
UNION ALL SELECT 'file hash_type', COALESCE(f.hash_type, '(null)'), COUNT(*) FROM file f JOIN document d ON f.datasetid = 'document' AND f.objectid = d.docid JOIN eprint e ON e.eprintid = d.eprintid WHERE e.type = 'thesis' GROUP BY f.hash_type
UNION ALL SELECT 'file datasetid', COALESCE(f.datasetid, '(null)'), COUNT(*) FROM file f GROUP BY f.datasetid
ORDER BY field, n DESC;
