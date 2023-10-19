

# Recent ${resource_type} from ${repository}

${for(content)}
- ${if(it.author_list)}${it.author_list}${endif} ${it.pub_year} [${it.title}](${it.official_url}) ${if(it.issn)} ISSN ${it.issn}${endif} ${if(it.doi)}[${it.doi}](https://doi.org/${it.doi})${endif}
${endfor}



