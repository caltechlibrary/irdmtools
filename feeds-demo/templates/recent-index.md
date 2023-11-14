
# Recent

${for(content)}
${if(it.repository)}

## ${it.repository}

${endif}
- ${it.label}  ([HTML](${it.name}.html), [HTML Include](${it.name}.include), [BibTeX](${it.name}.bib), [JSON](${it.name}.json), [RSS](${it.name}.rss)) 
${endfor}

