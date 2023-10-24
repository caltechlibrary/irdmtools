
# CL-doi-media.js test

This is a test file for putting CL-doi-media.js through its paces and 
to confirm we have a working `CL` object.


<div id="video-demo"></div>


<style>
#status {
    font-size: 1em;
}
</style>
<code><pre id="status"></pre></code>

<!-- START: test sequence for CL-feeds.js -->

<script src="CL-core.js"></script>
<script src="CL-ui.js"></script>
<script src="CL-doi-media.js"></script>
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
function testDoiMedia(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    tests.count++;
    println("Testing doi media handling");
    let doi = '10.22002/D1.1281',
        item_no = 0,
        elem = document.createElement("div"),
        expected_url = 'https://data.caltech.edu/tindfiles/serve/feff786e-1123-4e7b-ac8e-29a365d6bc9f/',
        expected_type = 'video/mp4';
    CL.doi_media(doi, item_no, function(obj, err) {
        if (err) {
            tests.errors++;
            println(`Go error from doi_media("${doi}", "${item_no}", fnRenderCallback) `, err);
            self.nextCallbackFn(tests, err);
            return;
        }
        if (obj.media_url !== expected_url) {
            tests.errors++;
            println(`expected ${expected_url}, got ${obj.media_url}`)
            self.nextCallbackFn(tests, err);
            return;
        }
        if (obj.media_type !== expected_type) {
            tests.errors++;
            println(`expected ${expected_url}, got ${obj.media_url}`)
            self.nextCallbackFn(tests, err);
            return;
        }
        tests.success++;
        println("Testing doi media handling, OK");
        self.nextCallbackFn(tests, err);
    });
}

function testDoiVideoPlayer(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        self.nextCallbackFn(tests, err);
        return;
    }
    tests.count++;
    println("Testing Doi Video Player");
    let div = document.getElementById("video-demo"),
        doi = '10.22002/D1.1281',
        item_no = 0;
    CL.doi_video_player(div, doi, item_no, 1024, 768);
    setTimeout(function() {
        let src = div.textContent;
        if (src.includes('error')) {
            tests.error++;
            println("FAILED to create video player");
            self.nextCallbackFn(tests, err);
            return;
        }
        tests.success++;
        println("Testing Doi Video Player, OK");
        self.nextCallbackFn(tests, err);
        return;
    }, 2000);
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
    testDoiMedia,
    testDoiVideoPlayer,
    testSummary);
}(document, window));

</script>

<!--   END: test sequence for CL-feeds.js -->
