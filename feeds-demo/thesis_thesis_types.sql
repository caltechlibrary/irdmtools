WITH t AS (
    SELECT src->>'thesis_type' AS thesis_type 
	FROM thesis 
	WHERE src->'thesis_type' IS NOT NULL
	GROUP BY src->>'thesis_type'
	ORDER BY src->>'thesis_type'

) SELECT jsonb_extract_path(jsonb_build_array(t.thesis_type), '0') FROM t;
