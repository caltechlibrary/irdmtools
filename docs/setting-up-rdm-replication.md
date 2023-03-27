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
