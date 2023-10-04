#!/bin/bash
for DB_NAME in authors thesis data groups people; do
	dropdb "${DB_NAME}"
	rm -fR "${DB_NAME}.ds"
done
