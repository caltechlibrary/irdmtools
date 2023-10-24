
# CL-feeds-ui.js test Advisor feed

GitHub issue #4 addresses adding advisor feeds to the Widget
and getPeopleJSON function. We should see ia list Advisors. 

<!-- START: test -->

## Advisor Feed for Abu-Mostafa Yaser

### Advisor for

[Thesis and Dissertations](https://feeds.library.caltech.edu/people/Abu-Mostafa-Y-S/advisor.html), [json data](https://feeds.library.caltech.edu/people/Abu-Mostafa-Y-S/advisor.json)

Compare with https://feeds.library.caltech.edu/people/Abu-Mostafa-Y-S/advisor.html

<div class="CaltechAUTHORS" id="thesis-items">Testing layout of thesis list
, if you see this Test failed!</div>

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
<!-- 
<script src="CL-core.js"></script>
<script src="CL-feeds.js"></script>
<script src="CL-feeds-ui.js"></script>
-->
<script>
(function(document, window) {
  "use strict";
  let cl = Object.assign({}, window.CL),
      config = {},
      thesis_elem = document.getElementById("thesis-items");

  /* issue-4, get a list of thesis and dissertation Yaser advised on */
   config = {
    "aggregation": "people",
    "feed_id": "Abu-Mostafa-Y-S",
    "feed_path": "advisor",
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
  config.parent_element = thesis_elem;
  config.filters.push(cl.normalize_view);
  cl.setAttribute("viewer", config);
  cl.getPeopleJSON("Abu-Mostafa-Y-S", "advisor", function(data, err) {
    cl.viewer(data, err);
  });
}(document, window));
</script>

<!--   END: test -->

