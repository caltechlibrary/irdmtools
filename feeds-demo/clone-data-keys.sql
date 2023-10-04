--
-- clone-authors-keys.sql select all the public authors records so they can
-- be cloned into the published CaltechAUTHORS.ds repository zip file.
--
SELECT to_json(_key) FROM data WHERE src->>'eprint_status' = 'archive';

