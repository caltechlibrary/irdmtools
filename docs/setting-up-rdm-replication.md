---
title: "Setting up RDM metadata replication"
author: "R. S. Doiel"
---

# Setting up RDM metadata replication

This document outlines a recipe for replicating RDM metadata using [PostgreSQL logical replication](https://www.postgresql.org/docs/14/logical-replication.html "PostgreSQL manual").

## Problem Out line

1. How do you setup replication when the "publisher" runs in a Docker instance?
2. What needs to be replicated?
    a. Dumping scheme from production PostgreSQL
    b. Generating snapshot to seed our "subscriber"
    c. Run and monitor replicated data?
3. Do we need to create JSON record views of the assemble objects, if so does this have to be done in Python using the RDM ORM or do we have other options?
4. Can we go from replicated database to dataset collections quickly?

## Setting up the publisher

1. Create a "publisher" using the CREATE statement described at https://www.postgresql.org/docs/14/sql-createpublication.html
2. Adjust the containers firewall and PostgreSQL configuration in the container to allow reading the PostgreSQL database from the machine where the data warehouse is implemented
3. Test connections from the "subscribe" machine

## Setting up the subscriber

1. On the "subscriber" system make sure you can connect and browse the "publisher" database
2. Install a snapshot from the "publisher" system
3. Setup a subscription using CREATE per the instructions on https://www.postgresql.org/docs/14/sql-createsubscription.html
4. Montor to make sure data is flowing from the publisher to the subscriber, see https://www.postgresql.org/docs/14/logical-replication-monitoring.html

## Additional References

- https://www.postgresql.org/docs/14/warm-standby.html
