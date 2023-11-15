COPY (
    WITH t AS (
    	SELECT json->>'id' AS id,
    		jsonb_path_query(json->'metadata'->'contributors', '$[*].role')->>'id' AS role_id,
    		jsonb_path_query(json->'metadata'->'contributors', '$[*].person_or_org')->>'name' AS local_group
    	FROM rdm_records_metadata
    ) SELECT id, role_id, local_group
    FROM t
    WHERE 
    	role_id LIKE 'researchgroup' AND
    	local_group IS NOT NULL
    ORDER BY local_group
) TO STDOUT  DELIMITER ',' CSV HEADER;
