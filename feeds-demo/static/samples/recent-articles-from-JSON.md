
# Sample, Examples and formats

## Creating a list of recent articles from JSON feed using an person id

### Whay you need to know to get started

1. Know the persond ID, we'll be using Newman-D-K in our example
2. Find your feed on https://feeds.library.caltech.edu/people
    + E.g. for Newman-D-K the URL would be https://feeds.library.caltech.edu/people/Newman-D-K
3. Decide which feed you want to include in your webpage
    + Articles, e.g. https://feeds.library.caltech.edu/people/Newman-D-K/article.html 
    + Combined Publications, e.g. https://feeds.library.caltech.edu/people/Newman-D-K/combined.html 
    + Recent Articles, e.g. https://feeds.library.caltech.edu/people/Newman-D-K/recent/article.html
    + Recent Combined Publications, e.g. https://feeds.library.caltech.edu/people/Newman-D-K/recent/combined.html
4. In your web page include the JavaScript library https://feeds.library.caltech.edu/scripts/CL.js
5. Adjust the example SCRIPT element to reflect your person ID, feed type and formatting you'd like in your web page.

### Example for Newman-D-K

Below is a HTML fragment that you could include in a web page to include recent articles titles for 
the person id Newman-D-K based on the JSON version of the feed.

```HTML
    <div id="recent-articles">
    Recent articles go here if JavaScript worked!
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js"></script>
    <script>
    let recent_articles = document.getElementById("recent-articles");
    
    CL.getPeopleJSON("Newman-D-K", "recent/article", function(articles, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        articles.forEach(function(article) {
            var elem = document.createElement("div"),
                h1 = document.createElement("h1"),
                anchor = document.createElement("a"),
                div = document.createElement("div");
            /* Setup to style our elements with CSS classes */
            elem.classList.add("article");
            h1.classList.add("article-title");
            div.classList.add("article-abstract");
            /* Not layout our data in our elements */
            anchor.setAttribute("href", article.official_url);
            anchor.innerHTML = article.title;
            h1.appendChild(anchor);
            div.innerHTML = article.abstract;
            elem.appendChild(h1);
            elem.appendChild(div);
            /* Finally add our composit element to the list */
            recent_articles.appendChild(elem)
        });
    });
    </script>
```

That example is rendered below--

<div id="recent-articles">See 
  Recent Articles go here if JavaScript worked!
</div>

<script src="/scripts/CL-core.js"></script>
<script>
let recent_articles = document.getElementById("recent-articles");

CL.getPeopleJSON("Newman-D-K", "recent/article", function(articles, err) {
    if (err != "") {
        console.log("ERROR", err);
        return;
    }
    articles.forEach(function(article) {
        var elem = document.createElement("div"),
            h1 = document.createElement("h1"),
            anchor = document.createElement("a"),
            div = document.createElement("div");
        /* Setup to style our elements with CSS classes */
        elem.classList.add("article");
        h1.classList.add("article-title");
        div.classList.add("article-abstract");
        /* Not layout our data in our elements */
        anchor.setAttribute("href", article.official_url);
        anchor.innerHTML = article.title;
        h1.appendChild(anchor);
        div.innerHTML = article.abstract;
        elem.appendChild(h1);
        elem.appendChild(div);
        /* Finally add our composit element to the list */
        recent_articles.appendChild(elem)
    });
});
</script>

