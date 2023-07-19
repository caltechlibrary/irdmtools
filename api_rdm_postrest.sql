

--
-- This is an example SQL file for extending RDM with additional views and 
-- functions for use with PostgREST.  You should copy and modify it as needed.
-- In this example the database is called "caltechauthors" and the scheme used
-- by RDM is "public" but we want to restrict our view to "irdm" schema in the
-- "caltechauthors" database.
--

--
-- Connect to our database "caltechathors" (as an example)
--
\c caltechauthors

CREATE OR REPLACE VIEW irdm.updated_record_ids AS
	SELECT pid_value AS record_id, pidstore_pid.updated AS updated FROM rdm_records_metadata JOIN pidstore_pid ON (rdm_records_metadata.id = pidstore_pid.object_uuid AND pidstore_pid.pid_type = 'recid') ORDER BY pidstore_pid.updated DESC;


--
-- Now that we have created some view(s) it is time to set the permissions.
--
GRANT SELECT ON irdm.updated_record_ids TO irdm_anonymous;
