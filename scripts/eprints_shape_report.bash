#!/bin/bash
#
# eprints_shape_report.bash runs eprints_shape_report.sql against the
# CaltechTHESIS EPrints database over SSH, one section at a time, and writes a
# Markdown report to standard output.
#
# It reads and reports counts only. See --help.

VERSION="0.0.1"

# Exit codes follow the workspace convention (DR-0014).
EX_OK=0
EX_USAGE=2
EX_DATA=65
EX_NOINPUT=66
EX_UNAVAILABLE=69
EX_INTERNAL=70

HERE="$(cd "$(dirname "$0")" && pwd)"
SQL_FILE="${SHAPE_SQL:-${HERE}/eprints_shape_report.sql}"
HOST="${SHAPE_HOST:-eprints.library.caltech.edu}"
DB="${SHAPE_DB:-caltechthesis}"
MODE="run"
ONLY=""

usage() {
	cat <<'TXT'
NAME

  eprints_shape_report.bash

SYNOPSIS

  eprints_shape_report.bash [-H HOST] [-d DATABASE] [-s N] [--check | --print]

DESCRIPTION

  Runs eprints_shape_report.sql against an EPrints MySQL database over SSH and
  writes a Markdown report to standard output. The report holds counts only:
  how many records, in which states, with which date precision, how many
  documents and files, and which embargo and security patterns. It never selects
  a title, abstract, note, name, email address or file name. A grouped count
  under 5 is shown as '<5'.

  Every query runs in a read-only MySQL session (SET SESSION TRANSACTION READ
  ONLY). Before anything is sent, the SQL file is checked: it may hold only
  SELECT and WITH statements, no write or file-output keywords, and no column
  whose name carries an email address, a password, a user name, a person's name
  or a file name. A file that breaks a rule is refused (exit 65).

  All sections are attempted. A section that fails is reported in the output and
  counted; the exit code is that of the first failure.

OPTIONS

  -H, --host HOST        SSH host (default eprints.library.caltech.edu,
                         or $SHAPE_HOST)
  -d, --db DATABASE      MySQL database (default caltechthesis, or $SHAPE_DB)
  -s, --section N        run only section N
      --check            lint the SQL and list the sections; no network
      --print            lint the SQL and print every section; no network
  -h, --help             show this help
      --version          show the version

  $SHAPE_SQL names a different SQL file (used by the tests).

EXIT STATUS

  0   ok
  2   usage: an unknown option, a bad host or database name, a section out of range
  65  the SQL file was refused by the safety check, or holds no sections
  66  the SQL file is missing
  69  the host or the database could not be reached, or the query timed out
  70  a query failed inside MySQL (a bug in the SQL), or an internal error

EXAMPLE

  eprints_shape_report.bash --check
  eprints_shape_report.bash > 2026-11-20-thesis-eprints-shape.md
TXT
}

die() {
	# die CODE MESSAGE
	printf '%s\n' "$2" >&2
	exit "$1"
}

# --- arguments ---------------------------------------------------------------
while [ $# -gt 0 ]; do
	case "$1" in
	-h | --help | help)
		usage
		exit ${EX_OK}
		;;
	--version)
		echo "eprints_shape_report.bash ${VERSION}"
		exit ${EX_OK}
		;;
	--check) MODE="check" ;;
	--print) MODE="print" ;;
	-H | --host)
		[ $# -ge 2 ] || die ${EX_USAGE} "$1 needs a value"
		HOST="$2"
		shift
		;;
	-d | --db)
		[ $# -ge 2 ] || die ${EX_USAGE} "$1 needs a value"
		DB="$2"
		shift
		;;
	-s | --section)
		[ $# -ge 2 ] || die ${EX_USAGE} "$1 needs a value"
		ONLY="$2"
		shift
		;;
	*) die ${EX_USAGE} "unknown option: $1 (see --help)" ;;
	esac
	shift
done

