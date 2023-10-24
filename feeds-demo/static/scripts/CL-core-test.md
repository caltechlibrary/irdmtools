
# CL-core.js test

This is a test file for putting CL-core.js through its paces and 
to confirm we have a working `CL` object.


<style>
#status {
    font-size: 1em;
}
</style>
<code><pre id="status"></pre></code>

<!-- START: test sequence for CL-feeds.js -->

<script src="CL-core.js"></script>
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
println("Testing pipeline()");
let func_cnt = 0;
    
function hello_one(data, err) {
    let self = this;
    if (err !== "") {
        println("FAILED hello_one: " + err);
        return;
    }
    if (data.one === undefined || data.one !== 1) {
        println('FAILED expected {"one": 1}', JSON.stringify(data));
        return;
    }
    func_cnt++;
    self.nextCallbackFn({"two": 2}, err);
}

function hello_two(data,err) {
    let self = this;
    if (err !== "") {
        println("FAILED hello_two: " + err);
        return;
    }
    if (data.two === undefined || data.two !== 2) {
        println('FAILED expected {"two": 2}', JSON.stringify(data));
        return;
    }
    func_cnt++;
    self.nextCallbackFn({"three": 3}, err);
}

cl.pipeline({"one":1}, "", hello_one, hello_two);
if (func_cnt !== 2) {
    println("FAILED expected func_cnt of 2", func_cnt);
    return;
}
println("Testing pipeline(), OK");

println("\nRunning tests in a pipeline\n");

function testAttributes(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    tests.count++;
    println("Testing attribute handling");
    if (cl._attributes !== undefined) {
        println("WARNING: expected cl._attributes to be undefined", typeof cl._attributes, cl._attributes);
        tests.warnings++;
    }
    cl.setAttribute("helloworld", { "one": 1, "two": 2});
    if (cl._attributes === undefined) {
        println("ERROR: expected cl._attributes to exist", typeof cl._attributes);
        tests.errors++;
        self.nextCallbackFn(tests, err);
        return;
    }
    if (cl.hasAttribute("helloworld") == false) {
        println("Expected cl. attributes to have a 'helloworld' attribute");
        tests.errors++;
        self.nextCallbackFn(tests, err);
        return;
    }
    let helloworld = cl.getAttribute("helloworld");
    if (helloworld === undefined) {
        println("Expected cl.getAttribute('helloworld') to return object.")
        tests.errors++;
        self.nextCallbackFn(tests, err);
        return;
    }
    if (helloworld.one === undefined || helloworld.one !== 1) {
        console.log("Expected helloworld to have a one attribute holding 1.", helloworld);
        tests.errors++;
        self.nextCallbackFn(tests, err);
        return;
    }
    tests.success++;
    println("Testing attribute handling, OK");
    self.nextCallbackFn(tests, err);
}


function testHttpGet(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    tests.count++;
    println("Testing httpGet()");
    cl.BaseURL = "https://feeds.library.caltech.edu";
    cl.httpGet("/recent/book.json", "application/json", function(data, err) {
        if (err !== "") {
            println("FAILED with err", err);
            tests.errors++;
            self.nextCallbackFn(tests, "");
            return;
        }
        if (data.length !== 25) {
            println("WARNINGS expected 25 books", data.length);
            tests.warnings++;
        } else {
            tests.success++;
            println("Testing httpGet(), OK");
        }
        self.nextCallbackFn(tests, "");
    });
}

function testSummary(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    println("\nFailures: " + tests.errors);
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
    testHttpGet, 
    testAttributes,
    testSummary);
}(document, window));

</script>

<!--   END: test sequence for CL-feeds.js -->
