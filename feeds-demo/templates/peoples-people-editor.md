
${if(sort_name)}

# [${sort_name}](./)

As editor.

${endif}

${for(content)}
${if(it.repository)}

## ${if(it.resource_label)}${it.resource_label}s${endif} from [${it.repository}](${it.href})

${endif}
${it:citation.md()}
${endfor}


