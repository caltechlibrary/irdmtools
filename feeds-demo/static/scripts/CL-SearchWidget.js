/**
 * CL-SearchWidget.js defines the SearchWidget based on CL-core.js
 * CL-ui.js, CL-feeds.js and CL-feeds-ui.js.
 *
 * CL.SearchWidget() creates a feed search and results widget embedded in
 * an DOM element. It is a tool for generating browser side search in your 
 * webpage for Caltech Library content.
 *
 * @params element to embed the search widget.
 *
 * CL-core.js provides browser side JavaScript access to 
 * feeds.library.caltech.edu and other Caltech Library resources.
 *
 * @author R. S. Doiel
 *
Copyright (c) 2019, Caltech
All rights not granted herein are expressly reserved by Caltech.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */
/*jshint esversion: 6 */
(function(document, window) {
    "use strict";
    let CL = {};
    if (window.CL === undefined) {
        window.CL = {};
    } else {
        CL = Object.assign(CL, window.CL);
    }

    /**
     * CL.SearchWidget() creates a search widget in the elements
     * indicated by element id.
     *
     * @param parent_element_selector the DOM element selector 
     *                                which wild the widget.
     */
    CL.SearchWidget = function(parent_element) {
        let self = this,
            u = new URL(window.location.href),
            searchParams = u.searchParams;

        /* Widget code goes here */
        let widget_ui = document.createElement("div"),
            widget_error = document.createElement("div"),
            css_classname = '',
            form,
            heading,
            div,
            code,
            pre,
            section,
            label,
            input,
            select_aggregation,
            select_feed_id,
            select_feed_path,
            generate_button,
            preview_button;

        /* Widget event handlers */
        function update_feed_id(evt) {
            let aggregation = select_aggregation.value,
                param_aggregation = decodeURIComponent(searchParams.get("aggregation")),
                option,
                code_block,
                select_feed_id,
                select_feed_path;

            //NOTE: do we want to use the default value for feed_id?
            let default_value = "",
                q = searchParams.get("q");
            if (q !== null && aggregation === param_aggregation) {
                default_value = decodeURIComponent(searchParams.get("feed_id"));
            } else {
                // NOTE: If we not submitting a new search test make sure we reset preview!
                let preview = document.querySelector("#previewed-code"),
                    u = new URL(window.location.href);
                if (preview) {
                    preview.innerHTML = "";
                }
                u.search = "";
                searchParams = u.searchParams;
            }

            
            select_feed_id = document.getElementById("feed-id");
            select_feed_id.innerHTML = "";
            option = document.createElement("option");
            option.innerHTML = "Step 2. pick a feed";
            select_feed_id.appendChild(option);
            option = document.createElement("option");
            option.innerHTML = "Step 3. pick the feed type (e.g. recent/article, combined)";
            select_feed_path = document.getElementById("feed-path");
            select_feed_path.innerHTML = "";
            select_feed_path.appendChild(option);
            code_block = document.getElementById("generated-code");
            if (code_block !== undefined) {
                code_block.innerHTML = "";
            }
            if (aggregation === "people") {
                self.getPeopleList(function(people, err) {
                    if (err != "") {
                        console.log("ERROR", err);
                        return;
                    }
                    people.forEach(function(profile, i) {
                        let option = document.createElement("option");
                        option.value = profile.id;
                        if ("orcid" in profile) {
                            option.innerHTML = profile.sort_name + "(" + 
                                profile.orcid + ")";
                        } else {
                            option.innerHTML = profile.sort_name;
                        }
                        select_feed_id.appendChild(option);
                    });
                    if (default_value) {
                        select_feed_id.value = default_value;
                        update_feed_path({});
                    }
                });
            } else if (aggregation === "groups") {
                self.getGroupsList(function(groups, err) {
                    if (err != "") {
                        console.log("ERROR", err);
                        return;
                    }
                    groups.forEach(function(group, i) {
                        let option = document.createElement("option");
                        option.value = group.key;
                        option.innerHTML = group.name;
                        select_feed_id.appendChild(option);
                    });
                    if (default_value) {
                        select_feed_id.value = default_value;
                        update_feed_path({});
                    }
                });
            } else {
                generate_button.disabled = true;
                preview_button.disabled = true;
            }
        }

        function update_feed_path(evt) {
            let aggregation = select_aggregation.value,
                param_aggregation = decodeURIComponent(searchParams.get("aggregation")),
                feed_id = select_feed_id.value,
                param_feed_id = decodeURIComponent(searchParams.get("feed_id")),
                option,
                code_block,
                select_feed_path;

            //NOTE: do we want to use the default value for feed_id, feed_path?
            let default_value = "",
                q = searchParams.get("q");
            if (q !== null && aggregation === param_aggregation && feed_id === param_feed_id) {
                default_value = decodeURIComponent(searchParams.get("feed_path"));
            } else {
                // NOTE: If we not submitting a new search test make sure we reset preview!
                let preview = document.querySelector("#previewed-code"),
                    u = new URL(window.location.href);
                if (preview) {
                    preview.innerHTML = "";
                }
                u.search = "";
                searchParams = u.searchParams;
            }

            option = document.createElement("option");
            option.innerHTML = "Step 3. Pick feed type";
            select_feed_path = document.getElementById("feed-path");
            select_feed_path.innerHTML = "";
            select_feed_path.appendChild(option);
            code_block = document.getElementById("generated-code");
            if (code_block !== undefined) {
                code_block.innerHTML = "";
            }

            if (aggregation === "people") {
                self.getPeopleInfo(feed_id, function(profile, err) {
                    if (err != "") {
                        console.log("ERROR", err);
                        return;
                    }
                    if ("CaltechTHESIS" in profile) {
                        for (let feed_label in profile.CaltechTHESIS) {
                            let option = document.createElement("option");
                            //NOTE: People don't have combined thesis, only
                            // Groups.
                            if (feed_label !== "combined") {
                                option.innerHTML = "CaltechTHESIS: " + feed_label;
                                option.value = feed_label.toLocaleLowerCase().replace(/ /g, "_") + ":caltechthesis";
                                select_feed_path.appendChild(option);
                            }
                        }
                    }
                    if ("CaltechAUTHORS" in profile) {
                        for (let feed_label in profile.CaltechAUTHORS) {
                            let option = document.createElement("option");
                            option.innerHTML = "CaltechAUTHORS: " + feed_label;
                            option.value = feed_label.toLocaleLowerCase().replace(/ /g, "_") + ":caltechauthors";
                            select_feed_path.appendChild(option);
                        }
                    }
                    if ("CaltechDATA" in profile) {
                        for (let feed_label in profile.CaltechDATA) {
                            let option = document.createElement("option"),
                                feed_type = feed_label.toLocaleLowerCase().replace(/ /g, "_");
                            if (feed_type === "combined") {
                                feed_type = "data";
                            } else if (feed_type === "interactive_resource") {
                                feed_type = "interactiveresource";
                            }
                            option.innerHTML = "CaltechDATA: " + feed_label;
                            option.value = feed_type + ":caltechdata";
                            select_feed_path.appendChild(option);
                        }
                    }
                    if (default_value) {
                        select_feed_path.value = default_value;
                        preview_code({});
                    }
                });
            } else if (aggregation === "groups") {
                self.getGroupInfo(feed_id, function(group, err) {
                    if (err != "") {
                        console.log("ERROR", err);
                        return;
                    }
                    if ("CaltechTHESIS" in group) {
                        for (let feed_label in group.CaltechTHESIS) {
                            let option = document.createElement("option");
                            option.innerHTML = "CaltechTHESIS: " + feed_label;
                            option.value = feed_label.toLocaleLowerCase().replace(/ /g, "_") + ":caltechthesis";
                            select_feed_path.appendChild(option);
                        }
                    }
                    if ("CaltechAUTHORS" in group) {
                        for (let feed_label in group.CaltechAUTHORS) {
                            let option = document.createElement("option");
                            option.innerHTML = "CaltechAUTHORS: " + feed_label;
                            option.value = feed_label.toLocaleLowerCase().replace(/ /g, "_") + ":caltechauthors";
                            select_feed_path.appendChild(option);
                        }
                    }
                    if ("CaltechDATA" in group) {
                        for (let feed_label in group.CaltechDATA) {
                            let option = document.createElement("option"),
                                feed_type = feed_label.toLocaleLowerCase().replace(/ /g, "_");
                            if (feed_type === "combined") {
                                feed_type = "data";
                            } else if (feed_type === "interactive_resource") {
                                feed_type = "interactiveresource";
                            }
                            option.innerHTML = "CaltechDATA: " + feed_label;
                            option.value = feed_type + ":caltechauthors";
                            select_feed_path.appendChild(option);
                        }
                    }
                    if (default_value) {
                        select_feed_path.value = default_value;
                        preview_code({});
                    }
                });
            }
        }
        
        function select_feed(evt) {
            let aggregation = select_aggregation.value,
                param_aggregation = decodeURIComponent(searchParams.get("aggregation")),
                feed_id = select_feed_id.value,
                param_feed_id = decodeURIComponent(searchParams.get("feed_id")),
                feed_path = select_feed_path.value,
                param_feed_path = decodeURIComponent(searchParams.get("feed_path")),
                code_block;

            // NOTE: If we not submitting a new search test make sure we reset preview!
            let preview = document.querySelector("#previewed-code"),
                u = new URL(window.location.href);
            if (preview) {
                preview.innerHTML = "";
                u.search = "";
                searchParams = u.searchParams;
                //update_feed_path(evt);
            }

            code_block = document.getElementById("generated-code");
            if (code_block !== undefined) {
                code_block.innerHTML = "";
            }
            if (select_feed_path.value.startsWith("Step ")) {
                generate_button.disabled = true;
                preview_button.disabled = true;
            } else {
                generate_button.disabled = false;
                preview_button.disabled = false;
                let parts = select_feed_path.value.split(":");
                if (parts.length === 2) {
                    switch (parts[1]) {
                        case 'caltechauthors':
                            css_classname = ".CaltechAUTHORS";
                            break;
                        case 'caltechthesis':
                            css_classname = ".CaltechTHESIS";
                            break;
                        case 'caltechdata':
                            css_classname = ".CaltechDATA";
                            break;
                        default:
                            css_classname = ".CaltechLibrary";
                            break;
                    }
                }
            }
        }


        // get_config scans the settings in the Search Widget form and creates a configuration to
        // suitable to pass to code_render().
        function get_config() {
            let config = {};

            ["aggregation", "feed-id", "feed-path"].forEach(function(id) {
                let elem = document.getElementById(id),
                    key = "",
                    val = "";
                if (elem !== undefined) {
                    key = id.replace(/-/g, "_");
                    config[key] = elem.value;
                }
            });
            ["feed-count", "creators", "pub-date", "title-link", "citation-details", "issn-or-isbn", "pmcid", "description"].forEach(function(id) {
                let elem = document.getElementById(id),
                    key;
                key = id.replace(/-/g, "_");
                if (elem.checked === true) {
                    config[key] = true;
                } else {
                    config[key] = false;
                }
            });
            return config;
        }

        // code_render take the contents of the form and render the 
        // resulting source code.
        function code_render(config) {
            let text = [],
                include_style = true,
                include_CL = true,
                developer_mode = false,
                elem_id = "cl",
                query_form;

            if (config.include_style !== undefined) {
                include_style = config.include_style;
            }
            if (config.include_CL !== undefined) {
                include_CL = config.include_CL;
            }
            if (config.developer_mode !== undefined) {
                developer_mode = config.developer_mode;
            }


            config.repository = css_classname.substr(1);
            config.css_classname = css_classname;
            if (config.feed_id !== undefined && config.feed_id !== "") {
                elem_id = config.feed_id;
            }
            if (config.feed_path !== undefined && config.feed_path !== "") {
                config.feed_path = config.feed_path.split(":")[0];
            }
            if (config.use_recent === undefined ||
                config.use_recent === false) {
                config.recent_n = 0;
            }
            // Generate Style Block and HTML block
            if (include_style === true) {
                text.push("<style>");
                text.push(css_classname + " .unknown-year { display: none; }");
                if (config.title_link === false) {
                    text.push(css_classname + " .title { padding-left: 0.24em }");
                    text.push(css_classname + " .link { padding-left: 0.24em }");
                }
                if (config.show_year_headings === true) {
                    text.push(css_classname + " .jump-list {");
                    text.push("    padding-bottom: 0.24em;");
                    text.push("    margin-bottom: 0.24em;");
                    text.push("    border-bottom: solid 0.24em black;");
                    text.push("}");
                    text.push(css_classname + " .jump-list-label {");
                    text.push("    padding-left:0.24em;");
                    text.push("    padding-right:0.24em;");
                    text.push("    border-right: solid 0.12em black;");
                    text.push("    text-decoration: none;");
                    text.push("}");
                    text.push(css_classname + " .jump-list:last-child {");
                    text.push("    border-right: none;");
                    text.push("}");
                }
                text.push(css_classname + " li {");
                text.push("    padding-bottom: 0.24em;");
                text.push("    margin-bottom: 0.24em;");
                text.push("    list-style: none;");
                text.push("}");
                text.push(css_classname + " a {");
                text.push("    padding-right: 0.24em;");
                text.push("}");
                text.push(css_classname + " span {");
                text.push("    padding-right: 0.24em;");
                text.push("}");
                text.push(css_classname + " div {");
                text.push("    padding-bottom: 0.24em;");
                text.push("    margin-bottom: 0.24em;");
                text.push("}");
                text.push("</style>\n");
            }

            text.push("<div id=\"" + elem_id + "\" class=\"" + css_classname.substr(1) + "\"></div>\n");


            // Generate JavaScript CL.js include 
            if (include_CL == true) {
                if (developer_mode === true) {
                    text.push("<script src=\"https://unpkg.com/lunr/lunr.js\"></script>")
                    text.push("<script src=\"/scripts/CL-core.js\"></script>");
                    text.push("<script src=\"/scripts/CL-ui.js\"></script>");
                    text.push("<script src=\"/scripts/CL-feeds.js\"></script>");
                    text.push("<script src=\"/scripts/CL-feeds-ui.js\"></script>");
                    text.push("<script src=\"https://feeds.library.caltech.edu/scripts/CL-feeds-lunr.js\"></script>");
                } else {
                    text.push("<script src=\"https://unpkg.com/lunr/lunr.js\"></script>")
                    text.push("<script src=\"https://feeds.library.caltech.edu/scripts/CL.js\"></script>");
                    text.push("<script src=\"https://feeds.library.caltech.edu/scripts/CL-feeds-lunr.js\"></script>");
                }
            }

            // Generate JavaScript src block
            config.filters = [];
            text.push("<script>");
            text.push("(function(document, window) {");
            text.push("  \"use strict\";");
            text.push("  let cl = Object.assign({}, window.CL),");
            text.push("      config = {},");
            text.push("      q = '',");
            text.push("      u = new URL(window.location.href),");
            text.push("      elem = document.getElementById(\"" +
                elem_id + "\"),");

            text.push("      query_form = document.createElement('div');");
            text.push("");

            text.push("  if (u.search !== '') {");
            text.push("      q = u.searchParams.get('q');"); 
            text.push("      if (q === null) {");
            text.push("          q = ''; ");
            text.push("      } else {"); 
            text.push("          q = decodeURIComponent(q);");
            text.push("      }");
            text.push("  }");
            
            query_form = `
<form method="get">
  <input type="text" name="q" value="` + "${q}" + `" placeholder="Enter search terms">
  <button>Search</button>
</form>
`;
            text.push("  query_form.innerHTML = `" + query_form + "`;");
            text.push("  elem.appendChild(query_form);");
            if (developer_mode === true) {
                text.push("");
                text.push("    /* NOTE: Remove the following when we're ready");
                text.push("       for production. */");
                text.push("    cl.BaseURL = \"\";");
                text.push("");
            } else {
                text.push("    cl.BaseURL = \"https://feeds.library.caltech.edu\";");
            }
            text.push("");
            text.push("  config = " +
                JSON.stringify(config, "", "    ") + ";");
            text.push("  config.show_search_box = true;");
            text.push("  config.parent_element = elem;");

            text.push("  config.filters.push(cl.normalize_view);");
            /* NOTE: lunr_search includes indexing if a query is submitted, otherwise
               all records are returned unfilter by lunr_search */
            text.push("  config.filters.push(cl.lunr_search);");
            
           
            text.push("  cl.setAttribute(\"viewer\", config);");
            switch (config.aggregation) {
                case "groups":
                    text.push("  cl.getGroupJSON(\"" + config.feed_id + "\", \"" + config.feed_path + "\", function(data, err) {");
                    break;
                case "people":
                    text.push("  cl.getPeopleJSON(\"" + config.feed_id + "\", \"" + config.feed_path + "\", function(data, err) {");
                    break;
            }
            text.push("     cl.viewer(data, err);");
            text.push("  });");
            text.push("}(document, window));");
            text.push("</script>");
            // Generate JavaScript code block 
            return text.join("\n");
        }

        function generate_code(evt) {
            let config = get_config(),
                code_block = document.getElementById("generated-code"),
                preview_block = document.getElementById("previewed-code");
            if (code_block !== undefined && code_block.innerHTML !== "") {
                code_block.innerHTML = "";
            }
            if (preview_block !== undefined && preview_block.innerHTML !== "") {
                preview_block.innerHTML = "";
            }
            config.developer_mode = false;
            config.include_style = true;
            config.include_CL = true;
            code_block.textContent = code_render(config);
        }

        function preview_code(evt) {
            let src = "",
                input = document.createElement("input"),
                config = get_config(),
                code_block = document.getElementById("generated-code"),
                preview_block = document.getElementById("previewed-code");
            if (code_block !== undefined && code_block.innerHTML !== "") {
                code_block.innerHTML = "";
            }
            if (preview_block !== undefined && preview_block.innerHTML !== "") {
                preview_block.innerHTML = "";
            }
            config.developer_mode = true;
            config.include_style = true;
            config.include_CL = false;
            src = code_render(config);
            let div = document.createElement("div"),
                style,
                js_src = "",
                form;
            div.innerHTML = src;
            js_src = div.querySelector("script").textContent;
            style = div.querySelector("style");
            /* NOTE: we only want to the div we were going to 
             * render into */
            preview_block.appendChild(style);
            preview_block.appendChild(div.querySelector("div"));
            /* UGLY: doing this eval so I can preview what the JS 
             * I generarted renders */
            eval(js_src);
            /** NOTE: we need to modify the input for so if they press search we can
                recreate everything with these settings including running the search results **/
            form = preview_block.querySelector("form");
            if (form !== null) {
                let q = searchParams.get("q");
                if (q !== null) {
                    let searchbox = form.querySelector("[name=q]")
                    searchbox.setAttribute("value", q);
                }
                input.setAttribute("name", "aggregation");
                input.setAttribute("type", "hidden");
                input.setAttribute("value", document.querySelector("#aggregation").value);
                form.appendChild(input);

                input = document.createElement("input");
                input.setAttribute("name", "feed_id");
                input.setAttribute("type", "hidden");
                input.setAttribute("value", document.querySelector("#feed-id").value);
                form.appendChild(input);

                input = document.createElement("input");
                input.setAttribute("name", "feed_path");
                input.setAttribute("type", "hidden");
                input.setAttribute("value", document.querySelector("#feed-path").value);
                form.appendChild(input);

                ["feed-count", "creators", "pub-date", "title-link", "citation-details", "issn-or-isbn", "pmcid", "description"].forEach(function(id) {
                    let name = id.replace(/-/g,"_");
                    input = document.createElement("input");
                    input.setAttribute("name", name);
                    input.setAttribute("type", "hidden");
                    if (document.querySelector("#"+id).checked === true) {
                        input.setAttribute("value", 1);
                    } else {
                        input.setAttribute("value", 0);
                    }
                    form.appendChild(input);
                });
            }
        }


        /*
         * Main Search Widget UI setup
         */

        /* Form holds our control panel for generating code */
        form = document.createElement("form");
        form.setAttribute("id", "feed-search-widget");
        

        heading = document.createElement("h1");
        heading.innerHTML = "Search Widget";
        form.appendChild(heading);
        heading = document.createElement("h2");
        heading.innerHTML = "Data Source";
        form.appendChild(heading);

        /* Step 1. Pick which aggregation you want to generate code for */
        div = self.createCompositElement("div", ["label", "select"], ["", "aggregation"]);
        label = div.querySelector("label");
        label.setAttribute("for", "aggregation");
        label.setAttribute("title", "Step 1. pick an aggregation (people or groups)");
        label.innerHTML = "Aggregation:";
        select_aggregation = div.querySelector("#aggregation");
        select_aggregation.setAttribute("name", "aggregation");
        ["", "Groups", "People"].forEach(function(value, i) {
            let option = document.createElement("option");
            if (i === 0) {
                option.setAttribute("value", "");
                option.setAttribute("title", "clear selection");
                option.innerHTML = "Step 1. pick an aggregation";
            } else {
                option.setAttribute("value", value.toLocaleLowerCase());
                option.innerHTML = value;
            }
            select_aggregation.appendChild(option);
        });
        select_aggregation.addEventListener("change", update_feed_id, false);
        form.appendChild(div);

        /* Step 2. Pick a feed (e.g. GALCIT, Newman-D-K) */
        div = self.createCompositElement("div", ["label", "select"], ["", "feed-id"]);
        label = div.querySelector("label");
        label.setAttribute("for", "feed-id");
        label.setAttribute("title", "Step 2. pick the feed id");
        label.innerHTML = "Feed:";
        select_feed_id = div.querySelector("#feed-id");
        select_feed_id.setAttribute("name", "feed-id");
        select_feed_id.setAttribute("title", "this list depends on the aggregation previously selected");
        select_feed_id.addEventListener("change", update_feed_path, false);
        form.appendChild(div);

        /* Step 3. Pick a feed type (e.g. article, recent/article, combined) */
        div = self.createCompositElement("div", ["label", "select"], ["", "feed-path"]);
        label = div.querySelector("label");
        label.setAttribute("for", "feed-path");
        label.setAttribute("title", "Step 3. pick the feed type (e.g. article, combined)");
        label.innerHTML = "Feed type:";
        select_feed_path = div.querySelector("#feed-path");
        select_feed_path.setAttribute("name", "feed-path");
        select_feed_path.setAttribute("title", "list of available feed paths");
        select_feed_path.addEventListener("change", select_feed, false);
        form.appendChild(div);

        // NOTE: we don't need the Filter Data section in the UI
        // this search box is the filter.

        heading = document.createElement("h2");
        heading.innerHTML = "Layout";
        heading.setAttribute("title", "Step 4. pick the fields to display");
        form.appendChild(heading);

        /* Step 4. Pick listing layout format */
        div = document.createElement("div");
        div.classList.add("checkbox-control");

        ["Feed Count", "Creators", "Pub Date", "Title Link", "Citation Details", "ISSN or ISBN", "PMCID", "Description"].forEach(function(s, i) {
            let elem_id = s.toLocaleLowerCase().replace(/ /g, "-"),
                elem_name = s.toLocaleLowerCase().replace(/ /g, "_"),
                control, label, input;

            control = self.createCompositElement("div", ["label", "input"], ["", elem_id]);
            control.classList.add("checkbox");
            input = control.querySelector("#" + elem_id);
            input.setAttribute("type", "checkbox");
            input.setAttribute("name", elem_name);
            input.setAttribute("label", s);
            if ([0, 2, 3, 4].indexOf(i) > -1) {
                input.setAttribute("checked", true);
            }
            label = control.querySelector("label");
            label.innerHTML = s + ":";
            div.append(control);
        });
        form.appendChild(div);

        /* Step 5. preview and generate the code */
        div = self.createCompositElement("div", ["input", "input"], ["preview", "generate"]);

        /* setup generate code and preview code buttons */
        input = div.querySelector("#preview");
        input.disabled = true;
        input.setAttribute("id", "preview-code");
        input.setAttribute("type", "button");
        input.setAttribute("value", "Preview Searchbox and Results");
        input.addEventListener("click", preview_code, false);
        preview_button = input;

        input = div.querySelector("#generate");
        /* NOTE: this input should become enabled 
         * when the data sources
         * have been defined. */
        input.disabled = true;
        input.setAttribute("id", "generate-code");
        input.setAttribute("type", "button");
        input.setAttribute("value", "Generate code");
        input.addEventListener("click", generate_code, false);
        generate_button = input;

        form.appendChild(div);

        /* Instantiate the form! */
        widget_ui.appendChild(form);
        /* Add <code><pre> bocks for generated output */
        code = document.createElement("code");
        pre = document.createElement("pre");
        pre.setAttribute("id", "generated-code");
        code.appendChild(pre);
        /* Add section to preview generated output of code */
        section = document.createElement("section");
        section.classList.add("preview");
        section.setAttribute("id", "previewed-code");

        widget_ui.setAttribute("id", "search-widget-ui");
        widget_error.setAttribute("id", "search-widget-error");
        parent_element.appendChild(widget_ui);
        parent_element.appendChild(widget_error);
        parent_element.appendChild(code);
        parent_element.appendChild(section);        
        
        /** Trigger form updates as needed **/
        (function() {
            let q = searchParams.get("q"),
                aggregation = document.querySelector("#aggregation"),
                params_aggregation = searchParams.get("aggregation"),
                feed_id = document.querySelector("#feed-id"),
                params_feed_id = searchParams.get("feed_id"),
                feed_path = document.querySelector("#feed-path"),
                params_feed_path = searchParams.get("feed_path");
              
            if (q !== null) {
                if (aggregation !== null && aggregation.value === "" && params_aggregation &&
                    feed_id !== null && feed_id.value === "" && params_feed_id &&
                    feed_path !== null && feed_path.value === "" && params_feed_path) {
                    ["feed-count", "creators", "pub-date", "title-link", "citation-details", "issn-or-isbn", "pmcid", "description"].forEach(function(id) {
                        let name = id.replace(/-/g, '_'),
                            elem = document.querySelector('#'+id),
                            default_value = searchParams.get(name);
                        if (default_value == 1) {
                            elem.setAttribute("checked", "checked");
                        } else {
                            elem.removeAttribute("checked");
                        }
                    });
                    aggregation.value = params_aggregation;
                    generate_button.disabled = false;
                    preview_button.disabled = false;
                    update_feed_id({});
                }
            }
        }());
    };

    /* Now add CL.SearchWidget to the CL in the window object. */
    if (window.CL === undefined) {
        window.CL = {};
    }
    window.CL = Object.assign(window.CL, CL);
}(document, window));
