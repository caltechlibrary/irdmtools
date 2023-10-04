--
-- thesis-updated.sql build a JSON view of updated keys and dates. NOTE this is the updated
-- date from the collection, not the EPRints datestamp. Might change in the future.
-- Query needs to pas the repository table name as a parameter.
SELECT jsonb_build_object(
	'_Key',
	_key,
	'updated',
	updated
)
FROM thesis
ORDER BY _key;
