
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

## ${it.resource_type} from [${it.repository}](${it.href})

- ${if(it.author_list)}${it.author_list}${endif} ${it.pub_year} [${it.title}](${it.official_url}) ${if(it.issn)} ISSN ${it.issn}${endif} ${if(it.doi)}[${it.doi}](https://doi.org/${it.doi})${endif}
${endif}
- ${if(it.author_list)}${it.author_list}${endif} ${it.pub_year} [${it.title}](${it.official_url}) ${if(it.issn)} ISSN ${it.issn}${endif} ${if(it.doi)}[${it.doi}](https://doi.org/${it.doi})${endif}
${endfor}


