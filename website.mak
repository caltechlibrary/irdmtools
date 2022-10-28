#
# Makefile for running pandoc on all Markdown docs ending in .md
#
PROJECT = irdmtools

PANDOC = $(shell which pandoc)

MD_PAGES = $(shell ls -1 *.md | grep -v 'nav.md')

HTML_PAGES = $(shell ls -1 *.md | grep -v 'nav.md' | sed -E 's/.md/.html/g')

build: $(HTML_PAGES) $(MD_PAGES)

$(HTML_PAGES): $(MD_PAGES) .FORCE
	if [ -f $(PANDOC) ]; then $(PANDOC) --metadata title=$(basename $@) -s --to html5 $(basename $@).md -o $(basename $@).html \
	    --template=page.tmpl; fi
	@if [ $@ = "README.html" ]; then mv README.html index.html; fi

clean:
	@if [ -f index.html ]; then rm *.html; fi

.FORCE:
