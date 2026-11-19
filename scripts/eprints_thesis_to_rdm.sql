-- eprints_thesis_to_rdm.sql builds one RDM-shaped JSON object per CaltechTHESIS
-- EPrints record (irdmtools DR-0010, caltechthesis plan Phase 1). Run it with
-- the mysql client and --raw, or through eprint2rdm's SQL mode.
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
    'custom_fields', CASE WHEN JSON_LENGTH(e.cf) = 0 THEN NULL ELSE e.cf END
))
FROM (
    SELECT x.*,
        -- cf holds the custom fields; JSON_MERGE_PATCH('{}', ...) drops every member
        -- whose value is NULL, so a field with nothing to say is left out.
        JSON_MERGE_PATCH('{}', JSON_OBJECT(
            'thesis:thesis', CASE WHEN JSON_LENGTH(x.thesis_cf) = 0 THEN NULL ELSE x.thesis_cf END,
            -- EPrints records divisions only; the ids go in as a list, as caltech:groups does
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
