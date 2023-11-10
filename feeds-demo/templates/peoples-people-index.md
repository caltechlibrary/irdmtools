
${if(sort_name)}

# ${sort_name}

${endif}

${if(links_and_identifiers)}
## Links and Identifiers

${for(links_and_identifiers)}
${if(it.link)}- ${it.description} [${it.label}](${it.link})${endif}
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

${if(caltech_degrees)}

## CaltechTHESIS

${for(caltech_degrees)}
${it:degrees.md()}
${endfor}
${endif}

${if(thesis_advisor)}

## Advisor

- Thesis and Dissertations: [HTML](advisor.html), [HTML Include](advisor.include), [Markdown](advisor.md), [BibTeX](advisor.bib), [JSON](advisor.json), [RSS](advisor.rss)
${endif}

${if(thesis_committee)}

## Committee Member

- Thesis and Dissertations: [HTML](committee.html), [HTML Include](committee.include), [Markdown](committee.md), [BibTeX](committee.bib), [JSON](committee.json), [RSS](committee.rss)
${endif}

${if(editor)}

## Editor

- Editor: [HTML](editor.html), [HTML Include](editor.include), [Markdown](editor.md), [BibTeX](editor.bib), [JSON](editor.json), [RSS](editor.rss)
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


