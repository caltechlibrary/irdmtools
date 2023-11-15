
# Sample, Examples and formats

## Three LIGO articles

This sample demonstrates getting the three most recent
LIGO group articles from feeds.

### What is required

1. Find the string that refers to the group you want to include (e.g. LIGO)
2. Find the feed type you want (e.g. article)
3. Create a JavaScript block and modify as below.
4. In the JavaScript function processing the article keep a count and stop when you have displayed three.

### Example for LIGO

Below is a HTML fragment that you could include in a web page to 
include recent articles titles for the group called 
[LIGO](https://feeds.library.caltech.edu/groups/LIGO)
using the JSON version of the article feed.

```HTML
    <div id="recent-3-articles">
    Recent 3 Articles go here if JavaScript worked!
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js"></script>
    <script>
    let recent_3_articles = document.getElementById("recent-3-articles");
    
    CL.getGroupJSON("LIGO", "recent/article", function(articles, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        let max_count = 3;
        let i = 0;
        articles.forEach(function (article) {
            i++;
            if (i > max_count) {
                return;
            }
            var elem = document.createElement("div"),
                h1 = document.createElement("h1"),
                anchor = document.createElement("anchor"),
                div = document.createElement("div");

            /* Add CSS classes for styling */
            elem.classList.add("article");
            h1.classList.add("article-title");
            div.classList.add("article-abstract");
            /* Add content to create a composit element */
            anchor.setAttribute("href", article.official_url);
            anchor.innerHTML = article.title;
            h1.appendChild(anchor);
            div.innerHTML = article.abstract;
            elem.appendChild(h1);
            elem.appendChild(div);
            /* Finally add our composit element to the list */
            recent_3_articles.appendChild(elem)
        });
    });
    </script>
```

That example is rendered below--

<div id="recent-3-articles">
   Recent 3 Articles go here if JavaScript Worked.
</div>

<script src="/scripts/CL-core.js"></script>
<script>CL.BaseURL = "";</script>
<script>
let recent_3_articles = document.getElementById("recent-3-articles");

CL.getGroupJSON("LIGO", "recent/article", function(articles, err) {
    if (err != "") {
        console.log("ERROR", err);
        return;
    }
    let max_count = 3;
    let i = 0;
    articles.forEach(function (article) {
        i++;
        if (i > max_count) {
            return;
        }
        var elem = document.createElement("div"),
            h1 = document.createElement("h1"),
            anchor = document.createElement("anchor"),
            div = document.createElement("div");

        /* Add CSS classes for styling */
        elem.classList.add("article");
        h1.classList.add("article-title");
        div.classList.add("article-abstract");
        /* Add content to create a composit element */
        anchor.setAttribute("href", article.official_url);
        anchor.innerHTML = article.title;
        h1.appendChild(anchor);
        div.innerHTML = article.abstract;
        elem.appendChild(h1);
        elem.appendChild(div);
        /* Finally add our composit element to the list */
        recent_3_articles.appendChild(elem)
    });
});
</script>

