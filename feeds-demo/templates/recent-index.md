
# Recent

${for(content)}
${if(it.repository)}

## from ${it.repository}

${endif}
- ${it.label}  ([HTML](${it.name}.html), [HTML Include](${it.name}.include), [Markdown](${it.name}.md), [BibTeX](${it.name}.bib), [JSON](${it.name}.json), [RSS](${it.name}.rss)) 
${endfor}