# A host or database name is letters, digits, dots, dashes and underscores only.
case "${HOST}" in
'' | *[!A-Za-z0-9._-]*) die ${EX_USAGE} "bad host name: ${HOST}" ;;
esac
case "${DB}" in
'' | *[!A-Za-z0-9_]*) die ${EX_USAGE} "bad database name: ${DB}" ;;
esac
if [ -n "${ONLY}" ]; then
	case "${ONLY}" in
	*[!0-9]*) die ${EX_USAGE} "section must be a number: ${ONLY}" ;;
	esac
fi

[ -f "${SQL_FILE}" ] || die ${EX_NOINPUT} "SQL file not found: ${SQL_FILE}"

WORK="$(mktemp -d)" || die ${EX_INTERNAL} "cannot make a work directory"
trap 'rm -rf "${WORK}"' EXIT

# --- the safety check --------------------------------------------------------
# Strip whole-line and trailing comments, then look at what is left.
sed -E 's/(^|[[:space:]])--.*$//' "${SQL_FILE}" >"${WORK}/code.sql"

# 1. Words that identify a person or a file, anywhere in the code.
if grep -i -E -n 'email|password|username|name_family|name_given|filename' "${WORK}/code.sql" >"${WORK}/hits" 2>/dev/null; then
	{
		echo "refused: the SQL names a column that identifies a person or a file:"
		cut -c1-120 "${WORK}/hits"
	} >&2
	exit ${EX_DATA}
fi
# 2. Write, file-output and delay keywords.
if grep -i -w -E -n 'insert|update|delete|drop|alter|create|truncate|grant|revoke|load|outfile|dumpfile|sleep|benchmark' "${WORK}/code.sql" >"${WORK}/hits" 2>/dev/null; then
	{
		echo "refused: the SQL holds a keyword a read-only report must not use:"
		cut -c1-120 "${WORK}/hits"
	} >&2
	exit ${EX_DATA}
fi
# 3. Every statement must start with SELECT or WITH.
awk 'BEGIN { RS = ";"; bad = 0 }
     {
       s = $0
       gsub(/^[ \t\r\n]+/, "", s)
       if (length(s) == 0) next
       head = toupper(substr(s, 1, 6))
       head4 = toupper(substr(s, 1, 4))
       if (head != "SELECT" && head4 != "WITH") {
         printf "statement %d starts with: %s\n", NR, substr(s, 1, 40)
         bad = 1
       }
     }
     END { exit bad }' "${WORK}/code.sql" >"${WORK}/hits" 2>&1 ||
	{
		{
			echo "refused: a statement is not a SELECT or WITH:"
			cat "${WORK}/hits"
		} >&2
		exit ${EX_DATA}
	}

# --- split into sections -----------------------------------------------------
# A line "-- SECTION: title" starts section N, written to sec_NN.sql with its
# title in sec_NN.title. Anything before the first marker is ignored.
awk -v dir="${WORK}" '
  /^-- SECTION:/ {
    n++
    title = $0
    sub(/^-- SECTION:[ \t]*/, "", title)
    file = sprintf("%s/sec_%02d", dir, n)
    print title > (file ".title")
    close(file ".title")
    printf "" > (file ".sql")
    next
  }
  n > 0 { print >> sprintf("%s/sec_%02d.sql", dir, n) }
' "${SQL_FILE}"

NSEC=0
for f in "${WORK}"/sec_*.sql; do
	[ -e "${f}" ] && NSEC=$((NSEC + 1))
done
[ "${NSEC}" -gt 0 ] || die ${EX_DATA} "no '-- SECTION:' lines in ${SQL_FILE}"

if [ -n "${ONLY}" ]; then
	if [ "${ONLY}" -lt 1 ] || [ "${ONLY}" -gt "${NSEC}" ]; then
		die ${EX_USAGE} "section ${ONLY} is out of range (1 to ${NSEC})"
	fi
fi

