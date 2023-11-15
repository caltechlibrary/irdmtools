
# Samples, Examples and Formats

## Caltech Groups

This site groups aggregation provides feeds for a Library currated list of Caltech division and research [groups](/groups/) with content in one of our repository systems.  Each group page may include group description, active date(s) as well as of list of [thesis and disertations](https://thesis.library.caltech.edu) associated with the group and content found in  [CaltechAUTHORS](https://authors.library.caltech.edu) and [CaltechDATA](https://data.caltech.edu).  If your group is not listed and you think it should be please contact a Caltech Librarian.

## Using our groups information

Many Caltech Groups have website and if you'd like to include bibliographical and publication related material that the Library curates you can do so easily with the feeds provided on this website. You can also combine the group feeds with people of group members that maybe hosted in this site too.

## Content types available

### CaltechAUTHORS

combined 
: includes all content from CaltechAUTHORS

individual types
: includes article, book, book section, conference item, lab notes, monograph, patent, teaching resource

### CaltechDATA

combined
: includes all content from CaltechDATA

individual types
: audio visual, datasets, image, interactive resources, software, models and text

### CaltechTHESIS

Bachelor, Master, and Engineer's Thesis as well as Ph.D dissertations

### Feed formats

We the link to feed documents come in a variety of flavors. We produce feeds in HTML (webpage), HTML include (webpage fragment), Markdown, BibTeX, JSON and RSS because they make it easier for some types of content re-use. The easiest to use is probably **HTML include**. You can cut and paste the version into an existing webpage as is. It is however not terribly flexible beyond inserting into an existing web page and styling it with CSS classes and ids.

**Markdown** is a format that many web content systems can use to turn into HTML and it is easier to ready and modify than raw HTML. Additionally with tools like [RMarkdown](https://rmarkdown.rstudio.com/) and [pandoc](https://pandoc.org/) you can convert Markdown into may other formats include HTML, PDF, or LaTeX. We use Markdown to generate the HTML content for this website. Throughout this website if replace the ".html" in the URL with ".md" in most pages you'll see the Markdown version of the content.

**BibTeX** is an old format use with LaTeX to manage bibliographies. It can be used with a huge number of tools as a result. An example of such a tool would be [JabRef](https://www.jabref.org/). BibTeX is also a format supported in many citation manager systems.

**JSON** is the most flexible format for working with structured data in languages like [JavaScript](https://openjs.org), [Python](https://python.org), and [R](https://www.r-project.org/). Our JSON feeds include the richest amount of metadata provided by our various repository systems. If you want fine grain control for data analyzing or rendering your content this is a good format to work with.

**RSS** is a format used by [news aggregators](https://en.wikipedia.org/wiki/News_aggregator) and news reading software. If you use a news reader or aggregator this is the format you want.

The most useful formats are JSON or HTML Include. HTML include being the simplest and least flexible format to work with. HTML Include is suitable for cutting and pasting and styling with CSS.

## A group tour

The feeds listed on the main group page include all content by types and the feeds under the recent directory cap the listing at the 25 most recent.

### Problem, How do we get a Group's summary?

Our curated list of Caltech Groups often include descriptive data and dates active. In this example we want to get the descriptive data for TCCON group (formally called "Total Carbon Column Observing Network").

#### What we know

First we need to find out where the TCONN group's data is listed on the website.  We'll need to find out the group id assigned by Caltech Library.  We will also need to include some JavaScript to fetch the data we want and to format the results.


Getting the group id
: Point your web browser at [/groups](/groups/) on this website.  Go to the You'll see an A to Z listing the of the groups. The look towards the bottom and you'll see "[Total Carbon Column Observing Network](groups/Total-Carbon-Column-Observing-Network/)", click that link. Look at the URL. You'll see the URL ends with "Total-Carbon-Column-Observing-Network", this is the Caltech Library assigned group id. 

CL.js
: CL.js is a JavaScript library that makes it easy to get information about a group and its feed. It can be found under [/scripts/CL.js](/scripts/CL.js) on this website.

Get the links to the description
: CL.js provides a function called `CL.getGroupSummary()` that can retrieve the summary information about a group.


#### Solution

Using the group id, the `CL.js` library and an element in an HTML page we can pull the summary content our page to display it.

In your webpage include the following.

```html
    <div id="group-summary">
    The Group Summary should show here if the JavaScript works.
    </div>

    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* We get a handle on our page element */
    let group_summary = document.getElementById('group-summary');

    /* Now we call getGroupSummary() to fill our element. */
    CL.getGroupSummary("Total-Carbon-Column-Observing-Network", 
        function(summary, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            /* Now we need to create some HTML 
               elements to display our content with. */
            var h1 = document.createElement("h1"),
                h2 = document.createElement("h2"),
                p = document.createElement("p");
            /* Create our H1 heading */
            h1.innerHTML = summary.name;
            /* List our alternate names in H2 */
            h2.innerHTML = "(Also known as: " + 
                summary.alternative.join(", ") + ")";
            /* Now get our description */
            p.innerHTML = summary.description;

            /* Now we can add these into our group_summary element. */
            group_summary.appendChild(h1);
            group_summary.appendChild(h2);
            group_summary.appendChild(p);
    });
    </script>
```

##### Solution output

**Start of solution output**

<div id="group-summary">
The Group Summary should show here if the JavaScript works.
</div>

<script src="/scripts/CL-core.js"></script>
<script>CL.BaseURL = "";</script>
<script>
/* We get a handle on our page element */
let group_summary = document.getElementById('group-summary');

/* Now we call getGroupSummary() to fill our element. */
CL.getGroupSummary("Total-Carbon-Column-Observing-Network", 
    function(summary, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        /* Now we need to create some HTML 
           elements to display our content with. */
        var h1 = document.createElement("h1"),
            h2 = document.createElement("h2"),
            p = document.createElement("p");
        /* Create our H1 heading */
        h1.innerHTML = summary.name;
        /* List our alternate names in H2 */
        h2.innerHTML = "(Also known as: " + 
            summary.alternative.join(", ") + ")";
        /* Now get our description */
        p.innerHTML = summary.description;

        /* Now we can add these into our 
           group_summary element. */
        group_summary.appendChild(h1);
        group_summary.appendChild(h2);
        group_summary.appendChild(p);
});
</script>

**End of solution output**

### Problem, How to get a list of the three of a Groups publications?

In the following example we're going to get the three most recent articles published by the
"Big Bear Solar Observatory". 

#### What we know

First we need to find out where the group's data is listed on the website.  We'll need to find out the group id assigned by Caltech Library.  We will also need to include some JavaScript to fetch the data we want and to format the results.

Getting the group id
: Point your web browser at [/groups](/groups/) on this website.  Go to the You'll see an A to Z listing the of the groups. The look towards the bottom and you'll see "[Big Bear Solar Observatory](/groups/Big-Bear-Solar-Observatory/)", click that link. Look at the URL. You'll see the URL ends with "Big-Bear-Solar-Observatory", this is the Caltech Library assigned group id. 

CL.js
: CL.js is a JavaScript library that makes it easy to get information about a group and its feed. It can be found under [/scripts/CL.js](/scripts/CL.js) on this website.

recent/article
: On the recent page you'll see a link to the last 25 articles. The path can be described as "recent/article", we'll use this knowledge to get the feed data in JSON format

Cl.getGroupJSON()
: CL.js provides a function called `CL.getGroupJSON()` for retrieving feed content, we can use that to display the three most recent items.

#### Solution

Using the group id, CL.js, CL.getGroupJSON() and an HTML element with a known id we can retrieve
the article list and insert it into the webpage.

In your webpage include the following.

```html
    <div id="recent-3-articles">
    Recent three articles should go here if JavaScript works.
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js"></script>
    <script>
    /* Get a handle to where we will insert our content on the page */
    let recent_3_articles = document.getElementById('recent-3-articles');

    /* Now we call getGroupJSON() to fill in our element. */
    CL.getGroupJSON("Big-Bear-Solar-Observatory", "recent/article", 
        function(articles, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            /* Now we can add these into our */
            articles.forEach(function(article, i) {
                let elem = document.createElement("div"),
                    h1 = document.createElement("h1"),
                    anchor = document.createElement("a"),
                    div = document.createElement("div");

                /* we're only going to show three items */
                if (i < 3) {
                    elem.classList.add("article");
                    h1.classList.add("article-title");
                    div.classList.add("article-abstract");
                    /* Here we format the content as HTML and build
                       the element we're going to insert */
                    anchor.setAttribute("href", article.official_url);
                    anchor.innerHTML = article.title;
                    div.innerHTML = article.abstract;
                    h1.appendChild(anchor);
                    elem.appendChild(h1);
                    elem.appendChild(div);
                    recent_3_articles.appendChild(elem)
                }
            });
    });
    </script>
```

##### Solution output

**Start of solution output**

<div id="recent-3-articles">
Recent three articles should show here if JavaScript works.
</div>

<script>
/* Get a handle to where we will insert our content on the page */
let recent_3_articles = document.getElementById('recent-3-articles');

/* Now we call getGroupJSON() to fill in our element. */
CL.getGroupJSON("Big-Bear-Solar-Observatory", "recent/article", 
    function(articles, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        /* Now we can add these into our */
        articles.forEach(function(article, i) {
            let elem = document.createElement("div"),
                h1 = document.createElement("h1"),
                anchor = document.createElement("a"),
                div = document.createElement("div");

            /* we're only going to show three items */
            if (i < 3) {
                elem.classList.add("article");
                h1.classList.add("article-title");
                div.classList.add("article-abstract");
                /* Here we format the content as HTML and build
                   the element we're going to insert */
                anchor.setAttribute("href", article.official_url);
                anchor.innerHTML = article.title;
                div.innerHTML = article.abstract;
                h1.appendChild(anchor);
                elem.appendChild(h1);
                elem.appendChild(div);
                recent_3_articles.appendChild(elem)
            }
        });
});
</script>

**End of solution output**


### Problem, How to list of a Group's articles titles by year?

We're going to use the "Big Bear Solar Observatory" group for this
example.  The group pages provides links to recent 25 and the full publication lists.  If we want to show articles by year we should use the full article list.

#### What we know

Group id
: From the previous example we know the "[Big Bear Solar Observatory](/groups/Big-Bear-Solar-Observatory/)" group id is "Big-Bear-Solar-Observatory"

CL.js
: CL.js is a JavaScript library that makes it easy to get information about a group and its feed. It can be found under [/scripts/CL.js](/scripts/CL.js) on this website.

article
: We want to full article list so our feed path will only be "article", we'll use this knowledge to get the feed data in JSON format

Cl.getGroupJSON()
: CL.js provides a function called `CL.getGroupJSON()` for retrieving feed content, we can use that to display the three most recent items.

#### Solution

We will create an HTML element in our web page with the id of "articles-by-year". We use our group id, feed path and the same `CL.js` function `CL.getGroupJSON()` to fetch our articles but as we loop through them we can check to see if the year value in the publication date has changed, if so we'll insert a heading for the new year.

In your webpage include the following.

```html
    <div id="articles-by-year">
    Articles by year should show up here if JavaScript works.
    </div>

    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* Get a handle to where we will insert our content on the page */
    let articles_by_year = document.getElementById('articles-by-year');

    /* Now we call getGroupJSON() to fill in our element. */
    CL.getGroupJSON("Big-Bear-Solar-Observatory", "article", 
        function(articles, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            /* Now we can add these into our */
            let year_label = "";
            articles.forEach(function(article) {
                let elem = document.createElement("div"),
                    h3 = document.createElement("h3"),
                    anchor = document.createElement("a");

                /* What is our publication date? 
                   Has our year label changed? */
                if (article.date_type == "published" &&
                    article.date.substring(0,4) != year_label)  {
                    /* Update our year label and append an h2 in our list */
                    year_label = article.date.substring(0, 4);
                    let h2 = document.createElement("h2");
                    h2.innerHTML = year_label;
                    articles_by_year.appendChild(h2);
                }


                /* Now we just add our article */
                elem.classList.add("article");
                h3.classList.add("article-title");
                /* Here we format the content as HTML and build
                   the element we're going to insert */
                anchor.setAttribute("href", article.official_url);
                anchor.innerHTML = article.title;
                h3.appendChild(anchor);
                elem.appendChild(h3);
                articles_by_year.appendChild(elem)
            });
    });
    </script>
```

##### Solution output

**Start of solution output**

<div id="articles-by-year">
Articles by year should show up here if JavaScript works.
</div>

<script>
/* Get a handle to where we will insert our content on the page */
let articles_by_year = document.getElementById('articles-by-year');

/* Now we call getGroupJSON() to fill in our element. */
CL.getGroupJSON("Big-Bear-Solar-Observatory", "article", 
    function(articles, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        /* Now we can add these into our */
        let year_label = "";
        articles.forEach(function(article) {
            let elem = document.createElement("div"),
                h3 = document.createElement("h3"),
                anchor = document.createElement("a");

            /* What is our publication date? 
               Has our year label changed? */
            if (article.date_type == "published" &&
                article.date.substring(0,4) != year_label)  {
                /* Update our year label and append an h2 in our list */
                year_label = article.date.substring(0, 4);
                let h2 = document.createElement("h2");
                h2.innerHTML = year_label;
                articles_by_year.appendChild(h2);
            }

            /* Now we just add our article */
            elem.classList.add("article");
            h3.classList.add("article-title");
            /* Here we format the content as HTML and build
               the element we're going to insert */
            anchor.setAttribute("href", article.official_url);
            anchor.innerHTML = article.title;
            h3.appendChild(anchor);
            elem.appendChild(h3);
            articles_by_year.appendChild(elem)
        });
});
</script>

**End of solution output**




### Problem, How to get a list of Ph.D dissertations associated with the group?

For this problem we want to know the Ph.D dissertations associated with the "Caltech Antenna Laboratory". This is very similar to getting a list of
all articles for the group except the feed we want comes from CaltechTHESIS data and is just the Ph.D.

#### What we know

Group id
: Going to the [groups](/groups) page we see "Caltech Antenna Laboratory" has a group id of "Caltech-Antenna-Laboratory".

feed path?
: Our feed path is a link on the group page, look for the link for PH.D. the path part we're interested in is "phd".

CL.js
: The CL.js provides the `CL.getGroupJSON()` function for fetching and displaying feed data in JSON form. 

#### Solution

Using our group id (Caltech-Antenna-Laboratory), our path (phd) and our JavaScript function to display our results.

In your webpage include the following.

```html
    <div id="list-of-phd">
    List of PhD goes here if JavaScript works.
    </div>

    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* Get a handle to where we will insert our content on the page */
    let list_of_phd = document.getElementById('list-of-phd');

    /* Now we call getGroupJSON() to fill in our element. */
    CL.getGroupJSON("Caltech-Antenna-Laboratory", "phd", 
        function(phds, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            /* Now we can add these into our */
            let year_label = "";
            phds.forEach(function(phd) {
                let elem = document.createElement("div"),
                    h3 = document.createElement("h3"),
                    anchor = document.createElement("a");

                /* Now we just add our phd, php-title CSS styles */
                elem.classList.add("phd");
                h3.classList.add("php-title");

                /* Here we format the content as HTML and build
                   the element we're going to insert */
                anchor.setAttribute("href", phd.official_url);
                anchor.inerrHTML = phd.title;
                h3.appendChild(anchor);
                elem.appendChild(h3);
                /* Finally add our element to the list_of_phd */
                list_of_phd.appendChild(elem)
            });
    });
    </script>
```

##### Solution output

**Start of solution output**

<div id="list-of-phd">
List of PhD goes here if JavaScript works.
</div>

<script>
/* Get a handle to where we will insert our content on the page */
let list_of_phd = document.getElementById('list-of-phd');

/* Now we call getGroupJSON() to fill in our element. */
CL.getGroupJSON("Caltech-Antenna-Laboratory", "phd", 
    function(phds, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        /* Now we can add these into our list*/
        phds.forEach(function(phd) {
            let elem = document.createElement("div"),
                h3 = document.createElement("h3"),
                anchor = document.createElement("a");

            /* Now we just add our phd, php-title CSS styles */
            elem.classList.add("phd");
            h3.classList.add("php-title");

            /* Here we format the content as HTML and build
               the element we're going to insert */
            anchor.setAttribute("href", phd.official_url);
            anchor.innerHTML = phd.title;
            h3.appendChild(anchor);
            elem.appendChild(h3);
            /* Finally add our element to the list_of_phd */
            list_of_phd.appendChild(elem)
        });
});
</script>

**End of solution output**



