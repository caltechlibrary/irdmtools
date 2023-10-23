WITH q AS (
	WITH t AS (
		SELECT
			jsonb_build_object(
				'rdmid', _key, 
				'orcid', jsonb_path_query(src::jsonb->'creators'->'items', '$[*].orcid')
			) AS obj
		FROM data
		ORDER BY _key
	)
	SELECT
		obj::jsonb->>'rdmid' AS rdmid,
		obj::jsonb->>'orcid' AS orcid
	FROM t
	ORDER BY orcid, rdmid
)
SELECT
	jsonb_build_object(
		'orcid', orcid,
		'data_count', COUNT(*) 
	) AS obj
FROM q
GROUP BY orcid
ORDER BY orcid
;
