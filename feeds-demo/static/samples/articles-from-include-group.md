
# Sample, Examples and formats

## Creating a list of articles from include feed using a group id

### What you need to know to get started

1. Find your feed on https://feeds.library.caltech.edu/group
2. Decide which feed you want to include in your webpage
    + Articles, e.g. https://feeds.library.caltech.edu/groups/Keck-Institute-for-Space-Studies/article.html
    + Combined Publications, e.g. https://feeds.library.caltech.edu/groups/Keck-Institute-for-Space-Studies/combined.html
    + Recent Articles, e.g. https://feeds.library.caltech.edu/groups/Keck-Institute-for-Space-Studies/recent/article.html
    + Recent Combined Publications, e.g. https://feeds.library.caltech.edu/groups/Keck-Institute-for-Space-Studies/recent/combined.html
3. Copy the code snippet below to your web page.
4. Adjust the example snippet to reflect the group ID, feed type and formatting you'd like in your web page.

### Example for Keck Institute for Space Studies

Below is a HTML fragment that you could include in a web page to include
articles for the Keck Institute for Space Studies.

```HTML
    <div id="feed">
        See <a href="/Keck-Institute-for-Space-Studies/combined.html">https://feeds.library.caltech.edu/groups/Keck-Institute-for-Space-Studies/combined.html</a>.
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js"></script>
    <script>
    let article_list = document.getElementById('feed');

    CL.getFeed("https://feeds.library.caltech.edu/groups/Keck-Institute-for-Space-Studies/combined.include", function(src, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        article_list.innerHTML = src;
    });
</script>
```

Now we can see the resulting bibliography below--

<div id="feed">
    See <a href="/Keck-Institute-for-Space-Studies/combined.html">https://feeds.library.caltech.edu/groups/Keck-Institute-for-Space-Studies/combined.html</a>.
</div>

<script src="/scripts/CL-core.js"></script>
<script>CL.BaseURL = "";</script>
<script>
let article_list = document.getElementById('feed');

CL.getFeed("/groups/Keck-Institute-for-Space-Studies/combined.include", function(src, err) {
    if (err != "") {
        console.log("ERROR", err);
        return;
    }
    article_list.innerHTML = src;
});
</script>
