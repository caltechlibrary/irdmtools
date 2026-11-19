#!/usr/bin/env bash
#
# scratch_mysql.bash starts, stops and resets a throwaway MySQL 8 that holds the
# structure of the CaltechTHESIS EPrints database (testdata/eprints_thesis_schema.sql)
# and no rows. It listens on 127.0.0.1 only, keeps its files in one directory,
# and has a passwordless root that is reachable only from this machine.
# Tests write their own rows; nothing here ever reads production.
#
# Usage: scratch_mysql.bash start|stop|reset|status|dsn
#
# Environment: SCRATCH_MYSQL_BIN (default /usr/local/mysql/bin), SCRATCH_MYSQL_DIR
# (default ${TMPDIR:-/tmp}/irdmtools-scratch-mysql), SCRATCH_MYSQL_PORT (default 33306).
#
# Exit status: 0 ok, 1 not running (status), 2 usage, 66 mysqld or the schema
# file is missing, 69 the server did not come up in time, 70 anything else.

set -u

BIN="${SCRATCH_MYSQL_BIN:-/usr/local/mysql/bin}"
DIR="${SCRATCH_MYSQL_DIR:-${TMPDIR:-/tmp}/irdmtools-scratch-mysql}"
PORT="${SCRATCH_MYSQL_PORT:-33306}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA="${SCRATCH_MYSQL_SCHEMA:-${HERE}/../testdata/eprints_thesis_schema.sql}"
DB="eprints_scratch"
SOCK="${SCRATCH_MYSQL_SOCK:-/tmp/irdmtools-scratch-${PORT}.sock}"
PIDFILE="${DIR}/mysqld.pid"

usage() {
	echo "usage: $(basename "$0") start|stop|reset|status|dsn" >&2
	exit 2
}

dsn() {
	echo "root@tcp(127.0.0.1:${PORT})/${DB}"
}

running() {
	[ -f "${PIDFILE}" ] && kill -0 "$(cat "${PIDFILE}")" 2>/dev/null
}

client() {
	"${BIN}/mysql" --no-defaults -uroot --protocol=socket -S "${SOCK}" "$@"
}

cmd_start() {
	[ -x "${BIN}/mysqld" ] || { echo "mysqld not found in ${BIN}" >&2; exit 66; }
	[ -f "${SCHEMA}" ] || { echo "schema file not found: ${SCHEMA}" >&2; exit 66; }
	if running; then
		dsn
		return 0
	fi
	local fresh=0
	if [ ! -d "${DIR}/data" ]; then
		mkdir -p "${DIR}" || exit 70
		"${BIN}/mysqld" --no-defaults --initialize-insecure --datadir="${DIR}/data" \
			--basedir="$(dirname "${BIN}")" >"${DIR}/init.log" 2>&1 || {
			echo "initialize failed, see ${DIR}/init.log" >&2
			exit 70
		}
		fresh=1
	fi
	"${BIN}/mysqld" --no-defaults --datadir="${DIR}/data" --basedir="$(dirname "${BIN}")" \
		--bind-address=127.0.0.1 --port="${PORT}" --socket="${SOCK}" \
		--mysqlx=OFF --pid-file="${PIDFILE}" --log-error="${DIR}/error.log" \
		>/dev/null 2>&1 &

	for _ in $(seq 1 60); do
		if client -e 'SELECT 1' >/dev/null 2>&1; then
			break
		fi
		sleep 1
	done
	client -e 'SELECT 1' >/dev/null 2>&1 || { echo "server did not come up, see ${DIR}/error.log" >&2; exit 69; }
	if [ "${fresh}" = 1 ] || ! client -N -e "SHOW DATABASES LIKE '${DB}'" | grep -q "${DB}"; then
		client -e "CREATE DATABASE IF NOT EXISTS ${DB} CHARACTER SET utf8mb4" || exit 70
		client "${DB}" <"${SCHEMA}" || { echo "schema load failed" >&2; exit 70; }
	fi
	dsn
}

cmd_stop() {
	if running; then
		client -e 'SHUTDOWN' 2>/dev/null || kill "$(cat "${PIDFILE}")"
		for _ in $(seq 1 30); do
			running || return 0
			sleep 1
		done
		echo "server did not stop" >&2
		exit 70
	fi
}

cmd_reset() {
	running || { cmd_start >/dev/null || exit $?; }
	client -e "DROP DATABASE IF EXISTS ${DB}; CREATE DATABASE ${DB} CHARACTER SET utf8mb4" || exit 70
	client "${DB}" <"${SCHEMA}" || { echo "schema load failed" >&2; exit 70; }
	dsn
}

[ $# -eq 1 ] || usage
case "$1" in
start) cmd_start ;;
stop) cmd_stop ;;
reset) cmd_reset ;;
status)
	if running; then
		echo "running"
	else
		echo "not running"
		exit 1
	fi
	;;
dsn) dsn ;;
*) usage ;;
esac
