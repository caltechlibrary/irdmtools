
# Meet the Builder Widget

The Builder Widget will generate HTML, CSS and JavaScript so you can 
include library content in your webpage. 
It needs to know whether you are interested in publications from groups and
people. [Groups](/groups/) is a curated list of Caltech
groups (mostly research groups). [People](/people/) is a curated list 
of Caltech related individuals including faculty, researchers and alumni. 
If you or your group is not listed, email library@caltech.edu for possible inclusion.

## Select a data source

Steps to select the source of the data to display

1. Select your aggregation (groups, people)
2. Select your feed (e.g. name from group or people)
3. Select the type of feed (e.g. combined, article, dataset, model)

## Turn on filtering if desired

By default all publications in the feed type are included. 
You can filter the list so it is shorter.  The most basic filter is 
"recent N" which lets you only show the most recent N number of 
publications.  E.g. if you only wanted to show the your last three 
articles you would check "recent N" and set the number to 3.

## Choose your layout

If you are showing a large number publications your webpage can get pretty 
long to read. One way to manage this is to only show titles with links to the 
publication page.  For shorter lists you want also want to show abstracts, 
authorship or publication dates.  These can all be selected in the layout section of the Builder Widget.

If you'd like you list to be broken up by publication year then
make sure you check "Year Headings" in the Builder Widget.

## See a preview of your list

If you press the "Preview List" button you'll see example output from the code we'd generate if you press the "Generate Code" button.

## Get the code

Finally when all the questions are answered the Widget will generate a fragment of HTML and JavaScript suitable for pasting into your webpage.



<section id="builder-widget" class="widget"><!-- This is where "the widget" should display --></section>

<noscript>JavaScript is required to display and use the Builder Widget</noscript>

<script src="/scripts/CL.js"></script>

<script src="/scripts/CL-BuilderWidget.js"></script>

<script>
(function (document, window) {
    let cl = Object.assign({}, window.CL),
        widget_element = document.getElementById("builder-widget");

    /* NOTE: We want the builder to be hosted 
     * where our code is deployed */
    cl.BaseURL = "";
    cl.BuilderWidget(widget_element);
}(document, window));
</script>
