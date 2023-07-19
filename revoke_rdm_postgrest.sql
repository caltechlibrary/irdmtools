
--
-- This is an example setup file for configuring Postgres for PostgREST access.
-- Copy and modify the file as needed. The database caltechauthors and schema 
-- irdm are example file values, replace as appropriate.
--

--
-- Connect to caltechauthors database.
--
\c caltechauthors

--
-- Revoke access to the Schema to PostgREST for each role.
--
REVOKE irdm_anonymous FROM caltechauthors;
REVOKE USAGE ON SCHEMA public FROM caltechauthors_anonymous;

--
-- Permissions for tables.
--

-- Revoke access for our anonymous role irdm_anonymous
REVOKE SELECT ON rdm_drafts_files FROM irdm_anonymous;
REVOKE SELECT ON rdm_drafts_metadata FROM irdm_anonymous;
REVOKE SELECT ON rdm_parents_community FROM irdm_anonymous;
REVOKE SELECT ON rdm_parents_metadata FROM irdm_anonymous;
REVOKE SELECT ON rdm_records_files FROM irdm_anonymous;
REVOKE SELECT ON rdm_records_metadata FROM irdm_anonymous;
REVOKE SELECT ON rdm_records_metadata_version FROM irdm_anonymous;
REVOKE SELECT ON rdm_records_secret_links FROM irdm_anonymous;
REVOKE SELECT ON rdm_versions_state FROM irdm_anonymous;

--
-- Remove access to our views
REVOKE SELECT ON updated_record_ids FROM irdm_anonymous;

--
-- Remove role "irdm_anonymous"
--
DROP ROLE IF EXISTS irdm_anonymous;

--
-- Remove our privileged role, not replace 'PASSWORD_GOES_HERE' with
-- an appropriate password.
--
DROP ROLE IF EXISTS irdm;

