%doi2rdm-text_normalization(1) irdmtools user manual | version 0.1.0 6ccc3239
% R. S. Doiel and Tom Morrell
% 2026-10-27

# NAME

doi2rdm-text_normalization - CrossRef returns titles and abstracts as publishers wrote them: with markup, character entities,...

# SYNOPSIS

doi2rdm -help text_normalization

# DESCRIPTION

CrossRef returns titles and abstracts as publishers wrote them: with markup,
character entities, MathML formulas, layout whitespace and the subtitle in a
separate field. doi2rdm converts three fields of a CrossRef record before
it writes the JSON, using the text normalization of the crossrefapi package.

title
: the CrossRef title as plain text (tags removed, entities decoded, formulas
  converted to Unicode, layout whitespace tidied) with the subtitle appended
  after a colon (after a space only, if the title already ends in a period,
  question mark, exclamation mark or colon), unless the title already holds
  the subtitle.

additional title
: the other CrossRef titles that add something to the first one, as plain
  text, each with the type "other". A second title that only differs by
  capitalization is not added. The first CrossRef title is the title.

description
: the abstract as HTML that Invenio RDM keeps: paragraphs, headings as bold
  paragraphs, and inline markup such as italics, with formulas written as TeX
  between \( and \). The leading "Abstract" heading and the publisher's
  copyright line are dropped. A record with no abstract has no description.
  The page that shows the description needs MathJax, or another renderer that
  reads the \( \) delimiters, for the formulas to appear as formulas.

DataCite records are not changed by this.

To change the table of symbols the conversion uses, name a crossrefapi
override file in the options YAML (a relative path is read relative to the
options file):

~~~
symbols: symbols.json
~~~

Without it the table built into crossrefapi is used. A missing file exits with
66 and an invalid one with 78, before any record is fetched.

# SEE ALSO

doi2rdm(1)
