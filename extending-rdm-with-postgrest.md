
# Extending RDM with PostgREST

Invenio RDM comes with a JSON API but it has some limitations. An alternative JSON API can easily be created using PostgREST. 

## Task, get a JSON array containing all record ids and updated timestamps

RDM does not provide an easy mechanism to return all record ids. I will describe the steps needed to create a "/updated_record_ids" end point that accesses a RDM repository database called "caltechauthors".  I am going to implement the "/updated_record_ids" endpoint by creating an "irdm.updated_record_ids" view in SQL. I will grant access to that view using to the "irdm_anonymous" role.  The view will return a list of objects that have the record id and updated timestamp from the pidstore_pid table in RDM. The rows are sorted in descending updated order.

While this is a fair amount of work to create a single end point in practice we can easily define more endpoints by creating additional views, functions and procedures and then restarting PostgREST to pickup the updated API. Long hall this is an easy way to build a robust JSON API that meets our needs without modifying RDM itself.


## Setting up access for PostgREST

When RDM creates the repository database defines tables in the "public" schema (public is the default schema for a database). When using PostgREST you want to restrict PostgREST access to only what we need in our extended API. The Postgres way to do this is to create a new "schema" within our repository database.  In this way we can be assured that PostgREST will only allow interactions with the specifics permissions we grant without tripping over RDM's own roles and permissions.  I am calling the schema for our extended JSON API "irdm".  I will create Postgres roles for "irdm" (privileged account) and "irdm_anonymous" (the user of our API). None of RDM's tables are defined within our "irdm" schema but I can create views of the tables I want the extended API to have access to (e.g. SELECT).   The SQL view also becomes a path end point for PostgREST so if we take care of what we name things we can create a rich API to meet our needs without modifying RDM's source code.

NOTE: In our production deployment RDM is containerized, I've configured Postgres's port to be reachable outside the container on the standard Postgres port "5432".

In this example our repository database is called "caltechauthors" but you should change this value to reflect you're RDM deployment.

First we use the "psql" client to connect to our database.

~~~
psql
--
-- Connect to caltechauthors database.
--
\c caltechauthors
~~~

RDM has already create our database but we need to add our restricted schema for use by PostgREST. The scheme name I've chosen is "irdm".

~~~
DROP SCHEMA IF EXISTS irdm CASCADE;
CREATE SCHEMA irdm;
~~~

Now I will define an anonymous and privileged roles for use by PostgREST. I am going to name our roles "irdm_anonymous" and "irdm". Also replace "PASSWORD_GOES_HERE" with a suitable password.

~~~
--
-- Create role "irdm_anonymous"
--
DROP ROLE IF EXISTS irdm_anonymous;
CREATE ROLE irdm_anonymous NOLOGIN;

--
-- Create our privileged role. 
-- NOTE: replace 'PASSWORD_GOES_HERE' with an appropriate password.
--
DROP ROLE IF EXISTS irdm;
CREATE ROLE irdm NOINHERIT LOGIN PASSWORD 'PASSWORD_GOES_HERE';

--
-- Grant minimal access for connecting to PostgREST.
--
GRANT irdm_anonymous TO irdm;
GRANT USAGE ON SCHEMA irdm TO irdm_anonymous;
GRANT USAGE ON SCHEMA public TO irdm;
~~~

Notice that our "irdm" role gets "USAGE" on "public" but our "irdm_anonymous" does not. The "irdm_anonymous" role only gets "USAGE" privileges on "irdm" schema. This is one way to restrict what is exposed in our PostgREST JSON API. 

The initial purpose of our extended JSON API is to efficiently provide a list of ALL record_ids and their updated timestamp. I am going to do that using an SQL view called "irdm.updated_record_ids". This will become the "/updated_record_ids" end point visible via PostgREST.

~~~
--
-- This view becomes the "/updated_record_ids" end point provided by PostgREST.
--
CREATE OR REPLACE VIEW irdm.updated_record_ids AS
  SELECT pid_value AS record_id, pidstore_pid.updated AS updated 
  FROM rdm_records_metadata JOIN pidstore_pid ON
    (rdm_records_metadata.id = pidstore_pid.object_uuid AND
     pidstore_pid.pid_type = 'recid')
  ORDER BY pidstore_pid.updated DESC;
~~~

Now allow our anonymous user to use this view.

~~~
--
-- Now that we have created some view(s) it is time to set the permissions.
--
GRANT SELECT ON irdm.updated_record_ids TO irdm_anonymous;
~~~

At this point we should be ready to create our PostgREST configuration file and start of PostgREST. For the purposes of this exercise I've called the PostgREST configuration file "postgrest.conf".

~~~
db-uri = "postgres://irdm:PASSWORD_GOES_HERE@localhost:5432/caltechauthors"
db-schemas = "irdm"
db-anon-role = "irdm_anonymous"
~~~

NOTE: In that you should replace "PASSWORD_GOES_HERE" with the password you assigned to the "irdm" role.

We can started up our PostgREST service using the configuration file we created, "postgrest.conf".

~~~
postgrest postgrest.conf
~~~

Now try pointing your web browser at <http://localhost:3000/updated_record_ids>. Assuming you already have a populated RDM repository you should see a JSON array of objects. Each object contains a "record_id" and "updated" attributes.

## Conclusion

As you find you need more end points not provided by RDM consider creating them via SQL views and functions and then accessing them via PostgREST. 



