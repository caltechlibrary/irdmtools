#!/bin/bash

for C_NAME in authors.ds thesis.ds data.ds groups.ds people.ds; do
	echo -n "Count ${C_NAME}: "
	dataset count "${C_NAME}"
done
