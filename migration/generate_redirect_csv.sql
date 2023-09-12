--
-- This SQL file generates a CSV file suitable for mapping redirects for imported EPrint ids.
-- The resulting file is is written to standard out in CSV format.
-- 
COPY (
	SELECT (t1.identifiers ->> 'identifier')::DECIMAL AS eprintid, t1.rdmid AS rdmid, t1.record_status AS record_status
	FROM (SELECT json ->> 'id' AS rdmid, json -> 'access' ->> 'record' AS record_status, 
       		jsonb_array_elements(json -> 'metadata' -> 'identifiers') AS identifiers
		FROM rdm_records_metadata
		) AS t1 
	WHERE (t1.identifiers ->> 'scheme' LIKE 'eprintid')
	AND (t1.record_status LIKE 'public') 
	ORDER BY (t1.identifiers ->> 'identifier')::DECIMAL
) 
TO STDOUT (FORMAT CSV, HEADER);
