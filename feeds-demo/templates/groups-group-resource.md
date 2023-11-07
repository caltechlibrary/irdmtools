
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

${for(content)}
${if(it.repository)}

## ${if(it.resource_type)}${it.resource_type}s${endif} from [${it.repository}](${it.href})

${endif}
${it:citation.md()}
${endfor}