section_title() { cat "${WORK}/$(printf 'sec_%02d' "$1").title"; }
section_file() { printf '%s/sec_%02d.sql' "${WORK}" "$1"; }

if [ "${MODE}" = "check" ]; then
	i=1
	while [ "${i}" -le "${NSEC}" ]; do
		printf 'section %d: %s\n' "${i}" "$(section_title "${i}")"
		i=$((i + 1))
	done
	echo "ok: ${NSEC} sections, only SELECT and WITH statements, no forbidden words"
	exit ${EX_OK}
fi

if [ "${MODE}" = "print" ]; then
	i=1
	while [ "${i}" -le "${NSEC}" ]; do
		if [ -z "${ONLY}" ] || [ "${ONLY}" = "${i}" ]; then
			printf -- '-- section %d: %s\n' "${i}" "$(section_title "${i}")"
			cat "$(section_file "${i}")"
			echo
		fi
		i=$((i + 1))
	done
	exit ${EX_OK}
fi

# --- run ---------------------------------------------------------------------
TIMEOUT="$(command -v gtimeout || command -v timeout || true)"
REMOTE="mysql --init-command=\"SET SESSION TRANSACTION READ ONLY\" --table --default-character-set=utf8mb4 ${DB}"

printf '# EPrints shape report\n\n'
printf -- '- Generated: %s (UTC)\n' "$(date -u '+%Y-%m-%d %H:%M:%S')"
printf -- '- Host: %s, database: %s\n' "${HOST}" "${DB}"
printf -- '- Counts only. No title, abstract, note, name, email address or file name is selected.\n'
printf -- "- A grouped count under 5 is shown as '<5'. Queries run in a read-only MySQL session.\n"
printf -- '- Dates are year, month and day columns; a date with no month is a year-only date.\n\n'

FAILED=0
FIRST_CODE=${EX_OK}
i=1
while [ "${i}" -le "${NSEC}" ]; do
	if [ -n "${ONLY}" ] && [ "${ONLY}" != "${i}" ]; then
		i=$((i + 1))
		continue
	fi
	printf '## %d. %s\n\n' "${i}" "$(section_title "${i}")"
	if [ -n "${TIMEOUT}" ]; then
		"${TIMEOUT}" 180 ssh -o BatchMode=yes -o ConnectTimeout=10 "${HOST}" "${REMOTE}" <"$(section_file "${i}")" >"${WORK}/out" 2>"${WORK}/err"
	else
		ssh -o BatchMode=yes -o ConnectTimeout=10 "${HOST}" "${REMOTE}" <"$(section_file "${i}")" >"${WORK}/out" 2>"${WORK}/err"
	fi
	RC=$?
	# Drop SSH's post-quantum key exchange notice; keep everything else.
	grep -v -E '^\*\*' "${WORK}/err" >"${WORK}/err.clean"
	if [ "${RC}" -eq 0 ]; then
		printf '```\n'
		cat "${WORK}/out"
		printf '```\n\n'
	else
		FAILED=$((FAILED + 1))
		if grep -q 'ERROR' "${WORK}/err.clean"; then
			CODE=${EX_INTERNAL}
		else
			CODE=${EX_UNAVAILABLE}
		fi
		[ "${FIRST_CODE}" -eq ${EX_OK} ] && FIRST_CODE=${CODE}
		printf 'FAILED (exit %d). The error was:\n\n```\n' "${RC}"
		cut -c1-200 "${WORK}/err.clean" | head -8
		printf '```\n\n'
		{
			printf 'section %d (%s) failed, exit %d\n' "${i}" "$(section_title "${i}")" "${RC}"
			cut -c1-200 "${WORK}/err.clean" | head -4
		} >&2
	fi
	i=$((i + 1))
done

RAN=${NSEC}
[ -n "${ONLY}" ] && RAN=1
printf -- '---\n\nSections run: %d, failed: %d\n' "${RAN}" "${FAILED}"
exit "${FIRST_CODE}"
