
${if(name)}
# ${name} 
${endif}

${if(alternative)}
## (${alternative})
${endif}


${if(description)}

${if(start)}
started ${start}${if(end)} ended ${end}${endif}
${endif}

${description}

${endif}

${if(website)}<${website}>${endif}

${for(content)}
${if(it.repository)}

... from [${it.repository}](${it.href})

- Combined [HTML](${it.combined}.html), [HTML Include](${it.combined}.include), [BibTeX](${it.combined}.bib), [JSON](${it.combined}.json), [RSS](${it.combined}.rss)
${endif}
${if(it.resource_type)}
- ${it.label} [HTML](${it.resource_type}.html), [HTML Include](${it.resource_type}.include), [BibTeX](${it.resource_type}.bib), [JSON](${it.resource_type}.json), [RSS](${it.resource_type}.rss)
${endif}
${endfor}


