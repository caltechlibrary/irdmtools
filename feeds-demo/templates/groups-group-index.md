
${if(name)}
# ${name} 
${endif}

${if(alternative)}
## (${alternative})
${endif}


${if(start)}
start ${start}${if(end)} through ${end}${endif}
${endif}

${if(description)}
${description}

${endif}

${if(website)}<${website}>${endif}

${for(content)}
${if(it.repository)}

... from [${it.repository}](${it.href})

- Combined [HTML](${it.combined}.html), [HTML Include](${it.combined}.include), [Markdown](${it.combined}.md), [BibTeX](${it.combined}.bib), [JSON](${it.combined}.json), [RSS](${it.combined}.rss)
${endif}
${if(it.resource_type)}
- ${it.label} [HTML](${it.resource_type}.html), [HTML Include](${it.resource_type}.include), [Markdown](${it.resource_type}.md), [BibTeX](${it.resource_type}.bib), [JSON](${it.resource_type}.json), [RSS](${it.resource_type}.rss)
${endif}
${endfor}

