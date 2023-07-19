#!/bin/bash

. caltechauthors.env
./reset_migration.bash
./run_migration.bash export test-ids.txt
./run_migration.bash import test-ids.txt
