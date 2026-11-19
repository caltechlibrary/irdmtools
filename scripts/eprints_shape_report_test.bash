#!/bin/bash
#
# eprints_shape_report_test.bash tests eprints_shape_report.bash without any
# network access. A fake ssh that always fails is put first on PATH, so a test
# passes only if the script under test never reaches for the network in the
# modes that must not (--check, --print, and every refusal).
#
# Usage: eprints_shape_report_test.bash
# Exit: 0 when every test passes, 1 otherwise.

HERE="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="${HERE}/eprints_shape_report.bash"
SQL="${HERE}/eprints_shape_report.sql"
WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT

# A fake ssh: any call is a test failure.
mkdir -p "${WORK}/bin"
cat >"${WORK}/bin/ssh" <<'FAKE'
#!/bin/bash
echo "FAKE-SSH-CALLED $*" >&2
exit 99
FAKE
chmod +x "${WORK}/bin/ssh"
export PATH="${WORK}/bin:${PATH}"

PASS=0
FAIL=0
check() {
	# check NAME EXPECTED_EXIT ACTUAL_EXIT
	if [ "$2" = "$3" ]; then
		PASS=$((PASS + 1))
		echo "ok   - $1"
	else
		FAIL=$((FAIL + 1))
		echo "FAIL - $1 (expected exit $2, got $3)"
	fi
}
check_contains() {
	# check_contains NAME NEEDLE HAYSTACK
	if printf '%s' "$3" | grep -q -F -- "$2"; then
		PASS=$((PASS + 1))
		echo "ok   - $1"
	else
		FAIL=$((FAIL + 1))
		echo "FAIL - $1 (output lacks: $2)"
	fi
}

# 1. --help exits 0 and says what the tool will and will not do.
OUT="$("${SCRIPT}" --help 2>&1)"
check "help exits 0" 0 $?
check_contains "help says counts only" "counts only" "${OUT}"
check_contains "help documents exit 65" "65" "${OUT}"

# 2. --check lists the sections, reaches no network, and exits 0.
OUT="$("${SCRIPT}" --check 2>&1)"
check "check on the real SQL exits 0" 0 $?
N="$(printf '%s\n' "${OUT}" | grep -c '^section ')"
if [ "${N}" -ge 8 ]; then
	PASS=$((PASS + 1))
	echo "ok   - check lists at least 8 sections (${N})"
else
	FAIL=$((FAIL + 1))
	echo "FAIL - check lists at least 8 sections (got ${N})"
fi
case "${OUT}" in
*FAKE-SSH-CALLED*)
	FAIL=$((FAIL + 1))
	echo "FAIL - check must not call ssh"
	;;
*)
	PASS=$((PASS + 1))
	echo "ok   - check does not call ssh"
	;;
esac

# 3. --print shows the SQL and reaches no network.
OUT="$("${SCRIPT}" --print 2>&1)"
check "print exits 0" 0 $?
check_contains "print shows a SELECT" "SELECT" "${OUT}"

# 4. The guard refuses anything that is not a read-only SELECT or WITH.
printf -- '-- SECTION: bad\nINSERT INTO eprint VALUES (1);\n' >"${WORK}/insert.sql"
SHAPE_SQL="${WORK}/insert.sql" "${SCRIPT}" --check >/dev/null 2>&1
check "an INSERT is refused with 65" 65 $?

printf -- '-- SECTION: bad\nSELECT 1; DROP TABLE eprint;\n' >"${WORK}/drop.sql"
SHAPE_SQL="${WORK}/drop.sql" "${SCRIPT}" --check >/dev/null 2>&1
check "a second statement that is a DROP is refused with 65" 65 $?

# 5. The guard refuses columns that identify a person or a file.
for COL in contact_email creators_email password username name_family name_given filename thesis_advisor_email; do
	printf -- '-- SECTION: bad\nSELECT %s FROM eprint;\n' "${COL}" >"${WORK}/col.sql"
	SHAPE_SQL="${WORK}/col.sql" "${SCRIPT}" --check >/dev/null 2>&1
	check "selecting ${COL} is refused with 65" 65 $?
done

# 6. A comment that merely names a forbidden word is not a violation.
printf -- '-- SECTION: ok\n-- never select contact_email or filename\nSELECT COUNT(*) FROM eprint;\n' >"${WORK}/comment.sql"
SHAPE_SQL="${WORK}/comment.sql" "${SCRIPT}" --check >/dev/null 2>&1
check "a comment naming a forbidden word is allowed" 0 $?

