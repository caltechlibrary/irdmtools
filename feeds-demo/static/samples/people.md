
# Samples, examples and formats

## Caltech People

This is a tour of feed content provided in our Caltech People [people](/people) aggregation.  

The Caltech Library currates a list of Caltech People which includes faculty, researchers and alumni. You can find a list of people with people at [https://feeds.library.caltech.edu/people](/people). For each person we may include some biographical information as well as lists of things in our various repositories (e.g. articles, publications, source code and data). The URL used to find a people is keyed by an identifier assigned by Caltech Library.  This is similar and often the same as the identifier which is used in [CaltechAUTHORS](https://authors.library.caltech.edu). For purposes of demonstration go https://feeds.library.caltech.edu/people and search for "Diane K. Newman".  Click the link for her name. Notice her people includes a short biography like you'd find in the [Caltech Directory](https://directory.caltech.edu). It also includes information about publications in [CaltechAUTHORS](https://authors.library.caltech.edu). If material is available from [CaltechDATA](https://data.caltech.edu) or [CaltechTHESIS](https://thesis.library.caltech.edu) that would be present in this page too.

### Feed formats

We the link to feed documents come in a variety of flavors. We produce feeds in HTML (webpage), HTML include (webpage fragment), Markdown, BibTeX, JSON and RSS because they make it easier for some types of content re-use. The easiest to use is probably **HTML include**. You can cut and paste the version into an existing webpage as is. It is however not terribly flexible beyond inserting into an existing web page and styling it with CSS classes and ids.  

**Markdown** is a format that many web content systems can use to turn into HTML and it is easier to ready and modify than raw HTML. Additionally with tools like [RMarkdown](https://rmarkdown.rstudio.com/) and [pandoc](https://pandoc.org/) you can convert Markdown into may other formats include HTML, PDF, or LaTeX. Markdown is what we use to generate the HTML content for this website. Throughout this website if replace the ".html" in the URL with ".md" in most pages you'll see the Markdown version of the content.

**BibTeX** is an old format use with LaTeX to manage bibliographies. It can be used with a huge number of tools as a result. An example of such a tool would be [JabRef](https://www.jabref.org/). BibTeX is also a format supported in many citation manager systems.

**JSON** is the most flexible format for working with structured data in languages like [JavaScript](https://openjs.org), [Python](https://python.org), and [R](https://www.r-project.org/). Our JSON feeds include the richest amount of metadata provided by our various repository systems. If you want fine grain control for data analyzing or rendering your content this is a good format to work with.

**RSS** is a format used by [news aggregators](https://en.wikipedia.org/wiki/News_aggregator) and news reading software. If you use a news reader or aggregator this is the format you want.

## A people tour

In this example we're going to look for a list of journal articles by 
[Professor Diane K. Newman](/people/Newman-D-K). We'll explore a few of the formats available then we'll look at ways to include these feeds in a website or webpage. 

### Finding a people

The people are organized by a Caltech Library assigned id. The people id or person id is similar to the one used in CaltechAUTHORS and CaltechTHESIS.Often it is the same id.  We use it to link people across our repository systems and build people for them on feeds.library.caltech.edu.

#### Using a name to find a person id

The easiest  place is looking in the [people](/people) directory of feeds. There is a link to it "people" at the top of most pages or you can just go straight to [https://feeds.library.caltech.edu/people](/people). The people page shows an A-Z listing of Caltech People. The people listed are a library curated list of Caltech facutly, researchers and alumni.  If you're not in this list or notice someone missing let the library know. Since we are currently curating this list by hand we know many names are missing. If you are on this list and wish to not appear also let us know and we will remove you from our feeds.

On the [people](/people) page look for "Diane K. Newman". You will see her name linked and a long number to the right. Click on Prof. Newman's name. This will take you to her people page.

On Prof. Newman's people page you'll see a short biography similar to that found in the Caltech Directory. You'll also see a section listing links from each of our repository systems - CaltechAUTHORS, CaltechDATA and CaltechTHESIS. Currently we only have material for Prof. Newman in CaltechAUTHORS so that is the only set of feeds listed for her.

In your list of feeds you'll find "combined". This is the combined publication list from CaltechAUTHORS. You'll also see specific publication types feeds (e.g. articles, book sections). The publications listed in these feeds are by descending date (most resent publication first). These feed documents are her complete list from CaltechAUTHORS for each type. We also provide a "recent" version of each of these feed documents. The "recent" lists are capped at 25 entries. For faculty and researchers with a long publication history working with the "recent" feeds provides a faster download because the list is truncated. If you need to have a shorter list then you'll need to look at one of our JavaScript examples. Using JavaScript you can a feed and choose to only show a subset of it.


To start find her  at [/people/Newman-D-K](/people/Newman-D-K). The "Newman-D-K" is what we call a person identifier.  This identifier is created by Caltech Library and is what we use to link between various records in various repositories. Many faculty, researchers have an id called [ORCID](https://orcid.org). You can use a person's ORCID to find their library people. Prof. Newman's  you could use the path [/person/0000-0003-1647-1918](/person/0000-0003-1647-1918) which will redirect your web browser to [/people/Newman-D-K](/people/Newman-D-K).

#### Using an ORCID to find an id

In addition to looking someone up on the main [people](/people) page if you know their ORCID you can also look them up by that. Prof. Newman's ORCID is "0000-0003-1647-1918". If you point your web browser at [https://feeds.library.caltech.edu/person/0000-0003-1647-1918](/person/0000-0003-1647-1918) your browser will redirect to Prof. Newman's people at [/people/Newman-D-K](/people/Newman-D-K).  This redirect should also work for any of the specific feed documents too - e.g. [https://feeds.library.caltech.edu/person/0000-0003-1647-1918/article.html](/people/Newman-D-K/article.html) will take you to her [article.html](/people/Newman-D-K/article.html) page. This was done because in an earlier version of feeds we only included people for people in CaltechAUTHORS who had an associated ORCID.

## Working with lists of articles

### Problem, show only 3 articles

A common request we get is for a feed of most recent "N" articles where "N" is a number less than 25.  In this example we want to show the three (3) most recent "articles" published by Prof. Newman.

#### What we know

[Newman-D-K](/people/Newman-D-K)
: Professor Newman's people id

[recent/article](/people/Newman-D-K/recent/article.json)
: We want the recent article feed

URL to CL.js
: [https://feeds.library.caltech.edu/scripts/CL.js](/scripts/CL.js)

The HTML element id to put results into
: e.g. for `<div id="recent-3-list"></div>`

#### Solution

Caltech Library provides a JavaScript library that makes it easy to work with our feeds, this script is called `CL.js`.  You can copy the JavaScript (or include it) from [https://feeds.library.caltech.edu/scripts/CL.js](/scripts/CL.js).

The `CL.js` script provides a function for getting a JSON version of a
people. The function is `CL.getPeopleJSON()`.  To use the function we need to know Prof Newman's people id (i.e. Newman-D-K), the feed path (i.e. recent/article), the HTML element id we want to put the article listing into (e.g. "recent-3-list" and of course the path to `CL.js`.


In your webpage include the following.

```html
    <div id="recent-3-list">Recent 3 Articles go here if JavaScript worked!</div>

    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* This is our script, in it we'll use the CL object to get
        and limit the number of articles displayed */
    let recent_list = document.getElementById('recent-3-list');
    
    CL.getPeopleJSON("Newman-D-K", "recent/article", function (records, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        /* This is the element we got in the "let" statement above */
        records.forEach(function(rec, i) {
            let elem = document.createElement("div"),
                h1 = document.createElement("h1"),
                anchor = document.createElement("a"),
                div = document.createElement("div");;

            /* we're only going to show three items */
            if (i < 3) {
                /* Add our classes to the our elements */
                elem.classList.add("article");
                h1.classList.add("article-title");
                div.classList.add("article-abstract");
                /* Here we put the content into our elements */
                anchor.setAttribute("href", rec.official_url);
                anchor.innerHTML = rec.title;
                h1.appendChild(anchor);
                div.innerHTML = rec.abstract;
                elem.appendChild(h1);
                elem.appendChild(div);
                /* Finally add our composit elements to the list */
                recent_list.appendChild(elem)
            }
        });
    })
    </script>
```

##### Solution output

**Start of solution output**


<div id="recent-3-list">Recent 3 Articles go here if the JavaScript worked!  </div>


**End of solution output**

<script src="/scripts/CL.js"></script>
<script>CL.BaseURL="";</script>
<script>
/* This is our script, in it we'll use the CL object to get
    and limit the number of articles displayed */
let recent_list = document.getElementById('recent-3-list');

CL.getPeopleJSON("Newman-D-K", "recent/article", function (records, err) {
    if (err != "") {
        console.log("ERROR", err);
        return;
    }
    /* This is the element we got in the "let" statement above */
    records.forEach(function(rec, i) {
        let elem = document.createElement("div"),
            h1 = document.createElement("h1"),
            anchor = document.createElement("a"),
            div = document.createElement("div");;

        /* we're only going to show three items */
        if (i < 3) {
            /* Add our classes to the our elements */
            elem.classList.add("article");
            h1.classList.add("article-title");
            div.classList.add("article-abstract");
            /* Here we put the content into our elements */
            anchor.setAttribute("href", rec.official_url);
            anchor.innerHTML = rec.title;
            h1.appendChild(anchor);
            div.innerHTML = rec.abstract;
            elem.appendChild(h1);
            elem.appendChild(div);
            /* Finally add our composit elements to the list */
            recent_list.appendChild(elem)
        }
    });
})
</script>


### Problem, show articles associated with group for a people

For this problem we'll work with Prof. Throne's people. We want to
list articles for Prof. Throne associated with the [LIGO](/groups/LIGO).
We want to list only the most recent three articles where Prof. Throne is one of the authors AND the article is associated with the LIGO group.

#### What we know

Thorne-K-S
: Prof. Thorne's people id.

LIGO
: The [LIGO group](/groups/LIGO)'s group id

URL to CL.js
: [https://feeds.library.caltech.edu/scripts/CL.js](/scripts/CL.js)

The HTML element id to put results into
: e.g. for `<div id="recent-3-ligo"></div>`

The data path for the group info
: article

The data path for the people info
: article

#### Solution

There is no specific feed for articles by Prof. Thorne from the LIGO group.  We can derive the articles we want by knowing the publication ids published by the LIGO group and the ones published by Prof. Thorne. If we get a list of all publications by LIGO and check if a specific publication by Prof. Thorne is in it we can then list the article in our output.

Implementation notes: we will be to get the whole list of combined ids (keys) from LIGO using `CL.getGroupKeys()` and get the detailed combined list of for Prof. Thorne from `CL.getPeopleJSON()` like in the previous example.


In your webpage include the following.

```html
    <div id="recent-3-ligo">Recent 3 LIGO Articles will go here if the JavaScript worked!</div>

    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* This is our script, in it we'll use the CL object to get
        and limit the number of articles displayed */
    let recent_ligo = document.getElementById('recent-3-ligo');

    /* First get our list of LIGO article feeds. */
    CL.getGroupKeys("LIGO", "article", function (ligo_keys, err) {
        CL.getPeopleJSON("Thorne-K-S", "article", function (records, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            let cnt = 0;
            records.forEach(function(rec) {
                let article_key = rec['_Key'];

                /*
                    Check if we have a LIGO article, increment cnt,
                    check if cnt > 3.
                */
                if (cnt < 3 && ligo_keys.includes(article_key)) {
                    cnt++;
                    let elem = document.createElement("div"),
                        h1 = document.createElement("h1"),
                        anchor = document.createElement("a"),
                        div = document.createElement("div");

                    /* Add our CSS classes for styling */
                    elem.classList.add("article");
                    h1.classList.add("article-title");
                    div.classList.add("article-abstract");
                    /* Here we add the content as a set of elements */
                    anchor.setAttribute("href", rec.official_url);
                    anchor.innerHTML = rec.title
                    h1.appendChild(anchor);
                    div.innerHTML = rec.abstract;
                    elem.appendChild(h1);
                    elem.appendChild(div);
                    /* Finally add our component element to the list */
                    recent_ligo.appendChild(elem);
                }
            });
        });
    });
    </script>
```

##### Solution output

**Start of solution output**


<div id="recent-3-ligo">Recent 3 LIGO Articles will go here if the JavaScript worked!</div>


**End of solution output**

<script>
/* This is our script, in it we'll use the CL object to get
    and limit the number of articles displayed */
let recent_ligo = document.getElementById('recent-3-ligo');

/* First get our list of LIGO article feeds. */
CL.getGroupKeys("LIGO", "article", function (ligo_keys, err) {
    CL.getPeopleJSON("Thorne-K-S", "article", function (records, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        let cnt = 0;
        records.forEach(function(rec) {
            let article_key = rec['_Key'];

            /*
                Check if we have a LIGO article, increment cnt,
                check if cnt > 3.
            */
            if (cnt < 3 && ligo_keys.includes(article_key)) {
                cnt++;
                let elem = document.createElement("div"),
                    h1 = document.createElement("h1"),
                    anchor = document.createElement("a"),
                    div = document.createElement("div");

                /* Add our CSS classes for styling */
                elem.classList.add("article");
                h1.classList.add("article-title");
                div.classList.add("article-abstract");
                /* Here we add the content as a set of elements */
                anchor.setAttribute("href", rec.official_url);
                anchor.innerHTML = rec.title
                h1.appendChild(anchor);
                div.innerHTML = rec.abstract;
                elem.appendChild(h1);
                elem.appendChild(div);
                /* Finally add our component element to the list */
                recent_ligo.appendChild(elem);
            }
        });
    });
});
</script>

### Problem, how to list articles for a people excluding a group

In this example we want to show the most recent three articles from Prof. Thorne excluding any associated with LIGO.  This is solved much like including the articles from a group we only need to change the logic when we check to make sure the key is exclude and we haven't displayed more then the number of articles we require (in this case 3).

#### What we know

Thorne-K-S
: Prof. Thorne's people id.

LIGO
: The [LIGO group](/groups/LIGO)'s group id

URL to CL.js
: [https://feeds.library.caltech.edu/scripts/CL.js](/scripts/CL.js)

The HTML element id to put results into
: e.g. for `<div id="recent-3-not-ligo"></div>`

The data path for the group info
: article

The data path for the people info
: article

We can't use "recent/article" because that will limit the number of items checked to 25 saving if the three articles aren't in the last 25 the list will be empty.


#### Solution

There is no specific feed for articles by Prof. Thorne excluding those associated with the LIGO group.  We can derive the articles we want by knowing the publication ids published by the LIGO group and the ones published by Prof. Thorne. If we get a list of all publications by LIGO and check if a specific publication by Prof. Thorne is in it we can exclude it from our output.

Implementation notes: we will be to get the whole list of article ids (keys) from LIGO using `CL.getGroupKeys()` and get the detailed recent article list of for Prof. Thorne from `CL.getPeopleJSON()` like in the previous example.


In your webpage include the following.

```html
    <div id="recent-3-not-ligo">Recent 3 NOT LIGO Articles go here if JavaScript worked!</div>

    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* This is our script, in it we'll use the CL object to get
        and limit the number of articles displayed */
    let recent_not_ligo = document.getElementById('recent-3-not-ligo');

    /* First get our list of LIGO article feeds. */
    CL.getGroupKeys("LIGO", "article", function (ligo_keys, err) {
        CL.getPeopleJSON("Thorne-K-S", "article", function (records, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            let cnt = 0;
            records.forEach(function(rec) {
                let article_key = rec['_Key'];

                /* 
                    Check if DONT' have a LIGO article, increment cnt,
                    check if cnt > 3.
                */
                if (cnt < 3 && ligo_keys.includes(article_key) == false) {
                    cnt++;
                    let elem = document.createElement("div"),
                        h1 = document.createElement("h1"),
                        anchor = document.createElement("a"),
                        div = document.createElement("div");

                    /* Add your CSS classes for styling */
                    elem.classList.add("article");
                    h1.classList.add("article-title");
                    div.classList.add("article-abstract");

                    /* Now create a composit set of elements to
                       hold the content */
                    anchor.setAttribute("href", rec.official_url);
                    anchor.innerHTML = rec.title;
                    h1.appendChild(anchor);
                    div.innerHTML = rec.abstract;
                    elem.appendChild(h1);
                    elem.appendChild(div);
                   /* Finally add our composit element to list */
                   recent_not_ligo.appendChild(elem);
                }
            });
        });
    });
    </script>
```

##### Solution output

**Start of solution output**


<div id="recent-3-not-ligo">Recent 3 NOT LIGO Articles go here if the JavaScript worked!</div>


**End of solution output**

<script>
/* This is our script, in it we'll use the CL object to get
    and limit the number of articles displayed */
let recent_not_ligo = document.getElementById('recent-3-not-ligo');

/* First get our list of LIGO article feeds. */
CL.getGroupKeys("LIGO", "article", function (ligo_keys, err) {
    CL.getPeopleJSON("Thorne-K-S", "article", function (records, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        let cnt = 0;
        records.forEach(function(rec) {
            let article_key = rec['_Key'];

            /* 
                Check if DONT' have a LIGO article, increment cnt,
                check if cnt > 3.
            */
            if (cnt < 3 && ligo_keys.includes(article_key) == false) {
                cnt++;
                let elem = document.createElement("div"),
                    h1 = document.createElement("h1"),
                    anchor = document.createElement("a"),
                    div = document.createElement("div");

                /* Add your CSS classes for styling */
                elem.classList.add("article");
                h1.classList.add("article-title");
                div.classList.add("article-abstract");

                /* Now create a composit set of elements to
                   hold the content */
                anchor.setAttribute("href", rec.official_url);
                anchor.innerHTML = rec.title;
                h1.appendChild(anchor);
                div.innerHTML = rec.abstract;
                elem.appendChild(h1);
                elem.appendChild(div);
               /* Finally add our composit element to list */
               recent_not_ligo.appendChild(elem);
            }
        });
    });
});
</script>

### Problem, how to list recent 3 years of  article titles?

In this example we want to show of the last 3 years of articles from 
Prof. Thorne grouped by year. This is a variation of showing recent
three articles but we want to cut off display when we hit the fourth
year rather than when we have hit a specific number of articles.

#### What we know

Thorne-K-S
: Prof. Thorne's people id.

URL to CL.js
: [https://feeds.library.caltech.edu/scripts/CL.js](/scripts/CL.js)

The HTML element id to put results into
: e.g. for `<div id="last-3-years"></div>`

The data path for the people info
: recent/article

We use "recent/article" because that will limit the number of items to 25.


#### Solution

There is a specific feed for recent 25, we use that but we need to add a year heading when the publication date changes years and count the changes in year to end our list at 3.

In your webpage include the following.

```html
    <div id="last-3-years">
    Recent article titles by year go here if JavaScript worked!
    </div>

    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* This is our script, in it we'll use the CL object to get
        and limit the number of articles displayed */
    let last_3_years = document.getElementById('last-3-years')

    /* First get our list of LIGO article feeds. */
    CL.getPeopleJSON("Thorne-K-S", "recent/article", function (articles, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        let year_label = '',
            year_count = 0;

        articles.forEach(function(article) {
            let elem = document.createElement("div"),
                h3 = document.createElement("h3"),
                anchor = document.createElement("a");

            if (article.date_type == "published" && article.date.substring(0,4) != year_label) {
                let h2 = document.createElement("h2");
                year_label = article.date.substring(0,4);
                year_count++;
                /* If we have more than 3 years we can end our listing */
                if (year_count > 3) {
                    return;
                }
                h2.innerHTML = year_label;
                last_3_years.appendChild(h2);
            }
            /* Add your CSS classes for styling */
            elem.classList.add("article");
            h3.classList.add("article-title");

            /* Now create a composit set of elements to
               hold the content */
            anchor.setAttribute("href", article.official_url);
            anchor.innerHTML = article.title;
            h3.appendChild(anchor);
            elem.appendChild(h3);
           /* Finally add our composit element to list */
           last_3_years.appendChild(elem);
        });
    });
    </script>
```

##### Solution output

**Start of solution output**


<div id="last-3-years">Recent articles titles by year go here if the JavaScript worked!</div>


**End of solution output**

<script>
/* This is our script, in it we'll use the CL object to get
    and limit the number of articles displayed */
let last_3_years = document.getElementById('last-3-years')

/* First get our list of LIGO article feeds. */
CL.getPeopleJSON("Thorne-K-S", "recent/article", function (articles, err) {
    if (err != "") {
        console.log("ERROR", err);
        return;
    }
    let year_label = '',
        year_count = 0;

    articles.forEach(function(article) {
        let elem = document.createElement("div"),
            h3 = document.createElement("h3"),
            anchor = document.createElement("a");

        if (article.date_type == "published" && article.date.substring(0,4) != year_label) {
            let h2 = document.createElement("h2");
            year_label = article.date.substring(0,4);
            year_count++;
            /* If we have more than 3 years we can end our listing */
            if (year_count > 3) {
                return;
            }
            h2.innerHTML = year_label;
            last_3_years.appendChild(h2);
        }
        /* Add your CSS classes for styling */
        elem.classList.add("article");
        h3.classList.add("article-title");

        /* Now create a composit set of elements to
           hold the content */
        anchor.setAttribute("href", article.official_url);
        anchor.innerHTML = article.title;
        h3.appendChild(anchor);
        elem.appendChild(h3);
       /* Finally add our composit element to list */
       last_3_years.appendChild(elem);
    });
});
</script>
