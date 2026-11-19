-- eprints_thesis_to_rdm.sql builds one RDM-shaped JSON object per CaltechTHESIS
-- EPrints record (irdmtools DR-0010, caltechthesis plan Phase 1). Run it with
-- the mysql client and --raw, or through eprint2rdm's SQL mode.
WITH doc_class AS (
    -- One row per document with the ep3StaticSite file rule applied (brief ruling 17).
    -- class is: excluded (an EPrints thumbnail or index document, by the main file's name
    -- without its extension), non_public (the eprint is not in the archive, or the security
    -- is not public: internal, validuser, or none), embargoed (public, but an embargo date
    -- lies in the future or cannot be read, so it is treated as active), or public.
    SELECT z.*,
        CASE
            WHEN z.is_thumb THEN 'excluded'
            WHEN z.eprint_status IS NULL OR z.eprint_status <> 'archive' THEN 'non_public'
            WHEN z.security IS NULL OR z.security <> 'public' THEN 'non_public'
            WHEN z.embargo_year IS NOT NULL AND (z.embargo_cmp IS NULL OR z.embargo_cmp > CURDATE()) THEN 'embargoed'
            ELSE 'public'
        END AS class
    FROM (
        SELECT d.docid, d.eprintid, d.pos, d.format, d.security, d.main, d.content, d.mime_type,
            e.eprint_status,
            d.date_embargo_year AS embargo_year,
            -- the years 2200 and 9999 are the sentinels for a permanent embargo (author's ruling)
            COALESCE(d.date_embargo_year >= 2200, 0) AS permanent,
            LOWER(REGEXP_REPLACE(COALESCE(d.main, ''), '\\.[^.]+$', ''))
                IN ('lightbox', 'preview', 'medium', 'small', 'indexcodes') AS is_thumb,
            -- the embargo date as EPrints holds it: year, year-month or year-month-day
            CASE WHEN d.date_embargo_year IS NULL THEN NULL
                 WHEN d.date_embargo_month IS NULL THEN CAST(d.date_embargo_year AS CHAR)
                 WHEN d.date_embargo_day IS NULL THEN CONCAT(d.date_embargo_year, '-', LPAD(d.date_embargo_month, 2, '0'))
                 ELSE CONCAT(d.date_embargo_year, '-', LPAD(d.date_embargo_month, 2, '0'), '-', LPAD(d.date_embargo_day, 2, '0')) END AS embargo_d,
            -- a day for comparing only: a missing month or day counts as the first; an
            -- unreadable date is NULL
            STR_TO_DATE(CONCAT(d.date_embargo_year, '-', COALESCE(d.date_embargo_month, 1), '-',
                COALESCE(d.date_embargo_day, 1)), '%Y-%m-%d') AS embargo_cmp
        FROM document d
        JOIN eprint e ON e.eprintid = d.eprintid
    ) z
)
SELECT JSON_MERGE_PATCH('{}', JSON_OBJECT(
    'metadata', JSON_OBJECT(
        'identifiers', JSON_ARRAY(JSON_OBJECT('scheme', 'eprintid', 'identifier', CAST(e.eprintid AS CHAR))),
        -- typed dates (the local date types, EPrints names minus thesis_); thesis
        -- submitted and defended dates live in thesis:thesis instead
        'dates', (
            SELECT JSON_ARRAYAGG(JSON_OBJECT('date', t.d, 'type', JSON_OBJECT('id', t.type_id)))
            FROM (
                SELECT 'approved_date' AS type_id, e.approved_d AS d
                UNION ALL SELECT 'gradofc_approval_date', e.gradofc_d
                UNION ALL SELECT 'degree_date', e.degree_d
            ) t
            WHERE t.d IS NOT NULL
        ),
        -- Advisors first, then the committee, each in EPrints list order. A row with no
        -- role takes its table's default (caltechthesis brief ruling 17): an advisor row
        -- is an advisor, or a co_advisor when the eprint has several advisor rows; a
        -- committee row is a committee_member.
        'contributors', COALESCE((
            SELECT JSON_ARRAYAGG(JSON_OBJECT(
                'person_or_org', JSON_OBJECT(
                    'type', 'personal',
                    'family_name', c.family,
                    'given_name', c.given,
                    'name', CONCAT_WS(', ', c.family, c.given)),
                'role', JSON_OBJECT('id', c.role_id)))
            FROM (
                SELECT n.eprintid, 0 AS grp, n.pos,
                    n.thesis_advisor_name_family AS family,
                    n.thesis_advisor_name_given AS given,
                    CASE r.thesis_advisor_role
                        WHEN 'advisor' THEN 'advisor'
                        WHEN 'co-advisor' THEN 'co_advisor'
                        ELSE CASE WHEN (SELECT COUNT(*) FROM eprint_thesis_advisor_name n2
                                        WHERE n2.eprintid = n.eprintid) > 1
                                  THEN 'co_advisor' ELSE 'advisor' END
                    END AS role_id
                FROM eprint_thesis_advisor_name n
                LEFT JOIN eprint_thesis_advisor_role r ON r.eprintid = n.eprintid AND r.pos = n.pos
                UNION ALL
                SELECT n.eprintid, 1 AS grp, n.pos,
                    n.thesis_committee_name_family,
                    n.thesis_committee_name_given,
                    CASE r.thesis_committee_role
                        WHEN 'chair' THEN 'committee_chair'
                        WHEN 'co-chair' THEN 'committee_co_chair'
                        ELSE 'committee_member'
                    END
                FROM eprint_thesis_committee_name n
                LEFT JOIN eprint_thesis_committee_role r ON r.eprintid = n.eprintid AND r.pos = n.pos
            ) c
            WHERE c.eprintid = e.eprintid
        ), JSON_ARRAY())
    ),
    'access', JSON_OBJECT('record', 'public', 'files', 'public'),
    'files', JSON_OBJECT(
        'enabled', CASE WHEN e.public_files IS NULL THEN CAST('false' AS JSON) ELSE CAST('true' AS JSON) END,
        'entries', e.public_files),
    'custom_fields', CASE WHEN JSON_LENGTH(e.cf) = 0 THEN NULL ELSE e.cf END,
    -- not part of an RDM record: every document and its files, for the loader
    'eprints_sidecar', JSON_OBJECT('documents', e.sidecar_docs)
))
FROM (
    SELECT x.*,
        -- cf holds the custom fields; JSON_MERGE_PATCH('{}', ...) drops every member
        -- whose value is NULL, so a field with nothing to say is left out.
        JSON_MERGE_PATCH('{}', JSON_OBJECT(
            'thesis:thesis', CASE WHEN JSON_LENGTH(x.thesis_cf) = 0 THEN NULL ELSE x.thesis_cf END,
            -- EPrints records divisions only; the ids go in as a list, as caltech:groups does
            -- the latest end among the embargoed documents, at EPrints' precision; a permanent
            -- or unreadable embargo date is never written here
            'caltech:embargo_until', (
                SELECT c.embargo_d FROM doc_class c
                WHERE c.eprintid = x.eprintid AND c.class = 'embargoed'
                    AND c.embargo_cmp IS NOT NULL AND NOT c.permanent
                ORDER BY c.embargo_cmp DESC LIMIT 1
            ),
            'caltech:division', (
                SELECT JSON_ARRAYAGG(JSON_OBJECT('id', d.divisions))
                FROM eprint_divisions d
                WHERE d.eprintid = x.eprintid AND d.divisions IS NOT NULL AND d.divisions <> ''
            )
        )) AS cf
    FROM (
        SELECT e0.*,
            -- a date keeps its precision: year, year-month or year-month-day, never padded
            CASE WHEN e0.thesis_submitted_date_year IS NULL THEN NULL
                WHEN e0.thesis_submitted_date_month IS NULL THEN CAST(e0.thesis_submitted_date_year AS CHAR)
                WHEN e0.thesis_submitted_date_day IS NULL THEN CONCAT(e0.thesis_submitted_date_year, '-', LPAD(e0.thesis_submitted_date_month, 2, '0'))
                ELSE CONCAT(e0.thesis_submitted_date_year, '-', LPAD(e0.thesis_submitted_date_month, 2, '0'), '-', LPAD(e0.thesis_submitted_date_day, 2, '0')) END AS submitted_d,
            CASE WHEN e0.thesis_defense_date_year IS NULL THEN NULL
                WHEN e0.thesis_defense_date_month IS NULL THEN CAST(e0.thesis_defense_date_year AS CHAR)
                WHEN e0.thesis_defense_date_day IS NULL THEN CONCAT(e0.thesis_defense_date_year, '-', LPAD(e0.thesis_defense_date_month, 2, '0'))
                ELSE CONCAT(e0.thesis_defense_date_year, '-', LPAD(e0.thesis_defense_date_month, 2, '0'), '-', LPAD(e0.thesis_defense_date_day, 2, '0')) END AS defended_d,
            CASE WHEN e0.thesis_approved_date_year IS NULL THEN NULL
                WHEN e0.thesis_approved_date_month IS NULL THEN CAST(e0.thesis_approved_date_year AS CHAR)
                WHEN e0.thesis_approved_date_day IS NULL THEN CONCAT(e0.thesis_approved_date_year, '-', LPAD(e0.thesis_approved_date_month, 2, '0'))
                ELSE CONCAT(e0.thesis_approved_date_year, '-', LPAD(e0.thesis_approved_date_month, 2, '0'), '-', LPAD(e0.thesis_approved_date_day, 2, '0')) END AS approved_d,
            CASE WHEN e0.gradofc_approval_date_year IS NULL THEN NULL
                WHEN e0.gradofc_approval_date_month IS NULL THEN CAST(e0.gradofc_approval_date_year AS CHAR)
                WHEN e0.gradofc_approval_date_day IS NULL THEN CONCAT(e0.gradofc_approval_date_year, '-', LPAD(e0.gradofc_approval_date_month, 2, '0'))
                ELSE CONCAT(e0.gradofc_approval_date_year, '-', LPAD(e0.gradofc_approval_date_month, 2, '0'), '-', LPAD(e0.gradofc_approval_date_day, 2, '0')) END AS gradofc_d,
            CASE WHEN e0.thesis_degree_date_year IS NULL THEN NULL
                WHEN e0.thesis_degree_date_month IS NULL THEN CAST(e0.thesis_degree_date_year AS CHAR)
                WHEN e0.thesis_degree_date_day IS NULL THEN CONCAT(e0.thesis_degree_date_year, '-', LPAD(e0.thesis_degree_date_month, 2, '0'))
                ELSE CONCAT(e0.thesis_degree_date_year, '-', LPAD(e0.thesis_degree_date_month, 2, '0'), '-', LPAD(e0.thesis_degree_date_day, 2, '0')) END AS degree_d,
            -- the public documents' files, keyed by file name, as RDM's files.entries
            (SELECT JSON_OBJECTAGG(f.filename, JSON_OBJECT('key', f.filename, 'size', f.filesize,
                    'checksum', CONCAT(LOWER(f.hash_type), ':', f.hash), 'mimetype', f.mime_type))
                FROM doc_class c JOIN file f ON f.datasetid = 'document' AND f.objectid = c.docid
                WHERE c.eprintid = e0.eprintid AND c.class = 'public' AND f.filename IS NOT NULL) AS public_files,
            COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT(
                    'docid', c.docid, 'pos', c.pos, 'class', c.class, 'format', c.format,
                    'security', c.security, 'main', c.main, 'content', c.content,
                    'embargo_until', c.embargo_d,
                    'permanent_embargo', CASE WHEN c.permanent THEN CAST('true' AS JSON) ELSE CAST('false' AS JSON) END,
                    'files', (SELECT JSON_ARRAYAGG(JSON_OBJECT('fileid', f.fileid, 'name', f.filename,
                            'size', f.filesize, 'hash', f.hash, 'hash_type', f.hash_type, 'mimetype', f.mime_type))
                        FROM file f WHERE f.datasetid = 'document' AND f.objectid = c.docid)))
                FROM doc_class c WHERE c.eprintid = e0.eprintid), JSON_ARRAY()) AS sidecar_docs,
            JSON_MERGE_PATCH('{}', JSON_OBJECT(
                'university', NULLIF(e0.institution, ''),
                -- for testing, the first division's display name stands in for the department
                'department', (
                    SELECT sn.name_name
                    FROM eprint_divisions d
                    JOIN subject_name_name sn ON sn.subjectid = d.divisions AND sn.pos = 0
                    WHERE d.eprintid = e0.eprintid AND sn.name_name <> ''
                    ORDER BY d.pos
                    LIMIT 1),
                -- the type as the Go crosswalk writes it: phd is PhD, the rest capitalized
                'type', CASE WHEN e0.thesis_type IS NULL OR e0.thesis_type = '' THEN NULL
                             WHEN e0.thesis_type = 'phd' THEN 'PhD'
                             ELSE CONCAT(UPPER(LEFT(e0.thesis_type, 1)), LOWER(SUBSTRING(e0.thesis_type, 2))) END,
                'date_submitted', CASE WHEN e0.thesis_submitted_date_year IS NULL THEN NULL
                WHEN e0.thesis_submitted_date_month IS NULL THEN CAST(e0.thesis_submitted_date_year AS CHAR)
                WHEN e0.thesis_submitted_date_day IS NULL THEN CONCAT(e0.thesis_submitted_date_year, '-', LPAD(e0.thesis_submitted_date_month, 2, '0'))
                ELSE CONCAT(e0.thesis_submitted_date_year, '-', LPAD(e0.thesis_submitted_date_month, 2, '0'), '-', LPAD(e0.thesis_submitted_date_day, 2, '0')) END,
                'date_defended', CASE WHEN e0.thesis_defense_date_year IS NULL THEN NULL
                WHEN e0.thesis_defense_date_month IS NULL THEN CAST(e0.thesis_defense_date_year AS CHAR)
                WHEN e0.thesis_defense_date_day IS NULL THEN CONCAT(e0.thesis_defense_date_year, '-', LPAD(e0.thesis_defense_date_month, 2, '0'))
                ELSE CONCAT(e0.thesis_defense_date_year, '-', LPAD(e0.thesis_defense_date_month, 2, '0'), '-', LPAD(e0.thesis_defense_date_day, 2, '0')) END
            )) AS thesis_cf
        FROM eprint e0
        WHERE e0.type = 'thesis'
    ) x
) e
ORDER BY e.eprintid;
