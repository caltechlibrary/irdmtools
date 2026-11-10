#!/usr/bin/env python3
"""READ-ONLY export of the RDM names vocabulary: one JSON object per line.

Runs on the RDM instance as the instance's own user, under its own `uv run`:

    uv run python - OUTFILE [ENVFILE] < people_rdm_export_names.py

Each line is {"rdm_id": <the entry's id>, "metadata": <the stored JSON, untouched>}, so stray
top-level keys (the legacy `id` and `pid`) are preserved for the matcher to see.

Every row of `name_metadata` is exported. The names service creates entries with NO
pidstore_pid row and no `id` in the stored JSON (found on the test instance, 2026-11-10), so a
join through pidstore_pid silently drops everything the sync itself created. An entry without a
pid row takes its one clpid as its id (the sync creates entries with id equal to the clpid,
cold DR-0031), and falls back to the row's uuid when it has none or several.

Safe by construction: a READ ONLY transaction and a single SELECT. The output file is mode 0600.
Also imported by people_rdm_apply.py, which is why nothing here runs at import time.
"""
import json
import os
import sys

ENTRIES_SQL = """
SELECT m.id, p.pid_value, m.json
FROM name_metadata m
LEFT JOIN pidstore_pid p ON p.object_uuid = m.id AND p.pid_type = 'names'
WHERE m.json IS NOT NULL
ORDER BY p.pid_value NULLS LAST, m.id
"""


def entry_from_row(uuid, pid_value, doc):
    """Turn one database row into a matcher entry {"rdm_id", "metadata"}.

    @param uuid {str} the row's primary key
    @param pid_value {str} the pidstore_pid value, or None when no pid row points at the row
    @param doc {dict|str} the stored JSON, or its text
    @returns {dict} the entry; the stored document is returned untouched
    @example
        entry_from_row("0b1c...", None, {"identifiers": [{"scheme": "clpid", "identifier": "Doe-J"}]})
    """
    if isinstance(doc, str):
        doc = json.loads(doc)
    if pid_value is not None:
        return {"rdm_id": pid_value, "metadata": doc}
    clpids = [i.get("identifier") for i in doc.get("identifiers") or []
              if i.get("scheme") == "clpid" and i.get("identifier")]
    return {"rdm_id": clpids[0] if len(clpids) == 1 else str(uuid), "metadata": doc}


def fetch_entries(db, text):
    """Read every names entry inside a READ ONLY transaction, then end it.

    @param db {object} the Invenio database object
    @param text {callable} sqlalchemy.text
    @returns {list} matcher entries
    """
    db.session.execute(text("SET TRANSACTION READ ONLY"))
    rows = db.session.execute(text(ENTRIES_SQL)).fetchall()
    db.session.rollback()          # ends the read-only transaction before anything is written
    return [entry_from_row(uuid, pid, doc) for uuid, pid, doc in rows]


def main():
    outfile = sys.argv[1]
    envfile = sys.argv[2] if len(sys.argv) > 2 else "/Sites/caltechauthors/.env"
    # A plain Python process does not read the instance's .env (the invenio CLI does).
    from dotenv import load_dotenv      # a Flask dependency, already installed
    load_dotenv(envfile)
    from invenio_app.factory import create_app
    from invenio_db import db
    from sqlalchemy import text

    with create_app().app_context():
        entries = fetch_entries(db, text)
        old = os.umask(0o077)
        try:
            with open(outfile, "w") as out:
                for e in entries:
                    out.write(json.dumps(e) + "\n")
        finally:
            os.umask(old)
    print(f"{len(entries)} names entries written to {outfile}", file=sys.stderr)


if __name__ == "__main__":
    main()
