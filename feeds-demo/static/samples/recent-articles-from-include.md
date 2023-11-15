
# Sample, Examples and formats

## Creating a list of recent articles from include feed using a person id

### What you need to know to get started

1. Know your Caltech Library person ID, we'll be using Newman-D-K in our example below
2. Find your feed on https://feeds.library.caltech.edu/people
    + E.g. for Newman-D-K the URL would be https://feeds.library.caltech.edu/people/Newman-D-K
3. Decide which feed you want to include in your webpage
    + Articles, e.g. https://feeds.library.caltech.edu/people/Newman-D-K/article.html
    + Combined Publications, e.g. https://feeds.library.caltech.edu/people/Newman-D-K/combined.html
    + Recent Articles, e.g. https://feeds.library.caltech.edu/people/Newman-D-K/recent/article.html
    + Recent Combined Publications, e.g. https://feeds.library.caltech.edu/people/Newman-D-K/recent/combined.html
4. In your web page include the JavaScript library https://feeds.library.caltech.edu/scripts/CL.js
5. Adjust the example SCRIPT element to reflect your ORCID ID, feed type and formatting you'd like in your web page.

### Example for Newman-D-K

Below is a HTML fragment that you could include in a web page to include recent articles titles for Newman-D-K.

```HTML
    <div id="recent-articles">
    Recent Article go here if JavaScript worked!
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js"></script>
    <script>
    let article_list = document.getElementById('recent-articles');

    CL.getPeopleInclude("Newman-D-K", "recent/article", function(src, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        article_list.innerHTML = src;
    });
    </script>
```

Now we can see the resulting bibliography below--

<div id="recent-articles">
    Recent Article go here if JavaScript worked!
</div>

<script src="/scripts/CL-core.js"></script>
<script>CL.BaseURL = "";</script>
<script>
let article_list = document.getElementById('recent-articles');

CL.getPeopleInclude("Newman-D-K", "recent/article", function(src, err) {
    if (err != "") {
        console.log("ERROR", err);
        return;
    }
    article_list.innerHTML = src;
});
</script>
