
${if(sort_name)}

# [${sort_name}](./)

${endif}

${if(alternative)}
## (${alternative})
${endif}

${for(content)}
${if(it.repository)}

### ${if(it.resource_label)}${it.resource_label}${endif} from [${it.repository}](${it.href})

${endif}
${it:citation.md()}
${endfor}


