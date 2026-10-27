# Captured CrossRef responses

Each `crossref-<DOI>.json` is the unedited response of
`https://api.crossref.org/works/<DOI>` (`/` in the DOI is written `_` in the
file name). The tests read them offline, so they must not be edited or
refreshed from the live service.

- Sixteen were copied unchanged from `crossrefapi/testdata/`, captured from
  `api.crossref.org` on 2026-10-26.
- `crossref-10.1093_mnras_stae1289.json` (issue #81) and
  `crossref-10.47456_cad.astro.v7nespecial.53384.json` (issue #87, a Portuguese
  and an English title) were captured the same way on 2026-10-27, so offline
  versions of `TestIssue81` and `TestIssue87` can exist.
