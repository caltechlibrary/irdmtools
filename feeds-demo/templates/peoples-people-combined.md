
${if(sort_name)}
# [${sort_name}](./)
${endif}

${for(content)}
${if(it.repository)}

## ${if(it.resource_type)}${it.resource_type}s${endif} from [${it.repository}](${it.href})

${endif}
${it:citation.md()}
${endfor}


