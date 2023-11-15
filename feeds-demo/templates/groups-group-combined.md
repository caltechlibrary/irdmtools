
${if(name)}
# [${name}](./)
${endif}

${if(alternative)}
## (${alternative})
${endif}


${if(start)}
started ${start}${if(end)} ended ${end}${endif}
${endif}

${if(description)}
${description}

${endif}

${if(website)}<${website}>${endif}

${if(repository)}

### combined from [${repository}](${href})

${endif}
${for(content)}
${it:citation.md()}
${endfor}


