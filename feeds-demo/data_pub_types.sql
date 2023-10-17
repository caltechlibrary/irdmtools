WITH t AS (
    SELECT src->>'type' AS pub_type FROM data GROUP BY src->>'type' ORDER BY src->>'type'
) SELECT jsonb_extract_path(jsonb_build_array(t.pub_type), '0') FROM t;
