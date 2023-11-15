
# CL-feeds-ui.js test viewer code

This is a test file for testing the "viewer" method of the CL
object. It is related to DR-139 bug report. Visually inspect the output
and make sure you can see a list of combined articles for matching articles listed at [Thirty Meter Telescope](https://feeds.library.caltech.edu/groups/Thirty-Meter-Telescope/combined.html).


<!-- START: test -->

<div id="Thirty-Meter-Telescope">Test failed!, combined article list should go here.</div>


<script src="CL.js"></script>
<script>
(function(document, window) {
  "use strict";
  let cl = Object.assign({}, window.CL),
      config = {},
      elem = document.getElementById("Thirty-Meter-Telescope");
  cl.BaseURL = "https://feeds.library.caltech.edu";

  config = {
    "aggregation": "groups",
    "feed_id": "Thirty-Meter-Telescope",
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
  config.parent_element = elem;
  config.filters.push(cl.normalize_view);
  cl.setAttribute("viewer", config);
  cl.getGroupJSON("Thirty-Meter-Telescope", "combined", function(data, err) {
    cl.viewer(data, err);
  });
}(document, window));
</script>


<!--   END: test -->
