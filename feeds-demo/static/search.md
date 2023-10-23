
# Feeds Search

NOTE: This page targets searching in our feeds content not in repository records.

<link href="./pagefind/pagefind-ui.css" rel="stylesheet">
<script src="./pagefind/pagefind-ui.js" type="text/javascript"></script>
<div id="search"></div>
<script>
    window.addEventListener('DOMContentLoaded', (event) => {
        let pse = new PagefindUI({ element: "#search" }),
            page_url = new URL(window.location.href),
            query_string = page_url.searchParams.get('q');
        if (query_string !== null) {
            console.log('Query string: ' + query_string);
            pse.triggerSearch(query_string);
        }
    });
</script>



