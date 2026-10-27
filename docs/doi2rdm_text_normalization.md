# doi2rdm: titles and abstracts from CrossRef

CrossRef gives titles and abstracts as publishers wrote them: with markup,
character entities, MathML formulas, layout whitespace, and the subtitle in a
separate field. `doi2rdm` converts three fields of a CrossRef record before it
writes the JSON for Invenio RDM, so a title is text and an abstract is HTML that
RDM keeps. The conversion is done by the
[crossrefapi](https://software.library.caltech.edu/crossrefapi/) package; its
[text normalization](https://software.library.caltech.edu/crossrefapi/text_normalization.html)
page has the rules, the full list of conversions and their known limits. This
page is about what `doi2rdm` does with them.

DataCite records are not changed.

## What changes

| Field | What `doi2rdm` writes |
|---|---|
| `metadata.title` | The first CrossRef title as plain text, with the subtitle appended after a colon (after a space only when the title already ends in `.`, `?`, `!` or `:`), unless the title already holds the subtitle. Tags are removed, entities decoded, formulas converted to Unicode and layout whitespace tidied. |
| `metadata.additional_titles` | The other CrossRef titles that add something, as plain text, each with the type `other`. A second title that only differs by capitalization is not added. |
| `metadata.description` | The abstract as HTML that RDM's sanitizer keeps, with formulas as TeX between `\(` and `\)`. The leading "Abstract" heading and the publisher's copyright line are dropped. No abstract, no description. |

The first CrossRef title is always the title. When CrossRef gives several, the
others become additional titles and a `multiple-titles` warning says so (see
Warnings), because `doi2rdm` does not guess which one is in which language.

## Examples

These are real CrossRef records, as `doi2rdm` now converts them.

| DOI | Before | After |
|---|---|---|
| 10.1051/0004-6361/202660273 | title `<i>Euclid</i>` and a line break, then `preparation`; the subtitle was dropped | `Euclid preparation: CII. Non-Gaussianity of two-point statistics likelihood: Parameter inference with a non-Gaussian likelihood in Fourier and configuration space` |
| 10.1103/physrevb.85.144303 | `…loop model with<mml:math …>…</mml:math>current-current interactions` | `…loop model with 1/r² current-current interactions` |
| 10.1103/physrevlett.134.141902 | `Phases of Theories with` and a `<mml:math>` element, then `1-Form Symmetry…` | `Phases of Theories with Z_N 1-Form Symmetry, and the Roles of Center Vortices and Magnetic Monopoles`, and a `math-linear` warning |
| 10.1007/s00717-026-00622-0 | additional title `X-linked juvenile` + no-break space + `retinoschisis` | `X-linked juvenile retinoschisis` (an ordinary space), type `other`, and a `multiple-titles` warning |
| 10.1103/physrevb.109.035167 | JATS with a `<jats:title>Abstract</jats:title>` heading, `<a:math>` formulas and "Published by the American Physical Society" | `<p>Interlayer coupling is strongly implicated in the complex electronic properties of \(1T-\mathrm{TaS}_{2}\). …</p>`, with no heading and no copyright line |

## Formulas in descriptions

A formula in an abstract is written as TeX between `\(` and `\)`. The page that
shows the description needs MathJax, or another renderer that reads those
delimiters, for the formula to appear as a formula; without one the reader
sees the TeX. Check this on your RDM instance before relying on it.

## Changing the symbol table

The conversion uses a table of symbols built into crossrefapi (entities,
subscripts and superscripts, TeX commands for formula tokens, and the abstract
headings to drop). To change it, name a crossrefapi override file in the
options YAML:

~~~yaml
# options.yaml
symbols: symbols.json
~~~

A relative path is read relative to the options file. The override file holds
only the entries it changes: an entry replaces the built in entry with the same
key, a new key is added, and a `dropped_headings` list replaces the built in
list. The crossrefapi page describes the format. Without the option the built
in table is used.

A missing symbols file stops the run with exit code 66 and an invalid one with
78, before any record is fetched. The default invocation, `doi2rdm options.yaml
DOI`, reports them the same way, so a script can tell a configuration problem
from a DOI that was not found (exit code 1). The manual lists every exit code.

## Warnings

When part of a title or abstract cannot be converted cleanly, the record is
still written with the best text available, and `doi2rdm` writes one line for
each such case to standard error:

~~~
doi2rdm: 10.1103/physrevlett.134.141902: title: math-linear: Z_N
doi2rdm: 10.1007/s00717-026-00622-0: title: multiple-titles: X-linked juvenile retinoschisis
~~~

Control characters in a fragment, such as terminal escape sequences, are replaced by spaces, so a warning is always one line of printable text. The fields are `title`, `additional title` and `description`. The codes are:

| Code | Meaning |
|---|---|
| `parse-fallback` | The markup could not be parsed and was kept as text. |
| `script-unmapped` | A subscript or superscript has no Unicode form and was kept as text. |
| `math-linear` | A formula was written on one line (for example `Z_N`) because it has no simple Unicode form. |
| `math-unconverted` | A formula element could not be converted and was kept as text. |
| `multiple-titles` | CrossRef gave more than one title. The first is used; the fragment is the others. |

Warnings never change the JSON on standard output or the exit status, so a
script that reads the record and the exit code is unaffected. They are there for
a person to look at the odd records.
