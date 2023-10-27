
${if(repository)}

## Combined from [${repository}](${href})

${endif}
${for(content)}
${it:citation.md()}
${endfor}


