
# Samples, Examples, and Formats

## Working with Custom Lists

While it is common to use publication date descending lists of articles,
books and other publications on websites sometimes you want to just
list specific works. This is what custom lists are for. These can be easily
implemented for both [groups](/groups) and [people](/people) if you
known the group or profile ids, the list you want to filter and the specific ids and order you want to display that list in.

For all custom lists we will be working with the JSON version of the feeds.
The [CL.js](/scripts/CL.js) library provides two functions for working
with custom likes-- `CL.getGroupCustomJSON()` and `CL.getPeopleCustomJSON()`. These work similarly to how `CL.getGroupJSON()` and `CL.getPeopleJSON()`. The primary difference is an extra parameter containing the list of ids in the order you want the items to be displayed.

## Working with Custom Group lists

### Problem display a custom list of works from Caltech Library group

The [Caltech Library](/groups/Caltech-Library) has a list of monographs and we'd like to take and display a subset. To do that we need to know the list of monograph ids which leads to two more questions.  How do we assemble the list?  How do we get it to display in our webpage?

#### What we know

Group ID
: The Caltech Library's group id is "Caltech-Library"

CL.js
: CL.js has a function `CL.getGroupCustomJSON()` for displaying custom lists

The feed path
: We want to display monographs so the feed name is "monograph"

How do we get the list of monograph ids?

#### Solution

The solution has two steps

1. Get your list of monograph ids
2. Assemble some HTML and JavaScript to display the list

Getting the list of monograph ids can be done one of several ways. The first though not necessarily the fastest is to review the monographs feed page, open each one in their own browser tab and follow the links back to CaltechAUTHORS landing page and find the "ID Code" listed for each. It takes a few mouse clicks but this doesn't require knowledge of JSON other than we want to save the numbers so we can type up a list of them as a JSON array.

In this example I've picked some ids from the monographs list. The JSON array is simply a comma separate list of id numbers with a leading and closing square bracket.

```javascript
    [ 25887, 25887, 25900, 25889 ]
```

Step two, in our web page include the following.

```html
    <div id="group-custom-list">
    The custom group list we've currated should appear below.
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* We save our list of monograph ids as id_list */
    let id_list = [ 25887, 25927, 25900, 25889 ];

    /* We get a handle on our page element */
    let group_custom_list = document.getElementById("group-custom-list");

    /* Now we call getGroupCustomJSON() to display our custom list */
    CL.getGroupCustomJSON("Caltech-Library", "monograph", id_list,
        function(monographs, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            monographs.forEach(function(monograph) {
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
                group_custom_list.appendChild(elem);
            });
        });
    </script>
```

##### Solution output

**Start of solution output**

<div id="group-custom-list">
The custom group list we've curated should appear below.
</div>

<script src="/scripts/CL-core.js"></script>
<script>CL.BaseURL = "";</script>
<script>
/* We save our list of monograph ids as id_list */
let id_list = [ 25887, 25927, 25900, 25889 ];

/* We get a handle on our page element */
let group_custom_list = document.getElementById("group-custom-list");

/* Now we call getGroupCustomJSON() to display our custom list */
CL.getGroupCustomJSON("Caltech-Library", "monograph", id_list,
    function(monographs, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        monographs.forEach(function(monograph) {
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
            group_custom_list.appendChild(elem);
        });
    });
</script>

**End of solution output**



## Working with Custom People lists

### Problem, display a custom list of works from a people

In this example I want to display a list of one article, one book 
section, one conference item and one monograph for Prof. Newman.
How to I do this?

#### What we know

People ID
: Prof. Newman's Caltech Library's people id is "Newman-D-K"

CL.js
: CL.js has a function `CL.getPeopleCustomJSON()` for displaying custom lists

The feed path
: Since we're picking across publication types we'll want to use "combined"

How do we get the list of ids?

#### Solution

The solution has two steps

1. Get your list of ids
2. Assemble some HTML and JavaScript to display the list

In this example the easiest way to proceed it to first lookup the article, book section, conference item and monograph ids in their respective feeds. If we click through to the CaltechAUTHORS land page for the documents we can make a note of the "ID Code" values. We're going to use these id numbers and create a JSON array with.


```javascript
    [ 92544, 77592, 82309, 91271 ]
```

Since we're working with more than one publication type we'll use the "combined" feed as our data source. It includes all the individual items from the various CaltechAUTHORS publication types.

Step two, in our web page include the following.

```html
    <div id="people-custom-list">
    The custom people list we've currated should appear below.
    </div>
    <script src="https://feeds.library.caltech.edu/scripts/CL.js">
    </script>
    <script>
    /* We save our list of ids as id_list */
    let id_list = [ 92544, 77592, 82309, 91271 ];

    /* We get a handle on our page element */
    let people_custom_list = document.getElementById("people-custom-list");

    /* Now we call getPeopleCustomJSON() to display our custom list */
    CL.getPeopleCustomJSON("Newman-D-K", "combined", id_list,
        function(publications, err) {
            if (err != "") {
                console.log("ERROR", err);
                return;
            }
            publications.forEach(function(publication) {
                let elem = document.createElement("div"),
                    h1 = document.createElement("h1"),
                    anchor = document.createElement("a");

                /* Add CSS classes for styling */
                elem.classList.add("publication");
                h1.classList.add("publication-title");
                /* Now we collect our content into our composit
                   HTML elements */
                anchor.setAttribute("href", publication.official_url);
                anchor.innerHTML = publication.title;
                /* Assemble our elements */
                h1.appendChild(anchor);
                elem.appendChild(h1);
                /* Finally add our elements to our list */
                people_custom_list.appendChild(elem);
            });
        });
    </script>
```

##### Solution output

**Start of solution output**

<div id="people-custom-list">
The custom group list we've curated should appear below.
</div>

<script>
/* We save our list of ids as id_list */
id_list = [ 92544, 77592, 82309, 91271 ];

/* We get a handle on our page element */
let people_custom_list = document.getElementById("people-custom-list");

/* Now we call getPeopleCustomJSON() to display our custom list */
CL.getPeopleCustomJSON("Newman-D-K", "combined", id_list,
    function(publications, err) {
        if (err != "") {
            console.log("ERROR", err);
            return;
        }
        publications.forEach(function(publication) {
            let elem = document.createElement("div"),
                h1 = document.createElement("h1"),
                anchor = document.createElement("a");

            /* Add CSS classes for styling */
            elem.classList.add("publication");
            h1.classList.add("publication-title");
            /* Now we collect our content into our composit
               HTML elements */
            anchor.setAttribute("href", publication.official_url);
            anchor.innerHTML = publication.title;
            /* Assemble our elements */
            h1.appendChild(anchor);
            elem.appendChild(h1);
            /* Finally add our elements to our list */
            people_custom_list.appendChild(elem);
        });
    });
</script>

**End of solution output**



