
Need to de-dup subjects when pulled from datacite dois data.

Duplicate listing for records are showing up for,

http://feeds-test.library.caltech.edu.s3-website-us-west-2.amazonaws.com/people/Orphan-V-J/combined.html

Reharvest, confirm that duplicated are solved and then switch production after notifying Donna that fix is happening.

Funding:

.fundingReferences -> .fundingReferences

Example: https://api.datacite.org/dois/10.22002/36sg9-yhj98 (might not be great example),

License

.metadata.rights -> .metadata.rights

Subject:

.subject.subject[*] -> .subject.subject[*]

Example: https://api.datacite.org/dois/10.48550/arXiv.2404.01326

Ignore:

.PublisherLocation, et el.

Date problem:

.publication_date, .publication_year

Cold Spring Harbor DOI
".accepted_date as last fallback"

Datacite added a relatedItems object. The following are the mappings suggested by Tom in issue #77

~~~
If relatedItems is present

If relatedItemType == Journal

Then

relatedItem.titles.title -> custom_fields.journal:journal.title
relatedItem.volume -> custom_fields.journal:journal.volume
relatedItem.issue -> custom_fields.journal:journal.issue
relatedItem.firstPage and relatedItem.lastPage or relatedItem.number -> custom_fields.journal:journal.pages
relatedItem.publisher -> metadata.publisher
if relatedIten.relatedItemIdentifier.relatedItemIdentifierType == ISSN, then relatedItem.relatedItemIdentifier.relatedItemIdentifier -> custom_fields.journal:journal.issn
~~~
