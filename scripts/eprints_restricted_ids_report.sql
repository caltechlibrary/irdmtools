-- eprints_restricted_ids_report.sql
--
-- Lists the eprintids (record numbers only, nothing else about the record) of
-- archive thesis eprints whose review_status says they are not released:
-- withheld, notapproved, review and gradoffice (the author's rulings, 2026-11-19). The ids go into a
-- TODO for ep3StaticSite, which serves them today. Same runner as the others:
--
--   SHAPE_SQL=eprints_restricted_ids_report.sql ./eprints_shape_report.bash

-- SECTION: archive thesis eprints with review_status withheld, notapproved, review or gradoffice (eprintid)
SELECT review_status, eprintid
FROM eprint
WHERE type = 'thesis' AND eprint_status = 'archive' AND review_status IN ('withheld', 'notapproved', 'review', 'gradoffice')
ORDER BY review_status, eprintid;
