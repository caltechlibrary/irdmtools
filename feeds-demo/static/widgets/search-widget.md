
# Meet the Search Widget

The Search Widget will generate HTML and JavaScript so your included JSON feeds can be searchable from your web browser. It uses [Lunrjs](https://lunrjs.com) to help you build both the indexes and deploy a search interface. The first step is to select a data source and this then select the elements you want to appear in the search results.


## Select a data source

Steps to select the source of the data to display

1. Select your aggregation (groups, people)
2. Select your feed (e.g. a groups or people)
3. Select the type of feed (e.g. combined, article, dataset, model)

## Choose your layout

If you are showing a large number publications your webpage can get pretty long to read. One way to manage this is to only show titles with links to the publication page.  For shorter lists you want also want to show abstracts, authorship or publication dates.  These can all be selected in the layout section of the Builder Widget.

If you'd like you list to be broken up by publication year then
make sure you check "Year Headings" in the Builder Widget.

## Get the code

Finally when all the questions are answered the Widget will generate a fragment of HTML and JavaScript suitable for pasting into your webpage. 


<!-- START: Search Widget Demo -->

<noscript>JavaScript is required to display and use the Search Widget</noscript>

<section id="search-widget" class="widget">
<!-- This is where "the widget" should display -->
</section>


<script src="https://unpkg.com/lunr/lunr.js"></script>

<script src="../scripts/CL.js"></script>

<script src="../scripts/CL-feeds-lunr.js"></script>

<script src="../scripts/CL-SearchWidget.js"></script>

<script>
(function (document, window) {
    let cl = Object.assign({}, window.CL),
        widget_element = document.getElementById("search-widget");

    /* NOTE: We want the Search Widget to be hosted
     * where our code is deployed */
    cl.BaseURL = "";
    cl.SearchWidget(widget_element);
}(document, window));
</script>



<!--   END: Search Widget Demo -->
