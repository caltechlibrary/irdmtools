
# Samples, Examples, and Formats

## Filtering lists

If you want to display a subset of a list you can control what is
added to your webpage using any of the JSON get functions. The process
flow is similar. Pick your list, pick the appropriate get function (e.g.
`CL.getPeopleJSON()` or `CL.getGroupJSON()` and then add the filter
on the records returned.

## Working with Group lists

### Problem display a list of publications with DOIs for a group

We want to display a list of monographs with a DOI for the [Keck Institute for Space Studies](/groups/Keck-Institute-for-Space-Studies).  

#### What we know

Group ID
: "Keck-Institute-for-Space-Studies" is the group id we need

CL.js
: CL.js has a function `CL.getGroupJSON()` to use to display lists

The feed path
: We want to display monographs so the feed name is "monograph"

#### Solution

The main challenge is in understanding the records we're getting back  in our list. The based way to do get a sense of how to test the objects representing the records is to look at the JSON feed it self. The structure is an array of objects where each object is a publication. In this case they are monographs.  Most modern web browsers have development tools built in. You can use these to pretty print the JSON to make it easier to understand.  You can also use a web browser plugin like [JSONView](https://jsonview.com) to do a similar thing. In our case each object that has a doi also has a "doi" field.  We check to each object to see if this field exists (older records might not have it) and to make sure that it is not an empty string. If those two conditions are met then we can choose to display that object on the webpage.

We can break our problem down to two steps

1. Find the field we want to use to filter on
2. Write the filter function in our callback

Here is a JSON fragment to illustrate what it looks like. The three
dots just mean we're skipping over some of what is printed out.

```JSON
    [
    {
        ...
    "_Key": "91407",
        ...
    "collection": "CaltechAUTHORS",
    "creators": {
        ...
    },
        ...
    "doi": "10.26206/HQ9P-YW49",
        ...
    },
        ...
    ]
```

Note the field called "doi".

Step two, in our web page include the following.

```html
    <div id="group-filter-list">
    The filter records by the presense of a DOI field.
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* We get a handle on our page element */
    let group_filter_list = document.getElementById("group-filter-list");

    /* Now we call getGroupJSON() to display our filtered list */
    CL.getGroupJSON("Keck-Institute-for-Space-Studies", "monograph", 
        function(monographs, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            /* Now we want to loop through our monographs
               and filter for  the ones we want to display */
            monographs.forEach(function(monograph) {
                if ("doi" in monograph && monograph.doi != "") {
                    let elem = document.createElement("div"),
                        h1 = document.createElement("h1"),
                        anchor = document.createElement("a");

                    /* Add CSS classes for styling */
                    elem.classList.add("monograph");
                    h1.classList.add("monograph-title");
                    /* Now we collect our content into our composit
                       HTML elements */
                    anchor.setAttribute("href", monograph.official_url);
                    anchor.innerHTML = monograph.title;
                    /* Assemble our elements */
                    h1.appendChild(anchor);
                    elem.appendChild(h1);
                    /* Finally add our elements to our list */
                    group_filter_list.appendChild(elem);
                }
            });
        });
    </script>
```

##### Solution output

**Start of solution output**

<div id="group-filter-list">
The custom group list we've curated should appear below.
</div>

<script src="/scripts/CL-core.js"></script>
<script>CL.BaseURL = "";</script>
<script>
/* We get a handle on our page element */
let group_filter_list = document.getElementById("group-filter-list");

/* Now we call getGroupJSON() to display our filtered list */
CL.getGroupJSON("Keck-Institute-for-Space-Studies", "monograph", 
    function(monographs, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        /* Now we want to loop through our monographs
           and filter for  the ones we want to display */
        monographs.forEach(function(monograph) {
            if ("doi" in monograph && monograph.doi != "") {
                let elem = document.createElement("div"),
                    h1 = document.createElement("h1"),
                    anchor = document.createElement("a");

                /* Add CSS classes for styling */
                elem.classList.add("monograph");
                h1.classList.add("monograph-title");
                /* Now we collect our content into our composit
                   HTML elements */
                anchor.setAttribute("href", monograph.official_url);
                anchor.innerHTML = monograph.title;
                /* Assemble our elements */
                h1.appendChild(anchor);
                elem.appendChild(h1);
                /* Finally add our elements to our list */
                group_filter_list.appendChild(elem);
            }
        });
    });
</script>

**End of solution output**
