
Todo
====

Bugs
----

Next
----

- [ ] Implement a first pass at a harvester
    - [ ] decide is it is an action in irdmutil or a separate program
    - [ ] decide if I'm harvesting to a SQL store using JSON columns or if I am harvesting directly to a dataset collection
    - [ ] if harvesting to a SQL store then I need a dataset dumper too
- [x] Figure out a way to retrieve RDM ids without using the API, e.g. safely SQL via PostgreSQL or create a separate end point. The API max results returned is 10K and we're going to have many more records than that
- [ ] Review [Go-app.dev](https://go-app.dev) and see if it would be useful for GUI tooling around irdmtools


