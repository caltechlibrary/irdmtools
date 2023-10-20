
# Sample, Examples and formats

## JSON content format

[JSON](https://en.wikipedia.org/wiki/JSON) is our pervasive content type for feeds. JSON is a simple notation for presenting structured data. It is easily used from many popular programming languages include Go, JavaScript, Julia, Perl, Python, PHP, R, and Rust. It was "discovered" by Douglas Crockford when he noticed the similarity of how people described data structures in JavaScript, Java, C and any "curly braced" languages. He had grown frustrated with the wait of XML which the was the common data format used at the time and embraced this simpler way of presenting data using a JavaScript compatible notation. You can read the whole specification at [json.org](https://json.org). It's impressively short for a specification that actually works pretty well.

### How this sites uses JSON

The feeds site for Caltech Library presents the majority of its
content in JSON. This is our unifying layer regardless of the data's origin (e.g. EPrints for CaltechAUTHORS and CaltechDATA, Invenio for CaltechDATA and ArchivesSpace for the archives catalog content).

#### Examples lists and objects

If you want to build a navigator it is handy to have
a list of the available groups and persons. This is
provided as JSON documents.

+ JSON list of [groups](/groups/group_list.json)
+ JSON list of [people](/people/people_list.json)

Getting details on a group or people try

+ JSON object document for [LIGO](/groups/LIGO/group.json) group
+ JSON object document for [Diane K. Newman's](/people/Newman-D-K/people.json) people 