# 7. A missing SQL file is 66; a bad host or database name is a usage error (2).
SHAPE_SQL="${WORK}/does-not-exist.sql" "${SCRIPT}" --check >/dev/null 2>&1
check "a missing SQL file is 66" 66 $?
"${SCRIPT}" -H 'bad host;rm' >/dev/null 2>&1
check "a host with a space and a semicolon is 2" 2 $?
"${SCRIPT}" -d 'db; DROP' >/dev/null 2>&1
check "a database name with a semicolon is 2" 2 $?
"${SCRIPT}" --no-such-flag >/dev/null 2>&1
check "an unknown flag is 2" 2 $?
"${SCRIPT}" -s 999 --check >/dev/null 2>&1
check "a section number out of range is 2" 2 $?

# 8. With a reachable-looking host but a failing ssh, the run reports 69 and
#    does not claim success.
"${SCRIPT}" -s 1 >/dev/null 2>&1
check "an unreachable host is 69" 69 $?

# 8b. The statement check is its own rule: SHOW and SET are not on the keyword
#     list, so only the "starts with SELECT or WITH" check can catch them.
printf -- '-- SECTION: bad\nSELECT 1; SHOW TABLES;\n' >"${WORK}/show.sql"
SHAPE_SQL="${WORK}/show.sql" "${SCRIPT}" --check >/dev/null 2>&1
check "a SHOW statement is refused with 65" 65 $?
printf -- '-- SECTION: bad\nSET @a = 1;\nSELECT @a;\n' >"${WORK}/set.sql"
SHAPE_SQL="${WORK}/set.sql" "${SCRIPT}" --check >/dev/null 2>&1
check "a SET statement is refused with 65" 65 $?

# 8c. A run really goes out in a read-only MySQL session, in batch mode, and a
#     run that works exits 0. A second fake ssh records its arguments.
mkdir -p "${WORK}/bin2"
cat >"${WORK}/bin2/ssh" <<'FAKE2'
#!/bin/bash
echo "$*" >>"${SSH_LOG}"
cat >/dev/null
echo "+---+"
echo "| 1 |"
echo "+---+"
exit 0
FAKE2
chmod +x "${WORK}/bin2/ssh"
SSH_LOG="${WORK}/ssh.log" PATH="${WORK}/bin2:${PATH}" "${SCRIPT}" -s 1 >"${WORK}/run.out" 2>&1
check "a run with a working ssh exits 0" 0 $?
SSHARGS="$(cat "${WORK}/ssh.log" 2>/dev/null)"
check_contains "the run asks for a read-only session" "SET SESSION TRANSACTION READ ONLY" "${SSHARGS}"
check_contains "the run uses BatchMode" "BatchMode=yes" "${SSHARGS}"
check_contains "the report says how many sections ran" "Sections run: 1, failed: 0" "$(cat "${WORK}/run.out")"

# 8d. The keyword list guards what can hide inside a SELECT: writing a file and
#     stalling the server.
printf -- '-- SECTION: bad\nSELECT eprintid INTO OUTFILE "/tmp/x" FROM eprint;\n' >"${WORK}/outfile.sql"
SHAPE_SQL="${WORK}/outfile.sql" "${SCRIPT}" --check >/dev/null 2>&1
check "SELECT INTO OUTFILE is refused with 65" 65 $?
printf -- '-- SECTION: bad\nSELECT SLEEP(30);\n' >"${WORK}/sleep.sql"
SHAPE_SQL="${WORK}/sleep.sql" "${SCRIPT}" --check >/dev/null 2>&1
check "SELECT SLEEP is refused with 65" 65 $?

# 8e. A query that MySQL rejects is a bug in the SQL (70), not an unreachable
#     host (69), and the report still says which section failed.
mkdir -p "${WORK}/bin3"
cat >"${WORK}/bin3/ssh" <<'FAKE3'
#!/bin/bash
cat >/dev/null
echo "ERROR 1054 (42S22) at line 1: Unknown column 'nope' in 'field list'" >&2
exit 1
FAKE3
chmod +x "${WORK}/bin3/ssh"
PATH="${WORK}/bin3:${PATH}" "${SCRIPT}" -s 1 >"${WORK}/err.out" 2>&1
check "a MySQL query error is 70" 70 $?
check_contains "a failed section is named in the report" "FAILED (exit 1)" "$(cat "${WORK}/err.out")"
check_contains "the failure count is reported" "Sections run: 1, failed: 1" "$(cat "${WORK}/err.out")"

# 9. Every query in the real SQL file states small-cell suppression where it
#    groups by a value: the marker must appear in the file.
if grep -q "<5" "${SQL}"; then
	PASS=$((PASS + 1))
	echo "ok   - the SQL suppresses small cells"
else
	FAIL=$((FAIL + 1))
	echo "FAIL - the SQL has no small-cell suppression"
fi

echo
echo "passed ${PASS}, failed ${FAIL}"
if [ "${FAIL}" -ne 0 ]; then
	exit 1
fi
exit 0
