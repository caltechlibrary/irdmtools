
${if(sort_name)}

# Committeee Member: [${sort_name}](./)

Thesis and Dissertations from CaltechTHESIS 

${endif}

${for(content)}
${if(it.repository)}

## ${if(it.resource_label)}${it.resource_label}s${endif} from [${it.repository}](${it.href})

${endif}
${it:citation.md()}
${endfor}


