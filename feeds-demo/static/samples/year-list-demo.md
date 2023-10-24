[samples](./)

# List with Year Headings.

In this example we display a article list with year headings for
the people id "Feynman-R-P". It uses a configuration object and
standard viewer to render the list. Included is an example of 
CSS, HTML and JavaScript to include in your webpage.

```html
    <style>
    
    .CaltechAUTHORS .jump-list {
        padding-bottom: 0.24em;
        margin-bottom: 0.24em;
        border-bottom: solid 0.24em black;
    }
    
    
    .CaltechAUTHORS .jump-list-label {
        padding-left:0.24em;
        padding-right:0.24em;
        border-right: solid 0.12em black;
    }
    
    .CaltechAUTHORS .jump-list:last-child {
        border-right: none;
    }
    
    .CaltechAUTHORS li {
        padding-bottom: 0.24em;
        list-style: none;
    }
    
    .CaltechAUTHORS span {
        padding-right: 0.24em;
    }
    </style>
    
    <section id="year-list-demo"></section>
    
    <script src="/scripts/CL.js"></script>
    <script>
    (function (document, window) {
        "use strict";
        let cl = Object.assign({}, window.CL),
            config = {},
            section = document.getElementById("year-list-demo");;
    
        config.parent_element = section;
        config.filters = [ cl.normalize_view ];
        config.feed_count = true;
        config.year_headings = true;
        config.creators = true;
        config.title_link = true;
        config.pub_date = true;
        config.citation = true;
        config.doi = true;
        config.description = false;
        cl.setAttribute("viewer", config);
    
        cl.getPeopleJSON("Feynman-R-P", "article", function(data, err) {
            cl.viewer(data, err);
        });
    }(document, window));
    
    </script>
```



<!-- START: Year List Demo -->

<style>

.CaltechAUTHORS .jump-list {
    padding-bottom: 0.24em;
    margin-bottom: 0.24em;
    border-bottom: solid 0.24em black;
}


.CaltechAUTHORS .jump-list-label {
    padding-left:0.24em;
    padding-right:0.24em;
    border-right: solid 0.12em black;
}

.CaltechAUTHORS .jump-list:last-child {
    border-right: none;
}

.CaltechAUTHORS li {
    padding-bottom: 0.24em;
    list-style: none;
}

.CaltechAUTHORS span {
    padding-right: 0.24em;
}
</style>



<section id="year-list-demo"></section>

<script src="/scripts/CL-core.js"></script>
<script src="/scripts/CL-ui.js"></script>

<script>
(function (document, window) {
    "use strict";
    let cl = Object.assign({}, window.CL),
        config = {},
        section = document.getElementById("year-list-demo");;

    cl.BaseURL = "";
    config.parent_element = section;
    config.filters = [ cl.normalize_view ];
    config.feed_count = true;
    config.year_headings = true;
    config.creators = true;
    config.title_link = true;
    config.pub_date = true;
    config.citation = true;
    config.doi = true;
    config.description = false;
    cl.setAttribute("viewer", config);

    cl.getPeopleJSON("Feynman-R-P", "article", function(data, err) {
        cl.viewer(data, err);
    });
}(document, window));

</script>

<!--   END: Year List Demo -->


