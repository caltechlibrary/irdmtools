
${if(sort_name)}
# ${sort_name}
${endif}

${if(links_and_identifiers)}
## Links and Identifiers

${for(links_and_identifiers)}
- ${it.description} [${label}](${link})
${endfor}
${endif}

${if(title)}
## Title

${title}
${endif}

${if(division)}

## Division

${division}
${endif}

${if(bio)}
## Biography

${bio}
${endif}

${if(advisor)}

## Advisor

${for(advisor)}
- Thesis and Dissertations: [HTML](${it.combined}.html), [HTML Include](${it.combined}.include), [Markdown](${it.combined}.md), [BibTeX](${it.combined}.bib), [JSON](${it.combined}.json), [RSS](${it.combined}.rss)
${endfor}
${endif}

${for(content)}
${if(it.repository)}

... from [${it.repository}](${it.href})

- Combined [HTML](${it.combined}.html), [HTML Include](${it.combined}.include), [Markdown](${it.combined}.md), [BibTeX](${it.combined}.bib), [JSON](${it.combined}.json), [RSS](${it.combined}.rss)
${endif}
${if(it.resource_type)}
- ${it.label} [HTML](${it.resource_type}.html), [HTML Include](${it.resource_type}.include), [Markdown](${it.resource_type}.md), [BibTeX](${it.resource_type}.bib), [JSON](${it.resource_type}.json), [RSS](${it.resource_type}.rss)
${endif}
${endfor}


