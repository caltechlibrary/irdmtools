
# CL-feeds-ui.js test

This is a test file for putting CL-feeds-ui.js through its paces 
and to confirm we have a working `CL` object.

<style>
#status {
    font-size: 1em;
}
.CaltechAUTHORS {
    width: 60%;
    border: solid 0.12em blue;
    margin: 2em;
}
.jump-list {
    border-bottom: solid 0.12em black;
    margin-bottom: 1em;
}
.jump-list-label {
    margin-right: 0.24em;
}
.book-title {
    font-weight: bolder;
    padding-right: 0.42em;
}
</style>

<div id="test-output"></div>

<code><pre id="status"></pre></code>


<!-- START: test sequence for CL-core.js -->

<script src="CL-core.js"></script>
<script src="CL-feeds.js"></script>
<script src="CL-feeds-ui.js"></script>
<script>
(function (document, window) {
"use strict";
let cl = Object.assign({}, window.CL),
    status = document.getElementById("status");

function println(...s) {
    s.forEach(function(s) {
        console.log(s);
        status.append(s + "\n");
    });
}

/*
 * Run the following test sequences
 */
function testRecentNAndViewer(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    tests.count++;
    let books = Object.assign({}, CL),
        div = document.getElementById("test-output"),
        n = 0,
        viewer_config = {},
        test_config = {};

    books.setAttribute("recentN", 3);
    n = books.getAttribute("recentN");
    if (n !== 3) {
        println("WARNING: expected 3 got " + n);
        tests.warnings++;
    }
    viewer_config.parent_element = div;
    viewer_config.filters = [ books.recentN, books.normalize_view ];
    viewer_config.show_year_headings = true;
    viewer_config.title_link = true;
    viewer_config.pub_date = true;
    viewer_config.creators = false;
    viewer_config.description = false;
    books.setAttribute("viewer", viewer_config);

    test_config = books.getAttribute("viewer");
    for (let key in viewer_config) {
        if (test_config[key] === undefined) {
            println("FAILED: config error for " + key + ", expected "+viewer_config[key], ", got " + test_config[key]);
            self.nextCallbackFn(tests, "");
            return;
        }
    }

    books.getPeopleJSON("Newman-D-K", "book_section", function(data, err) {
        books.viewer(data, err);
        /*console.log(div); */
        /*NOTE: this is a minimal test, really need to check the whole DOM tree created. */
        if (div.childNodes.length !== 5) {
            tests.errors++;
            println("FAILED: Expected five child nodes got " + div.childNodes.length);
            console.log("FAILED: div should have <DIV> and <UL> sibblings ", div);
            testSummary(tests, "");
            return;
        }
        let elem = div.firstChild;
        if (elem === undefined) {
            tests.errors++;
            println("FAILED: Expected firstChild to be element got undefined");
            testSummary(tests, "");
            return;
        }
        if (elem.childNodes.length !== 2) {
            tests.errors++;
            println("FAILED: Expected ul element to have two children, got " + elem.childNodes.length);
            console.log("FAILED: ul should have li inside ", elem);
            testSummary(tests, "");
            return;
        }
        /* Get the next sibling after the year the heading */
        elem = div.firstChild.nextSibling;
        elem = elem.nextSibling.firstChild;
        if (elem.childNodes.length !== 3) {
            tests.errors++;
            println("FAILED: Expected li element to have three spans as children", elem.childNodes.length);
            console.log("FAILED: li should have two spans inside ", elem);
            testSummary(tests, "");
            return;
        }

        tests.success++;
        testSummary(tests, "");
    });
}

function testSummary(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    println("Failures: " + tests.errors);
    println("Warnings: " + tests.warnings);
    println("Successful: " + tests.success + "/" + tests.count);
}

/* Run the rest of the tests in a pipeline */
let tests = {
        "success": 0,
        "warnings": 0,
        "errors": 0,
        "count": 0
    };
cl.pipeline(tests, "", 
    testRecentNAndViewer,
    testSummary);
}(document, window));

</script>

<!--   END: test sequence for CL-feeds-ui.js -->
