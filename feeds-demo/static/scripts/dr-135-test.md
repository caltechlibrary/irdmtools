
# CL-feeds-ui.js test viewer code

This is a test file for testing the "viewer" method of the CL
object. It is related to DR-135 bug report revising what is displayed for
"inpress" items, conference events and monographs.

<!-- START: test -->

### Conference items

Compare with [CCI-Solar-Fuels, conference items](https://feeds.library.caltech.edu/groups/CCI-Solar-Fuels/conference_item.html), [json data](https://feeds.library.caltech.edu/groups/CCI-Solar-Fuels/conference_item.json)

<div class="CaltechAUTHORS" id="conference-items">Testing layout of conference items, if you see this Test failed!</div>

### Monographs, add series name and issue number

Compare with [Applied-&-Computational-Mathematics, monographs](https://feeds.library.caltech.edu/groups/Applied-&-Computational-Mathematics/monograph.html), [json data](https://feeds.library.caltech.edu/groups/Applied-&-Computational-Mathematics/monograph.json)

<div class="CaltechAUTHORS" id="monograph-items">Testing layout of monograph items, if you see this Test failed!</div>

### Handle "In press" and "Submitted" published states

Compare with [Earthquake-Engineering-Research-Laboratory, combined](https://feeds.library.caltech.edu/groups/Earthquake-Engineering-Research-Laboratory/combined.html), [json data](https://feeds.library.caltech.edu/groups/Earthquake-Engineering-Research-Laboratory/combined.json)

<div class="CaltechAUTHORS" id="ispublished-items">Testing layout of items that are "inpress" and "submitted", if you see this Test failed!</div>


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
      config = {},
      conference_elem = document.getElementById("conference-items"),
      monograph_elem = document.getElementById("monograph-items"),
      ispublished_elem = document.getElementById("ispublished-items");
  cl.BaseURL = "https://feeds.library.caltech.edu";
  cl2.BaseURL = "https://feeds.library.caltech.edu";
  cl3.BaseURL = "https://feeds.library.caltech.edu";

  /* DR-135, conference_item changes */
  config = {
    "aggregation": "groups",
    "feed_id": "CCI-Solar-Fuels",
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
  cl.getGroupJSON("CCI-Solar-Fuels", "conference_item", function(data, err) {
    cl.viewer(data, err);
  });


  /* DR-135, monograph changes */
  config = {
    "aggregation": "groups",
    "feed_id": "Applied-&-Computational-Mathematics",
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
  cl2.getGroupJSON("Applied-&-Computational-Mathematics", "monograph", function(data, err) {
    cl2.viewer(data, err);
  });

  /* DR-135 ispublished citation changes */
   config = {
    "aggregation": "groups",
    "feed_id": "Earthquake-Engineering-Research-Laboratory",
    "feed_path": "combined",
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
  config.parent_element = ispublished_elem;
  config.filters.push(cl.normalize_view);
  cl3.setAttribute("viewer", config);
  cl3.getGroupJSON("Earthquake-Engineering-Research-Laboratory", "combined", function(data, err) {
    cl3.viewer(data, err);
  });
}(document, window));
</script>


<!--   END: test -->
