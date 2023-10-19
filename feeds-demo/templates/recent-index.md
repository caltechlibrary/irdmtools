
# Recent from ${repository}

- Combined ([HTML](${combined}.html), [HTML Include](${combined}.include), [Markdown](${combined}.md), [BibTeX](${combined}.bib), [JSON](${combined}.json), [RSS](${combined}.rss)
${for(content)}
- ${it.label}  ([HTML](${it.name}.html), [HTML Include](${it.name}.include), [Markdown](${it.name}.md), [BibTeX](${it.name}.bib), [JSON](${it.name}.json), [RSS](${it.name}.rss)) 
${endfor}


