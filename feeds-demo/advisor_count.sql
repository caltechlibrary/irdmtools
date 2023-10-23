WITH q AS (
	WITH t AS (
	SELECT
		jsonb_build_object(
			'rdmid', _key, 
			'clpid', jsonb_path_query(src::jsonb->'thesis_advisor'->'items', '$[*].id')
		) AS obj
	FROM thesis
	ORDER BY _key
	)
	SELECT
		obj::jsonb->>'rdmid' AS rdmid,
		obj::jsonb->>'clpid' AS clpid
	FROM t
	ORDER BY clpid, rdmid
)
SELECT
	jsonb_build_object(
		'clpid', clpid,
		'advisor_count', COUNT(*) 
	) AS obj
FROM q
GROUP BY clpid
ORDER BY clpid
;
