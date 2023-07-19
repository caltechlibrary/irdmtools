
--
-- This is an example setup file for configuring Postgres for PostgREST access.
-- Copy and modify the file as needed. The database (caltechauthors and schema 
-- caltechauthors are example file values, replace as appropriate).
--

--
-- Connect to caltechauthors database.
--
\c caltechauthors

--
-- Create role "irdm_anonymous"
--
DROP ROLE IF EXISTS irdm_anonymous;
CREATE ROLE irdm_anonymous NOLOGIN;

--
-- Create our privileged role, not replace 'PASSWORD_GOES_HERE' with
-- an appropriate password.
--
DROP ROLE IF EXISTS irdm;
CREATE ROLE irdm NOINHERIT LOGIN PASSWORD 'PASSWORD_GOES_HERE';

--
-- Give access to the Schema to PostgREST for each role.
--
GRANT irdm_anonymous TO irdm;
GRANT USAGE ON SCHEMA irdm TO irdm_anonymous;
GRANT USAGE ON SCHEMA public TO irdm;



