%doi2rdm-warnings(1) irdmtools user manual | version 0.1.0 6ccc3239
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

doi2rdm-warnings - When part of a title or abstract cannot be converted cleanly, the record is still written with...

# SYNOPSIS

doi2rdm -help warnings

# DESCRIPTION

When part of a title or abstract cannot be converted cleanly, the record is
still written with the best text available and doi2rdm writes one line
for each such case to standard error (stderr):

~~~
doi2rdm: DOI: field: code: fragment
~~~

field is "title", "additional title" or "description". code is one of
"parse-fallback" (the markup could not be parsed and was kept as text),
"script-unmapped" (a subscript or superscript has no Unicode form),
"math-linear" (a formula was written on one line, for example Z_N),
"math-unconverted" (a formula element was left as text) and
"multiple-titles" (CrossRef gave more than one title; the first is used and the
fragment is the others). fragment is the part of the source the warning is
about. Control characters in a fragment, such as terminal escape sequences,
are replaced by spaces, so a warning is always one line of printable text.
Warnings never change the JSON written to standard output or the exit
status. With -diff only the warnings for the current record are written.

# SEE ALSO

doi2rdm(1)
