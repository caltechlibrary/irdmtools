
# CL-ui.js test

This is a test file for putting CL-ui.js through its paces and 
to confirm we have a working `CL` object.

<div id="test-output"></div>

<style>
#status {
    font-size: 1em;
}
</style>
<code><pre id="status"></pre></code>


<!-- START: test sequence for CL-feeds.js -->
<script src="CL-core.js"></script>
<script src="CL-ui.js"></script>
<script>

(function (document, window) {
'use strict';
let cl = Object.assign({}, window.CL),
    status = document.getElementById('status');

function println(...s) {
    s.forEach(function(s) {
        console.log(s);
        status.append(s + "\n");
    });
}

/*
 * Run the following test sequences
 */
println("\nRunning tests in a pipeline\n");

function testTitleField(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    tests.count++;
    println("Testing testTitleField()");
    let field = CL.field({"title": "Hello World"}, 
        '<label>Title:</label> <input name="title" value="{{title}}" placeholder="Title goes here">',
        function(obj) {
            if ('title' in obj) {
                if (obj.title.length > 0) {
                    return true;
                }
            }
            return false;
        });
    if ('title' in field) {
        if (field.title !== "Hello World") {
            tests.errors++;
            println("FAILED, expected title of 'Hello World', got "+field.json());
        }
    } else {
        tests.errors++;
        println("FAILED: field is missing title attribute");
        self.nextCallbackFn(tests, err);
        return;
    }

    let s = field.html(),
        elem = document.getElementById('test-output');
    elem.innerHTML = elem.innerHTML + '<p><h2>testTitleField()</h2>' + s + '<hr>';

    tests.success++;
    println("Testing testTitleField() OK");
    self.nextCallbackFn(tests, err);
}

function testCreatorField(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    tests.count++;
    println("Testing testCreatorField()");
    let last_name = "Doiel",
        first_name = "Robert",
        orcid = "0000-0003-0900-6903",
        field = CL.field({
            "last_name": last_name,
            "first_name": first_name,
            "orcid": orcid
        }, 
        '<label>Last Name:</label> <input name="last_name" value="{{last_name}}"><br><label>First Name:</label> <input name="first_name" value="{{first_name}}"><br><label>ORCID: </label> <input name="orcid" value="{{orcid}}"><p>',
        function () {
            let obj = this;
            // NOTE: This initialization is validating data only.
            // You could actually translate the data to something
            // useful for rendering in your HTML, e.g. true/false
            // to "checked" attribute in a checkbox.
            if (obj.last_name === undefined ||  obj.last_name.length === 0) {
                return false;
            }
            if (obj.first_name === undefined || obj.first_name.length === 0) {
                return false;
            }
            if (obj.orcid === undefined || obj.orcid.length === 0) {
                return false;
            }
            return true;
        });

    if ('init' in field) {
        if (field.init() !== true) {
            tests.errors++;
            println('FAILED: expected init() return true, got false');
        }
    } else {
        tests.errors++;
        println('FAILED: expected init() function in field');
    }
    let val = field.get('orcid');
    if (val !== orcid) {
        tests.errors++;
        println("FAILED: expected "+orcid+", got "+val);
    }
    val = field.get('last_name');
    if (val !== last_name) {
        tests.errors++;
        println("FAILED: expected "+last_name+", got "+val);
    }
    val = field.get('first_name');
    if (val !== first_name) {
        tests.errors++;
        println("FAILED: expected "+first_name+", got "+val);
    }
    val = field.html();
    let expected = '<label>Last Name:</label> <input name="last_name" value="Doiel"><br><label>First Name:</label> <input name="first_name" value="Robert"><br><label>ORCID: </label> <input name="orcid" value="0000-0003-0900-6903"><p>';
    if (val !== expected) {
        tests.errors++;
        println("FAILED: expected\n"+expected+"\n, got\n"+val);
    }
    let o = JSON.parse(field.json());
    if (o === undefined) {
        tests.errors++;
        println("FAILED: expected an object, got undefined for field");
        self.nextCallbackFn(tests, err);
    }
    if (! 'orcid' in o) {
        tests.errors++;
        println("FAILED: expected orcid attribute, got ", o);
    }
    if (! 'last_name' in o) {
        tests.errors++;
        println("FAILED: expected last_name attribute, got ", o);
    }
    if (! 'first_name' in o) {
        tests.errors++;
        println("FAILED: expected first_name attribute, got ", o);
    }

    let s = '<p><h2>testCreatorField()</h2>' + val + '<hr>',
        elem = document.getElementById('test-output');
    elem.innerHTML = elem.innerHTML + s;

    println("Testing testCreatorField() OK");
    tests.success++;
    self.nextCallbackFn(tests, err);
}

function testCreatorList(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    tests.count++;
    println("Testing testCreatorList()");

    let items = [],
        creators = [],
        new_creator = {"last_name":"", "first_name":""},
        list_object = {};

    creators.push({
        "last_name": "Steinbeck",
        "first_name": "John"
    });
    creators.push({
        "last_name": "Verne",
        "first_name": "Jules"
    });
    creators.push({
        "last_name": "Valdez",
        "first_name": "Luis"
    });
    creators.push({
        "last_name": "Lopez",
        "first_name": "Tom"
    });

    for (let i in creators) {
        items[i] = CL.field(creators[i], '<li><span class="display_name">{{last_name}}, {{first_name}}</span></li>');
    }
    list_object = CL.field({"items": items}, "<ul>{{items}}</ul>");
    if (list_object === undefined) {
        tests.errors++;
        println("FAILED: expected list object, got undefined");
        self.nextCallbackFn(tests, err);
        return;
    }
    if (list_object.items.length !== 4) {
        tests.errors++;
        println("FAILED: expected 4, got ", list_object.json());
        self.nextCallbackFn(tests, err);
        return;
    }
    let s = list_object.html();
    if (s.includes('[object Object]')) {
        tests.errors++;
        println("FAILED: expected html to render, got ",s);
        self.nextCallbackFn(tests, err);
        return;
    }
    let elem = document.getElementById('test-output');
    elem.innerHTML = elem.innerHTML + '<p><h2>testCreatorList()</h2>' + s + '<hr>';
    println("Testing testCreatorList() OK");
    tests.success++;
    self.nextCallbackFn(tests, err);
}

function testAssembleFields(tests, err) {
    let self = this;
    if (err !== "") {
        println("FAILED: error", err, tests);
        return;
    }
    tests.count++;
    println("Testing testAssembleFields()");
    
    let book = {},
        books = [],
        steinbeck = {
            last_name: "Steinbeck",
            first_name: "John"
        },
        pratchett = {
            last_name: "Pratchett",
            first_name: "Terry"
        },
        gaiman = {
            last_name: "Gaiman",
            first_name: "Neil",
        };
     
    steinbeck = CL.field(steinbeck, 
        '<span class="last_name">{{last_name}}</span>, ' +
        '<span class="first_name">{{first_name}}</span>');
     
    book = CL.field({
            "title": "Short Reign of Pippen IV",
            "description": "A novella length satire set in post-war Paris", 
            "creators": [ steinbeck ]
        }, `
<div class="book">
    <h3 class="title">{{title}}</h3>
    <div class="creators">By {{creators}}</div>
    <div class="description">{{description}}</div>
<div>
`, undefined, '; ');
    books.push(book);
     
    pratchett = CL.field(pratchett, 
        '<span class="last_name">{{last_name}}</span>, ' +
        '<span class="first_name">{{first_name}}</span>');
     
    gaiman = CL.field(gaiman, 
        '<span class="last_name">{{last_name}}</span>, ' +
        '<span class="first_name">{{first_name}}</span>');
     
    book = CL.field({
            "title": "Good Omens",
            "description": "A book about angels and demons set in London for the most part", 
            "creators": [ pratchett, gaiman ]
        }, 
        '<div class="book">' +
        '   <h3 class="title">{{title}}</h3>' +
        '   <div class="creators">By {{creators}}</div>' + 
        '   <div class="description">{{description}}</div>' +
        '</div>',
        undefined, 
        '; ');
    books.push(book);
     
    let element = CL.assembleFields(
        document.getElementById("test-output"), ...books);
     
    println("Testing testAssembleFields() OK");
    tests.success++;
    self.nextCallbackFn(tests, err);
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
    testTitleField,
    testCreatorField,
    testCreatorList,
    testAssembleFields,
    testSummary);
}(document, window));

</script>

<!--   END: test sequence for CL-feeds.js -->
