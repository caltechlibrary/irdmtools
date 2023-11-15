
# Sample, Examples and formats

## Markdown

[Markdown](https://en.wikipedia.org/wiki/Markdown) is a way of writing plain text documents that can be easily converted to HTML while remaining easy to read for humans. Markdown is plain old text.  Markdown was created by [John Grubber](https://en.wikipedia.org/wiki/John_Gruber) and described on his website [Daring Fireball](https://daringfireball.net/). He based it on practice he observed in email and news group posts that people used to indicate formatting or emphasis. He looked for approaches that could be easily translated to other formats like HTML via regular expression transforms. While HTML was the original format targeted you can also convert Markdown to TeX/LaTeX and PDFs.

Formatting documents using plain text was not new. Caltech alumni [Donald Knuth](https://cs.stanford.edu/~knuth/) created [TeX](https://tug.org/) typesetting system that can typeset everything from a simple memo, to music and complex mathematic equations in text books. Unix also evolved a document formatting system called [troff](https://troff.org) to meet similar purposes. These both used documents written in plain old text.  What really sets Markdown appart from systems like these is Markdown's simplicity. John Grubber's insists on keeping Markdown's feature set small and limited. This made his system easy to read, easy to learn and easy to port. If you need something fancier you can always switch to another presentation (e.g. HTML, TeX, LaTeX, or Troff).  

Markdown's very limited feature set does come at a cost.  For many types of writing you need just a little bit more. Because Markdown was easy to implement this lead to an explosion of dialect notations based on Markdown.  At various points the writing community using these different systems came together to try and unify what they had done while keeping the spirit of the Markdown. One of these efforts evolved in to [CommonMark](https://commonmark.org). Two other Markdown dialects are also common in academic, software and scientific writing -- [Github Flavored Markdown](https://github.github.com/gfm/) and [RMarkdown](https://github.github.com/gfm/ "also called gfm").  With these extensions to Markdown you can easily include code samples or complex math equations (using LaTeX style math notation). A set of tools called [pandoc](https://pandoc.org) also evolved to meet the challenges of coping with many of the variant dialects of Markdown. Additionally platforms like [Jupyter Notebooks](https://jupyter-notebooks.org) and [R-Studio](https://rstudio.com) support working with Markdown content.


### How Caltech Library uses Markdown on feeds.library.caltech.edu?

The prose in this website is written in Markdown as are the programmatically created pages. The Markdown documents are rendered into HTML creating the website.  This HTML page you're reading was written in Markdown. For most webpages on this website you can see the Markdown version by swapping the `.html` extension at the end of the URL for `.md`. Give it a [try](markdown.md "view the markdown version of this html page").

### Where else is Markdown used?

One of the places you'll find Markdown is in [Jupyter Notebooks](https://jupyter.org/), (see Jupyter Notebooks' docs on Markdown at [juptyter-notebook.readthedocs.io](https://jupyter-notebook.readthedocs.io/en/stable/examples/Notebook/Working%20With%20Markdown%20Cells.html)). Jupyter Notebooks are becoming a mainstream way to distribute interactive scientific papers.  In addition to Jupyter Notebooks, [R-Studio](https://rstudio.com) supports [RMarkdown](https://rmarkdown.rstudio.com) for much the same purpose.

### Where can I learn more about Markdown?

Caltech Library hosts [Author Carpentry](https://authorcarpentry.github.io/). Markdown is featured as a topic in many workshops offered by Author Carpentry as well as by [Data Carpentry](https://datacarpentry.org/) and [Software Carpentry](https://software-carpentry.org/). Check the library's [event schedule](https://libcal.caltech.edu/calendar/classes/?cid=3754&t=d&d=0000-00-00&cal=3754) or contact a Caltech Librarian to find out about workshops we offer.


#### More reading material

+ [Markdown and Pandoc](https://authorcarpentry.github.io/markdown-pandoc/) lessons from AuthorCarpentry
+ [Pandoc](https://pandoc.org)
+ [CommonMark](https://commonmark.org)
+ [Jupyter Notebooks](https://jupyter-notebooks.org)
+ [R-Markdown](https://rmarkdown.rstudio.com/) from R-Studio
+ Caltech Library's [MkPage](https://caltechlibrary.github.io/mkpage) and [Datatools](https://caltechlibrary.github.io/datatools) used in building our feeds site
+ [Hugo](https://gohugo.io/) a popular open-source static site generator
