/**
 * CL-feeds-ui.js builds on the CL object providing simple support
 * for constructing DOM based UI.
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
        CL = Object.assign({}, window.CL);
    }

    /** 
     * CL.createCompositElement() takes an element type (e.g.
     * div, span, h1, a) and append children based on an array
     * of type information, ids and CSS classes.
     *
     * @param element_type (string, required) string
     * @param child_element_types (array of string, required)
     * @param child_element_ids (array of string, optional)
     * @param child_element_classes (array of string, optional)
     * @return DOM element containing children
     */
    CL.createCompositElement = function(element_type, child_element_types, child_element_ids = [], child_element_classes = []) {
        let outer = document.createElement(element_type);
        child_element_types.forEach(function(child_element_type, i) {
            let inner = document.createElement(child_element_type);
            if (i < child_element_ids.length && child_element_ids[i] !== "") {
                inner.setAttribute("id", child_element_ids[i]);
            }
            if (i < child_element_classes.length && child_element_classes[i] !== "") {
                let css_classes = [];
                if (child_element_classes[i].indexOf(" ") > -1) {
                    css_classes = child_element_classes[i].split(" ");
                } else {
                    css.classes.push(child_element_classes[i]);
                }
                css_classes.forEach(function(css_class) {
                    inner.classList.add(css_class);
                });
            }
            outer.appendChild(inner);
        });
        return outer;
    };


    /* isEPrintsRecord is an interal function, not exported. */
    function isEPrintsRecord(record) {
        if (record.collection !== undefined &&
            (record.collection === "CaltechAUTHORS" ||
                record.collection === "CaltechTHESIS")) {
            return true;
        }
        if (record.id !== undefined && typeof record.id === "string" &&
            record.id.indexOf("eprint") > -1) {
            return true;
        }
        return false;
    }

    /**
     * normalize_view is a function to use with a CL.pipeline. It expects
     * data and error parameters and will envoke `this.nextCallbackFn(data, err)`
     * before existing. The purpose of normalize_view is to extract titles, links,
     * pub_date, creator and description from both Invenion and EPrints style JSON 
     * lists.
     */
    CL.normalize_view = function(data, err) {
        let self = this,
            normal_view = [];

        for (let i in data) {
            let record = data[i],
                view = {
                    "href": "",
                    "title": "",
                    "creators": [],
                    "description": "",
                    "pub_date": "",
                    "collection": "",
                    "doi": "",
                    "primary_object": {},
                    /* Citation Info to be removed per DR-273, each shold be individually selectable */
                    "resource_type": ""
                };
            /* Normalize our view between EPrint and Invenio style records */
            /*NOTE: maybe creating the view should be its own filter function? */
            if (isEPrintsRecord(record) === true) {
                view.collection = record.collection;
                view.title = record.title;
                if (record.type !== undefined && record.type !== "") {
                    view.resource_type = record.type;
                }
                if (record.book_title !== undefined && record.book_title !== "") {
                    view.book_title = record.book_title;
                }
                /* NOTE: DR-313 we no longer prefer the DOI for link */
                view.href = record.official_url;
                if (record.doi !== undefined && record.doi !== "") {
                    view.doi = record.doi;
                }
                /* NOTE: We're adding a primary_object link if available. */
                if (record.primary_object !== undefined && record.primary_object.mime_type !== undefined && record.primary_object.url) {
                    let label = record.primary_object.mime_type;
                    if (label.indexOf('/') > -1) {
                        label = label.split('/', 2)[1].toUpperCase();
                    }
                    view.primary_object.url = record.primary_object.url;
                    view.primary_object.label = label;
                }
                if (record.volume !== undefined && record.volume !== "") {
                    view.volume = record.volume;
                }
                /* number is issue number in Journals, Magazines */
                if (record.number !== undefined && record.number !== "") {
                    view.number = record.number;
                }
                if (record.series !== undefined && record.series !== "") {
                    view.series = record.series;
                }
                /* Removed per DR-273, pages is populated by arxiv imported and in spurious */
                if (record.pagerange !== undefined && record.pagerange !== "") {
                    view.page_range = record.pagerange;
                }
                if (record.publisher !== undefined && record.publisher !== "") {
                    view.publisher = record.publisher;
                }
                if (record.publication !== undefined && record.publication !== "") {
                    view.publication = record.publication;
                }
                if (record.issn !== undefined && record.issn !== "") {
                    view.issn = record.issn;
                }
                if (record.isbn !== undefined && record.isbn !== "") {
                    view.isbn = record.isbn;
                }
                if (record.edition !== undefined && record.edition !== "") {
                    view.edition = record.edition;
                }
                if (record.event_title !== undefined && record.event_title !== "") {
                    view.event_title = record.event_title;
                }
                if (record.event_dates !== undefined && record.event_dates !== "") {
                    view.event_dates = record.event_dates;
                }
                if (record.event_location !== undefined && record.event_location !== "") {
                    view.event_location = record.event_location;
                }
                if (record.series !== undefined && record.series !== "") {
                    view.series = record.series;
                }
                if (record.ispublished !== undefined && record.ispublished !== "") {
                    if (record.ispublished === "inpress") {
                        view.ispublished = "(In Press)";
                    }
                    if (record.ispublished === "submitted") {
                        view.ispublished = "(Submitted)";
                    }
                }
                //FIXME: we had this field wrongly labeled in our EPrint 
                // output, it was called .pmc_id rather than .pmcid we
                // can simplify this if/else when that has propagated
                // throughout our collections.
                if (record.pmcid !== undefined && record.pmcid !== "") {
                    view.pmcid = record.pmcid;
                } else if (record.pmc_id !== undefined && record.pmc_id !== "") {
                    view.pmcid = record.pmc_id;
                }
                // NOTE: Some records have no publication date because 
                // there is no date in the material provided
                // when it was digitized and added to the repository.
                view.pub_date = '';
                if (record.date_type !== undefined &&
                    (record.date_type === 'completed' ||
                        record.date_type === 'published' ||
                        record.date_type === 'inpress' ||
                        record.date_type === 'submitted' ||
                        record.date_type === 'degree')) {
                    view.pub_date = '(' + record.date.substring(0, 4) + ')';
                } else if (record.type !== undefined && record.date !== undefined &
                    (record.type === 'conference_item' || record.type === 'teaching_resource') && record.date !== '') {
                    view.pub_date = '(' + record.date.substring(0, 4) + ')';
                }
                if (record.creators !== undefined && record.creators.items !== undefined) {
                    view.creators = [];
                    record.creators.items.forEach(function(creator, i) {
                        let display_name = "",
                            orcid = "",
                            id = "";

                        if (creator.name.given !== undefined && creator.name.family !== undefined) {
                            display_name = creator.name.family + ", " + creator.name.given;
                        } else if (creator.name.family !== undefined) {
                            display_name = creator.name.family;
                        }
                        if (creator.id !== undefined) {
                            id = creator.id;
                        }
                        if (creator.orcid !== undefined) {
                            orcid = creator.orcid;
                        }
                        view.creators.push({
                            "id": id,
                            "display_name": display_name,
                            "orcid": orcid,
                            "pos": i
                        });
                    });
                    /* DR-135, add additional fields for conference items. */
                    if (record.type !== undefined && record.type === 'conference_item') {
                        view.event_title = record.event_title;
                        view.event_dates = record.event_dates;
                        view.event_location = record.event_location;
                    }
                }
                view.description = record.abstract;
            } else {
                view.collection = "CaltechDATA";
                view.title = record.titles[0].title;
                if (record.resourceType !== undefined && record.resourceType.resourceTypeGeneral !== undefined && record.resourceType.resourceTypeGeneral !== "") {
                    view.resource_type = record.resourceType.resourceTypeGeneral;
                }
                view.pub_date = record.publicationYear;
                if (record.creators !== undefined) {
                    view.creators = [];
                    record.creators.forEach(function(creator, i) {
                        let display_name = "",
                            orcid = "";

                        if (creator.creatorName !== undefined) {
                            display_name = creator.creatorName;
                        }
                        if (creator.nameIdenitifiers !== undefined) {
                            creator.nameIdentifiers.forEach(function(identifier) {
                                if (identifier.nameIdentifierScheme === "ORCID") {
                                    orcid = identifier.nameIdentifier;
                                }
                            });
                        }
                        view.creators.push({
                            "display_name": display_name,
                            "orcid": orcid,
                            "pos": i
                        });
                    });
                }
                view.description = record.descriptions.join("<p>");
            }
            normal_view.push(view);
        }
        self.nextCallbackFn(normal_view, "");
    };

    /**
     * recentN is a function to use with CL.pipeline. It expects data and error parameters and will
     * envoke `this.nextCallbackFn(data, error)` before exiting.
     *
     * @param data (a JS data type, required) this is usually a list to iterate filter for N items.
     * @param err (string, required) is an error string which is empty of no errors present.
     */
    CL.recentN = function(data, err) {
        let self = this,
            N = 0;

        if (err !== "") {
            self.nextCallbackFn(data, err);
            return;
        }
        N = self.getAttribute("recentN");
        if (N === undefined || Number.isInteger(N) === false || N < 1) {
            self.nextCallbackFn(data, "recentN attribute not set properly, an integer greater than zero required");
            return;
        }
        if (Array.isArray(data) === true) {
            self.nextCallbackFn(data.slice(0, N), err);
        }
        self.nextCallbackFn(data, "data was not an array, can't take N of them");
    };

    /**
     * titleCase is a naive title case function. Splits in spaces,
     * capitalizes each first let, lower casing the rest of the string and
     * finally joins the array of strings with spaces.
     */
    function titleCase(s) {
        return s.split(" ").map(function(word) {
            if (word.endsWith(".")) {
                return word;
            }
            if (word in ["of", "the", "a", "and", "or"]) {
                return word.toLowerCase();
            }
            return word[0].toUpperCase() + word.substr(1).toLowerCase();
        }).join(" ");
    }


    /**
     * viewer is a callback suitible to be used by functions like getPeopleJSON() and getGroupJSON().
     * it takes a configuration from the attribute "viewer" and will apply a filter pipeline if provided.
     * If no configuration provided then viewer will display unlinked titles.
     *
     * @param data (object, required) the data received from the calling function
     * @param err (string, required) holds any existing error message passed to it by calling function.
     */
    CL.viewer = function(data, err) {
        let self = this,
            filters = [],
            show_feed_count = false,
            show_year_headings = false,
            show_creators = false,
            show_pub_date = false,
            show_title_linked = false,
            show_link = false,
            show_issn = false,
            show_isbn = false,
            show_pmcid = false,
            show_doi = false,
            show_primary_object = false,
            show_publisher = false,
            show_publication = false,
            show_page_numbers = false,
            show_chapters = false,
            show_volume = false,
            show_issue = false,
            show_description = false,
            show_search_box = false,
            config = {},
            parent_element,
            __display;
        config = self.getAttribute("viewer");
        /* To be cautious we want to validate our configuration object */
        if (config.show_search_box !== undefined && config.show_search_box === true) {
            show_search_box = true;
        }
        if (config.filters !== undefined && Array.isArray(config.filters)) {
            filters = config.filters;
        }
        if (config.feed_count !== undefined && config.feed_count === true) {
            show_feed_count = true;
        }
        if (config.show_year_headings !== undefined && config.show_year_headings === true) {
            show_year_headings = true;
        }
        if (config.creators !== undefined && config.creators === true) {
            show_creators = true;
        }
        if (config.pub_date !== undefined && config.pub_date === true) {
            show_pub_date = true;
        }
        if (config.title_link !== undefined && config.title_link === true) {
            show_title_linked = true;
        }
        if (config.link !== undefined && config.link === true) {
            show_link = true;
        }
        if (config.publisher !== undefined && config.publisher === true) {
            show_publisher = true;
        }
        if (config.publication !== undefined && config.publication === true) {
            show_publication = true;
        }
        if (config.page_numbers !== undefined && config.page_numbers === true) {
            show_page_numbers = true;
        }
        if (config.chapters !== undefined && config.chapters === true) {
            show_chapters = true;
        }
        if (config.issue !== undefined && config.issue === true) {
            show_issue = true;
        }
        if (config.volume !== undefined && config.volume === true) {
            show_volume = true;
        }
        if (config.issn_or_isbn !== undefined && config.issn_or_isbn === true) {
            show_issn = true;
            show_isbn = true;
        }
        if (config.pmcid !== undefined && config.pmcid === true) {
            show_pmcid = true;
        }
        if (config.doi !== undefined && config.doi === true) {
            show_doi = true;
        }
        if (config.primary_object !== undefined && config.primary_object === true) {
            show_primary_object = true;
        }
        if (config.description !== undefined && config.description === true) {
            show_description = true;
        }
        if (config.parent_element !== undefined && config.parent_element) {
            parent_element = config.parent_element;
        } else if (self.element !== undefined) {
            parent_element = self.element;
        } else {
            /* Worst case append a section element to body with a class CL-Library-Feed */
            let body = document.querySelector("body");
            parent_element = document.createElement("section");
            parent_element.classList.add("CL-library-Feed");
            body.appendChild(parent_element);
        }


        __display = function(records, err) {
            if (err != "") {
                parent_element.classList.addClass("error");
                parent_element.innerHTML = err;
                return;
            }
            let ul = document.createElement("ul"),
                feed_count = document.createElement("div"),
                year_jump_list = document.createElement("div"),
                year_heading = "";
            /* Clear the inner content of our element. */
            parent_element.innerHTML = "";
            /* Handle Managing Year Jump List */
            if (show_year_headings === true) {
                year_heading = "";
                parent_element.append(year_jump_list);
            }
            /* Handle feed count */
            if (show_feed_count === true) {
                feed_count.innerHTML = "(" + records.length + " records)";
                parent_element.append(feed_count);
            }
            /* NOTE: If we're not showing headings we're ready to attach our UL list
             * which will be populated record by record, otherwise we need a 
             * alternate with divs and uls for each grouping */
            if (show_year_headings === false) {
                parent_element.appendChild(ul);
            }
            records.forEach(function(record) {
                let current_year = "",
                    li = document.createElement("li"),
                    a,
                    span,
                    div,
                    creators,
                    pub_date,
                    book_title,
                    title,
                    link,
                    description,
                    css_prefix = record.collection;
                if (record.pub_date !== undefined && record.pub_date !== "") {
                    current_year = record.pub_date.substring(1, 5).trim();
                } else {
                    current_year = "unknown year";
                }
                if (show_year_headings === true && current_year != "" && year_heading !== current_year) {
                    if (year_heading === "") {
                        parent_element.classList.add(css_prefix);
                        year_jump_list.classList.add("jump-list");
                    }
                    /* Add link to jump list */
                    year_heading = current_year;
                    a = document.createElement("a");
                    a.classList.add("jump-list-label");
                    if (current_year === "unknown year") {
                        a.classList.add("unknown-year");
                    }
                    a.setAttribute("href", "#" + year_heading);
                    a.setAttribute("title", "Jump to year " + year_heading);
                    a.innerHTML = year_heading;
                    year_jump_list.append(a);

                    /* Add local year element to parent */
                    div = document.createElement("div");
                    div.setAttribute("id", year_heading);
                    div.classList.add("year-heading");
                    if (current_year === "unknown year") {
                        div.classList.add("unknown-year");
                    }
                    div.innerHTML = year_heading;
                    /* Add a new UL list after heading */
                    parent_element.appendChild(div);
                    ul = document.createElement("ul");
                    parent_element.appendChild(ul);
                }
                /* Create our DOM elements, add classes and populate from our common view */
                if (show_creators === true && record.creators.length > 0) {
                    creators = document.createElement("span");
                    creators.classList.add("creator");
                    record.creators.slice(0, 2).forEach(function(creator, i) {
                        if (creator.display_name !== undefined && creator.display_name !== "") {
                            let span = document.createElement("span");
                            if (i > 0) {
                                span.innerHTML = ";";
                                creators.appendChild(span);
                                span = document.createElement("span");
                            }
                            span.classList.add("creator-name");
                            if (creator.orcid !== undefined) {
                                span.setAttribute("title", "orcid: " + creator.orcid);
                            }
                            span.innerHTML = creator.display_name;
                            creators.appendChild(span);
                        }
                    });
                    if (record.creators.length > 2) {
                        creators.append(" et al.");
                    }
                    li.appendChild(creators);
                }
                if (show_pub_date === true && record.pub_date !== undefined && record.pub_date !== "") {
                    pub_date = document.createElement("span");
                    pub_date.classList.add("pub-date");
                    pub_date.innerHTML = " " + record.pub_date + " ";
                    li.appendChild(pub_date);
                }

                title = document.createElement("span");
                title.classList.add("title");
                link = document.createElement("a");
                link.classList.add("link");
                link.setAttribute("href", record.href);
                link.setAttribute("title", "linked to " + record.collection);
                if (show_title_linked === true) {
                    link.innerHTML = record.title;
                    title.appendChild(link);
                    li.appendChild(title);
                } else {
                    title.innerHTML = '<em>' + record.title + '</em>';
                    li.appendChild(title);
                }
                if (record.book_title !== undefined && record.book_title !== "") {
                    book_title = document.createElement("span");
                    book_title.classList.add("book-title");
                    book_title.innerHTML = 'In: <em>' + record.book_title + '</em>';
                    li.appendChild(book_title);
                }
                /* DR-273 makes citation details **individually** selectable
                   citiation checkbox removed DR-135, removed page_range.
                   Pages removed DR-273 as arxiv imported generates pages
                   and they're not part of our currated metadata. */
                [
                    "publisher", "publication", "series", "volume", "number",
                    "chapters", "page_range",
                    "issn", "isbn", "pmcid", 
                    "event_title", "event_dates", "event_location", 
                    "ispublished"
                ].forEach(function(key) {
                    if (record[key] !== undefined &&
                        record[key] !== "") {
                        let span = document.createElement("span"),
                            val = record[key],
                            label = "";
                        span.classList.add(key);
                        switch (key) {
                            case "ispublished":
                                span.innerHTML = val;
                                break;
                            case "publisher":
                                if (show_publisher) {
                                    span.innerHTML = val;
                                }
                                break;
                            case "publication":
                                if (show_publication) {
                                    if (show_title_linked === false) {
                                        span.innerHTML = "; " + val;
                                    } else{
                                        span.innerHTML = val;
                                    }
                                }
                                break;
                            case "volume":
                                if (show_volume === true) {
                                    span.innerHTML = "; Vol. " + val;
                                }
                                break;
                            case "series":
                                if (show_volume === true) {
                                    if (record.number !== undefined && record.number !== "") {
                                        span.innerHTML = "Series " + val + ", " +
                                        record["number"] + ".";
                                    } else {
                                        span.innerHTML = "Series " + val + ".";
                                    }
                                }
                                break;
                            case "number":
                                if (show_issue === true) {
                                    if (record.series === undefined || record.series === "") {
                                        span.innerHTML = "; No. " + val + "";
                                    }
                                }
                                break;
                            case "chapters":
                                if (show_chapters === true) {
                                    span.innerHTML = "; ch. " + val;
                                }
                                break;
                            /* DR-135 remove page_range pages, */
                            /* DR-273 puts pages, ranges back, removes pages as they
                                have dubious value due to the arxiv importer. */
                            case "page_range":
                                if (show_page_numbers === true) {
                                    span.innerHTML = "; pp. " + val;
                                }
                                break; 
                            case "issn":
                                if (show_issn === true) {
                                    span.innerHTML = "ISSN " + val;
                                }
                                break;
                            case "isbn":
                                if (show_isbn === true) {
                                    span.innerHTML = "ISBN " + val;
                                }
                                break;
                            case "pmcid":
                                if (show_pmcid === true) {
                                    span.innerHTML = "PMCID " + val;
                                }
                                break;
                            case "event_title":
                                span.innerHTML = "In: " + val;
                                break;
                            case "event_dates":
                                span.innerHTML = ", " + val;
                                break;
                            case "event_location":
                                span.innerHTML = ", " + val;
                                break;
                            default:
                                label = titleCase(key.replace("_", " "));
                                span.innerHTML = label + " " + val;
                                break;
                        }
                        /* only add the span if we have content */
                        if (span.innerHTML !== "") {
                            li.appendChild(span);
                        }
                    }
                });
                    
                if (show_description === true && record.description !== undefined && record.description !== "") {
                    description = document.createElement("div");
                    description.classList.add("description");
                    description.innerHTML = record.description;
                    li.appendChild(description);
                }
                /* Various links to object or landing page */
                if (show_link === true) {
                    span = document.createElement("span");
                    span.classList.add("official-url");
                    span.innerHTML = `<a href="${record.href}">${record.href}</a>`;
                    li.appendChild(span);
                }
                if (show_doi === true && record.doi !== undefined && record.doi !== "") {
                    span = document.createElement("span");
                    span.classList.add("doi");
                    span.innerHTML = `DOI <a href="https://doi.org/${record.doi}">${record.doi}<a/>`;
                    li.appendChild(span);
                }
                if (show_primary_object === true && record.primary_object !== undefined && record.primary_object.url !== undefined) {
                    span = document.createElement("span");
                    span.classList.add("primary_object");
                    span.innerHTML = `<a href="${record.primary_object.url}">${record.primary_object.label}<a/>`;
                    li.appendChild(span);
                }
                /* Now add our li to the list */
                ul.appendChild(li);
            });
        };
        /* Add it as our final display element in the pipeline */
        filters.push(__display);
        /* Now run our pipeline */
        self.pipeline(data, err, ...filters);
    };

    /* NOTE: we need to update the global CL after adding our methods */
    if (window.CL === undefined) {
        window.CL = {};
    }
    window.CL = Object.assign(window.CL, CL);
}(document, window));
