

# CL-SearchWidget test


Show search page for Big Bear Solar Observatory.


<script src="https://unpkg.com/lunr/lunr.js"></script>

<script src="CL.js"></script>
<script src="CL-feeds-lunr.js"></script>

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

<div id="Big-Bear-Solar-Observatory" class="CaltechAUTHORS"></div>

<script src="https://feeds.library.caltech.edu/scripts/CL.js"></script>
<script src="https://feeds.library.caltech.edu/scripts/CL-feeds-lunr.js"></script>
<script>
(function(document, window) {
  "use strict";
  let cl = Object.assign({}, window.CL),
      config = {},
      q = '',
      u = new URL(window.location.href),
      elem = document.getElementById("Big-Bear-Solar-Observatory"),
      query_form = document.createElement('div');

  if (u.search !== '') {
      q = u.searchParams ? (u.searchParams).get('q') : '';
  }
  query_form.innerHTML = `
<form method="get">
  <input type="text" name="q" value="${q}" placeholder="Enter search terms">
  <button>Search</button>
</form>
`;
  elem.appendChild(query_form);
    cl.BaseURL = "https://feeds.library.caltech.edu";

  config = {
    "aggregation": "groups",
    "feed_id": "Big-Bear-Solar-Observatory",
    "feed_path": "combined",
    "feed_count": false,
    "creators": false,
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
    "recent_n": 0,
    "filters": []
};
  config.show_search_box = true;
  config.parent_element = elem;
  config.filters.push(cl.normalize_view);
  config.filters.push(cl.lunr_search);
  cl.setAttribute("viewer", config);
  cl.getGroupJSON("Big-Bear-Solar-Observatory", "combined", function(data, err) {
     cl.viewer(data, err);
  });
}(document, window));
</script>
