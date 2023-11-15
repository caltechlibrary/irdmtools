
# Sample, Examples and formats

## List articles and excluding a group from the list

Sometimes you need a list of a person's publications excluding
those from a given group's publication. This can be done in 
JavaScript easily taking advantage of both the groups' published
feeds' keys and the person's own people feed. In our example
we will be using a feed for Goddard-W-A-III and the group we
will exclude will be LIGO.


### What you need to know to get started

1. Find the person id from the people page (our example uses Goddard-W-A-III)
2. Find the list of articles for the Group feed to exclude (ours is LIGO)
3. Get a list of article IDs to exclude
4. Get the people feed you want for your person
5. Loop through the people feed  articles IDs and if not also in the exclude list display it

### Example excluding LIGO article from recent combined.

Below is a HTML fragment that you could include in a web page to present
this modified view of recent combined publications from CaltechAUTHORS.

```HTML
    <div id="articles">
        <!-- This is where our data will land -->
        Building our list...
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js"></script>
    <script>
    let article_titles = document.getElementById("articles");
    let excluded_keys = [];

    /* Initiate getting the keys we want to exclude */
    CL.get("/groups/LIGO/combined.keys", "text/plain", function (src, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        } 
        // convert our "src" value into a list of excluded keys.
        excluded_keys = src.split("\n");

        // Initiate getting the list of recent articles for 0000-0003-0097
        CL.getPeopleJSON("Goddard-W-A-III", "recent/article", function(articles, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            } 
            article_titles.innerHTML = "";
            articles.forEach(function(article) {
                var key = article._Key;
                // NOTE: we need to check to see if the key is excluded
                if (excluded_keys.indexOf(key) == -1) {
                    let elem = document.createElement("div"),
                        h1 = document.createElement("h1"),
                        anchor = document.createElement("a"),
                        div = document.createElement("div");
                    /* Add CSS style to our elements */
                    elem.classList.add("article");
                    h1.classList.add("article-title");
                    div.classList.add("article-abstract");
                    /* Add our content to build up a composit element */
                    anchor.setAttribute("href", article.official_url);
                    anchor.innerHTML = article.title;
                    h1.appendChild(anchor);
                    div.innerHTML = article.abstract;
                    elem.appendChild(h1);
                    elem.appendChild(div);
                    /* Finally, add our element to the list */
                    article_titles.appendChild(elem);
                }
            });
        });
    });
    </script>
```

That example is rendered below--

<div id="articles" >
    <!-- This is where our data will land -->
    Building our list...
</div>

<script src="/scripts/CL-core.js"></script>
<script>
/* Empty our BaseURL string and use the feed content from this deployment */
CL.BaseURL = "";
</script>
<script>
let article_titles = document.getElementById("articles");
let excluded_keys = [];

/* Initiate getting the keys we want to exclude */
CL.get("/groups/LIGO/combined.keys", "text/plain", function (src, err) {
    if (err != "") {
        console.log("ERROR", err);
        return;
    }
    // convert our "src" value into a list of excluded keys.
    excluded_keys = src.split("\n");

    // Initiate getting the list of recent articles for 0000-0003-0097
    CL.getPeopleJSON("Goddard-W-A-III", "recent/article", function(articles, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        article_titles.innerHTML = "";
        articles.forEach(function(article) {
            var key = article._Key;
            // NOTE: we need to check to see if the key is excluded
            if (excluded_keys.indexOf(key) == -1) {
                let elem = document.createElement("div"),
                    h1 = document.createElement("h1"),
                    anchor = document.createElement("a"),
                    div = document.createElement("div");
                /* Add CSS style to our elements */
                elem.classList.add("article");
                h1.classList.add("article-title");
                div.classList.add("article-abstract");
                /* Add our content to build up a composit element */
                anchor.setAttribute("href", article.official_url);
                anchor.innerHTML = article.title;
                h1.appendChild(anchor);
                div.innerHTML = article.abstract;
                elem.appendChild(h1);
                elem.appendChild(div);
                /* Finally, add our element to the list */
                article_titles.appendChild(elem);
            }
        });
    });
});
</script>



