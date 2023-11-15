
${if(sort_name)}

# Editor: [${sort_name}](./)

from CaltechAUTHORS

${endif}

${for(content)}
${if(it.repository)}

## ${if(it.resource_label)}${it.resource_label}s${endif} from [${it.repository}](${it.href})

${endif}
${it:citation.md()}
${endfor}


