~~~shell
it: ${it}
~~~

Author List
: "${if(it.author_list)} ${it.author_list}${endif}"

Pub Year
: "${if(it.pub_year)} (${it.pub_year})${endif}"

Title
: "${if(it.title)} [${it.title}](${if(it.official_url)}${it.official_url}${else}${it.id}${endif})${endif}"

ISSN
: "${if(it.issn)} ISSN: ${it.issn}${endif}"

ISBN
: "${if(it.isbn)} ISBN: ${it.isbn}${endif}"

DOI
: "${if(it.doi)} DOI: [${it.doi}](https://doi.org/${it.doi})${endif}"

Official URL
: "${if(it.official_url)}${it.official_url}${endif}"

HRef
: <${if(it.href)}${it.href}${endif}>

Record ID
: "${if(it.rdmid)}${it.rdmid}${endif}"

Publisher
: "${if(it.publisher)}${it.publisher}${endif}"

Publication
: "${if(it.publication)}${it.publication}${endif}"


