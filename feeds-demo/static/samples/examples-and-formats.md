
# Sample, Examples and formats

Each repository provides content retrievable by content type. In CaltechAUTHORS you have content such as articles, books, book sections, conference proceedings. In CaltechTHESIS you have thesis and dissertations and in CaltechDATA you have datasets and software source code.

On this site we aggregate content by Caltech people and groups. We provide the same general level of formats and content feeds for each person and group on the site. This usually includes descriptive information (e.g. a short bio for a person, the description of a group or their active dates). We then link to feeds for each repository system.

Additionally this site provides Caltech wide "recent" list for content deposited in CaltechAUTHORS and CaltechDATA.  All lists are sorted by publication date with most recent publication date first. "recent" lists are capped at 25 entries.  In the people and groups directories you'll find full lists as well.

## Example content

+ [Recent Articles](/recent/article.html) holds recent articles from [CaltechAUTHORS](http://authors.library.caltech.edu)
    + formats available: [JSON](/recent/article.json), HTML [include](/recent/article.include), [BibTeX](/recent/article.bib), [RSS](/recent/article.rss)
+ [Recent Combined](/recent/combined.html) holds recent all types of publications (including journal articles) from [CaltechAUTHORS](http://authors.library.caltech.edu)
    + formats available: [JSON](/recent/combined.json), HTML [include](/recent/combined.include), [BibTeX](/recent/combined.bib), [RSS](/recent/combined.rss)
+ [Groups](/groups/) holds a list of content from [CaltechAUTHORS](http://authors.library.caltech.edu) and [CaltechDATA](https://data.caltech.edu)  based around a curated list of groups. The groups correspond to research groups, come active and some historic.
    + [groups/group_list.json](/groups/group_list.json) holds a list of group names 
+ [People](/people/) listed by a Caltech Library generated person id contains content feeds for [CaltechTHESIS](https://thesis.library.caltech.edu), CaltechAUTHORS, and CaltechDATA
    + [people/people_list.json](/people/people_list.json) holds a list of person ids used to retrieve feed data.

### How to select a different format of feed?

Selecting another format, [https://feeds.library.caltech.edu/recent/article.html](/recent/article.html) becomes ...

+ an HTML Include version [https://feeds.library.caltech.edu/recent/article.include](/recent/article.include)
+ a BibTeX version [https://feeds.library.caltech.edu/recent/article.bib](/recent/article.bib)
+ a JSON version [https://feeds.library.caltech.edu/recent/article.json](/recent/article.json)
+ an RSS 2 version [https://feeds.library.caltech.edu/recent/article.rss](/recent/article.rss)

#### But what format should I work with?

While this site provides six different formats to work with which one depends on your purpose, tools and skills available. For building websites JSON is the most flexible and complete representation of our content. Using HTML include is probably the easiest to use though it is harder to modify (e.g. CSS is straight forward but using regular expressions to modify the content is trickier).  For tracking changes and updates in a news aggregator an RSS feed is convenient. If you're using a Citation manager BibTeX is a good choice. If you want to use a people page as a starting point for a document then Markdown is a likely choice. The examples in this website focus on JavaScript and JSON for the most part. Since the feeds are available as JSON documents you should be able to use the content from most modern programming languages.


## Sample code

Below are links to JavaScript/HTML sample code useful for integrating feeds.library.caltech.edu content in your website.

+ recent journal articles from CaltechAUTHORS
    + [HTML include file for a person](recent-articles-from-include.html) (simplest and not very flexible)
    + [HTML include file for a group](articles-from-include-group.html) (simplest and not very flexible)
    + [from JSON file](recent-articles-from-JSON.html) (requires knowledge of JavaScript and HTML)
    + [three recent articles from group list](three-recent-articles-from-JSON.html) (requires knowledge fo JavaScript and HTML)
    + [excluding articles from a list](excluding-articles-from-a-list.html) (requires knowledge of JavaScript and HTML)

These techniques can be applied across all our feeds including [recent](../recent/), [groups](../groups/) as well as [people](../people/) feeds.
