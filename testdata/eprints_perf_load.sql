-- eprints_perf_load.sql fills a scratch EPrints database with synthetic thesis
-- rows at about the size of production (13,000 theses, 6 documents each, 1 to 2
-- files per document), to time scripts/eprints_thesis_to_rdm.sql. Every value is
-- generated; nothing is copied from a record. Run it against an empty schema
-- (testdata/eprints_thesis_schema.sql) with cte_max_recursion_depth raised.
SET SESSION cte_max_recursion_depth = 100000;

INSERT INTO eprint (eprintid, eprint_status, type, title, abstract, keywords, institution, thesis_type,
    date_year, date_type, datestamp_year, datestamp_month, datestamp_day,
    thesis_defense_date_year, thesis_defense_date_month, thesis_approved_date_year, thesis_approved_date_month)
WITH RECURSIVE n AS (SELECT 1 AS i UNION ALL SELECT i + 1 FROM n WHERE i < 13000)
SELECT i, 'archive', 'thesis', CONCAT('Synthetic title ', i), CONCAT('Synthetic abstract ', i),
    CONCAT('kw', i MOD 50, '; kw', i MOD 7, '; kw', i MOD 13), 'California Institute of Technology',
    ELT(1 + i MOD 4, 'phd', 'masters', 'engd', 'bachelors'),
    1930 + i MOD 95, 'degree', 2020 + i MOD 6, 1 + i MOD 12, 1 + i MOD 28,
    1930 + i MOD 95, 1 + i MOD 12, 1930 + i MOD 95, 1 + i MOD 12
FROM n;

INSERT INTO eprint_creators_name (eprintid, pos, creators_name_family, creators_name_given)
SELECT eprintid, 0, CONCAT('Family', eprintid), CONCAT('Given', eprintid) FROM eprint;
INSERT INTO eprint_creators_id (eprintid, pos, creators_id) SELECT eprintid, 0, CONCAT('Fam-', eprintid) FROM eprint;
INSERT INTO eprint_creators_orcid (eprintid, pos, creators_orcid)
SELECT eprintid, 0, CONCAT('0000-0000-0000-', LPAD(eprintid, 4, '0')) FROM eprint WHERE eprintid MOD 3 = 0;

INSERT INTO eprint_thesis_advisor_name (eprintid, pos, thesis_advisor_name_family, thesis_advisor_name_given)
SELECT eprintid, 0, CONCAT('Adv', eprintid), 'A' FROM eprint;
INSERT INTO eprint_thesis_advisor_role (eprintid, pos, thesis_advisor_role) SELECT eprintid, 0, 'advisor' FROM eprint WHERE eprintid MOD 2 = 0;
INSERT INTO eprint_thesis_committee_name (eprintid, pos, thesis_committee_name_family, thesis_committee_name_given)
WITH RECURSIVE p AS (SELECT 0 AS pos UNION ALL SELECT pos + 1 FROM p WHERE pos < 3)
SELECT e.eprintid, p.pos, CONCAT('Com', e.eprintid, '-', p.pos), 'C' FROM eprint e CROSS JOIN p;
INSERT INTO eprint_thesis_committee_role (eprintid, pos, thesis_committee_role)
SELECT eprintid, pos, IF(pos = 0, 'chair', 'member') FROM eprint_thesis_committee_name WHERE eprintid MOD 4 <> 0;

INSERT INTO subject_name_name (subjectid, pos, name_name)
WITH RECURSIVE n AS (SELECT 1 AS i UNION ALL SELECT i + 1 FROM n WHERE i < 72)
SELECT CONCAT('subj', i), 0, CONCAT('Subject name ', i) FROM n;
INSERT INTO eprint_divisions (eprintid, pos, divisions) SELECT eprintid, 0, CONCAT('subj', 1 + eprintid MOD 6) FROM eprint;
INSERT INTO eprint_option_major (eprintid, pos, option_major) SELECT eprintid, 0, CONCAT('subj', 7 + eprintid MOD 40) FROM eprint;
INSERT INTO eprint_option_minor (eprintid, pos, option_minor) SELECT eprintid, 0, CONCAT('subj', 7 + eprintid MOD 30) FROM eprint WHERE eprintid MOD 3 = 0;

-- six documents per eprint: a public final, a public supplement, two internal forms, an archival copy
-- with an embargo on a share of them, and a thumbnail
INSERT INTO document (docid, eprintid, pos, format, security, main, content, date_embargo_year, date_embargo_month, date_embargo_day, mime_type)
WITH RECURSIVE p AS (SELECT 1 AS pos UNION ALL SELECT pos + 1 FROM p WHERE pos < 6)
SELECT e.eprintid * 10 + p.pos, e.eprintid, p.pos, 'application/pdf',
    CASE p.pos WHEN 1 THEN 'public' WHEN 2 THEN 'public' WHEN 5 THEN IF(e.eprintid MOD 50 = 0, 'public', 'internal') ELSE 'internal' END,
    CASE p.pos WHEN 6 THEN 'lightbox.jpg' ELSE CONCAT('doc', p.pos, '.pdf') END,
    CASE p.pos WHEN 1 THEN 'final' WHEN 2 THEN 'supplemental' WHEN 5 THEN 'archival' ELSE 'approvalform' END,
    IF(p.pos = 5 AND e.eprintid MOD 50 = 0, 2100, NULL), NULL, NULL, 'application/pdf'
FROM eprint e CROSS JOIN p;

INSERT INTO file (fileid, datasetid, objectid, filename, mime_type, hash, hash_type, filesize)
WITH RECURSIVE f AS (SELECT 0 AS k UNION ALL SELECT 1)
SELECT d.docid * 10 + f.k, 'document', d.docid, CONCAT(d.main, IF(f.k = 0, '', '.txt')), d.mime_type,
    MD5(d.docid * 10 + f.k), 'MD5', 1000
FROM document d CROSS JOIN f WHERE f.k = 0 OR d.pos IN (1, 2);
