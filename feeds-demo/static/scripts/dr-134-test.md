
# CL-feeds-ui.js test viewer code

This is a test file for testing the "viewer" method of the CL
object on "People" feeds data. It is related to DR-134 bug report 
revising what is displayed for conference events and monographs as well 
as some misc. fixes to bring the output more inline with the templates 
at feeds.library.caltech.edu.

<!-- START: test -->

### Conference items

#### Greer-J-R

Compare with [feeds conference item](https://feeds.library.caltech.edu/people/Greer-J-R/conference_item.html), [json data](https://feeds.library.caltech.edu/people/Greer-J-R/conference_item.json)

<div class="CaltechAUTHORS" id="conference-items">Testing layout of conference items, if you see this Test failed!</div>

#### Newman-D-K

Compare with [feeds conference items](https://feeds.library.caltech.edu/people/Newman-D-K/conference_item.html), [json data](https://feeds.library.caltech.edu/people/Newman-D-K/conference_item.json)

<div class="CaltechAUTHORS" id="conference-items-2">Testing layout of confernce items, if you see this Test failed!</div>

### Monographs, add series name and issue number

#### Geer-J-R

Compare with https://feeds.library.caltech.edu/people/Greer-J-R/monograph.html

<div class="CaltechAUTHORS" id="monograph-items">Testing layout of monograph items, if you see this Test failed!</div>

#### Newman-D-K

Compare with https://feeds.library.caltech.edu/people/Newman-D-K/monograph.html

<div class="CaltechAUTHORS" id="monograph-items-2">Testing layout of monograph items, if you see this Test failed!</div>

<style>
.CaltechAUTHORS .unknown-year { display: none; }
.CaltechAUTHORS li {
    padding-bottom: 0.24em;
    margin-bottom: 0.24em;
    list-style: none;
}
.CaltechAUTHORS a {
    padding-right: 0.24em;
}
.CaltechAUTHORS span {
    padding-right: 0.24em;
}
.CaltechAUTHORS div {
    padding-bottom: 0.24em;
    margin-bottom: 0.24em;
}
</style>

<script src="CL.js"></script>
<script>
(function(document, window) {
  "use strict";
  let cl = Object.assign({}, window.CL),
      cl2 = Object.assign({}, window.CL),
      cl3 = Object.assign({}, window.CL),
      cl4 = Object.assign({}, window.CL),
      config = {},
      conference_elem = document.getElementById("conference-items"),
      monograph_elem = document.getElementById("monograph-items"),
      conference_elem2 = document.getElementById("conference-items-2"),
      monograph_elem2 = document.getElementById("monograph-items-2");
  cl.BaseURL = "https://feeds.library.caltech.edu";
  cl2.BaseURL = "https://feeds.library.caltech.edu";
  cl3.BaseURL = "https://feeds.library.caltech.edu";
  cl4.BaseURL = "https://feeds.library.caltech.edu";

  /* DR-134, conference_item changes */
  config = {
    "aggregation": "people",
    "feed_id": "Greer-J-R",
    "feed_path": "conference_item",
    "recent_n": 0,
    "use_recent": false,
    "feed_count": false,
    "show_year_headings": false,
    "creators": true,
    "pub_date": true,
    "title_link": true,
    "citation_details": true,
    "issn_or_isbn": false,
    "pmcid": false,
    "description": false,
    "developer_mode": false,
    "include_style": true,
    "include_CL": true,
    "repository": "CaltechAUTHORS",
    "css_classname": ".CaltechAUTHORS",
    "filters": []
};

  config.parent_element = conference_elem;
  config.filters.push(cl.normalize_view);
  cl.setAttribute("viewer", config);
  cl.getPeopleJSON("Greer-J-R", "conference_item", function(data, err) {
    cl.viewer(data, err);
  });


  /* DR-134, monograph changes */
  config = {
    "aggregation": "people",
    "feed_id": "Greer-J-R",
    "feed_path": "monograph",
    "recent_n": 0,
    "use_recent": false,
    "feed_count": false,
    "show_year_headings": false,
    "creators": true,
    "pub_date": true,
    "title_link": true,
    "citation_details": true,
    "issn_or_isbn": false,
    "pmcid": false,
    "description": false,
    "developer_mode": false,
    "include_style": true,
    "include_CL": true,
    "repository": "CaltechAUTHORS",
    "css_classname": ".CaltechAUTHORS",
    "filters": []
  };
  config.parent_element = monograph_elem;
  config.filters.push(cl2.normalize_view);
  cl2.setAttribute("viewer", config);
  cl2.getPeopleJSON("Greer-J-R",  "monograph", function(data, err) {
    cl2.viewer(data, err);
  });

  /* DR-134 monograph citation changes */
   config = {
    "aggregation": "people",
    "feed_id": "Newman-D-K",
    "feed_path": "monograph",
    "recent_n": 0,
    "use_recent": false,
    "feed_count": false,
    "show_year_headings": false,
    "creators": true,
    "pub_date": true,
    "title_link": true,
    "citation_details": true,
    "issn_or_isbn": false,
    "pmcid": false,
    "description": false,
    "developer_mode": false,
    "include_style": true,
    "include_CL": true,
    "repository": "CaltechAUTHORS",
    "css_classname": ".CaltechAUTHORS",
    "filters": []
};
  config.parent_element = monograph_elem2;
  config.filters.push(cl.normalize_view);
  cl3.setAttribute("viewer", config);
  cl3.getPeopleJSON("Newman-D-K", "monograph", function(data, err) {
    cl3.viewer(data, err);
  });

  /* DR-134 conference_item citation changes */
   config = {
    "aggregation": "people",
    "feed_id": "Newman-D-K",
    "feed_path": "conference_item",
    "recent_n": 0,
    "use_recent": false,
    "feed_count": false,
    "show_year_headings": false,
    "creators": true,
    "pub_date": true,
    "title_link": true,
    "citation_details": true,
    "issn_or_isbn": false,
    "pmcid": false,
    "description": false,
    "developer_mode": false,
    "include_style": true,
    "include_CL": true,
    "repository": "CaltechAUTHORS",
    "css_classname": ".CaltechAUTHORS",
    "filters": []
};
  config.parent_element = conference_elem2;
  config.filters.push(cl.normalize_view);
  cl4.setAttribute("viewer", config);
  cl4.getPeopleJSON("Newman-D-K", "conference_item", function(data, err) {
    cl4.viewer(data, err);
  });
}(document, window));
</script>


<!--   END: test -->
