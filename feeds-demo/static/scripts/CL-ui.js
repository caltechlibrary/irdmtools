/**
 * CL-ui.js provides browser side JavaScript form building
 * functions for Caltech Library resources (e.g. feeds.library.caltech.edu).
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

/* jshint esversion: 6 */
(function(document, window) {
    "use strict";
    let CL = {};
    if (window.CL === undefined) {
        window.CL = {};
    } else {
        CL = Object.assign({}, window.CL);
    }

    /**
     * __template processes a Python-like template string, 
     * containing object attribute names with 
     * prefixes of `{{` and suffixes of `}}` and replaces
     * them with the attributes's value. The processed string
     * is then returned by the function.
     */ 
    function __template(tmpl, obj, sep = "") {
        let out = tmpl;
        for (let key in obj) {
            let re = new RegExp('{{' + key + '}}', 'g');
            if (obj[key] === undefined) {
                console.log(`ERROR: you have assigned 'undefined' to object attribute, ${key} -> ${obj[key]}`);
            } else if (Array.isArray(obj[key])) {
                let a = [];
                for (let i in obj[key]) {
                    if (obj[key][i].html !== undefined) {
                        a.push(obj[key][i].html())
                    }
                }
                out = out.replace(re, a.join(sep));
            } else if (obj[key].html !== undefined) {
                out = out.replace(re, obj[key].html())
            } else {
                out = out.replace(re, obj[key])
            }
        }
        return out;
    }

    /**
     * field takes a default_attributes object, a template
     * string and an optional init function. It returns 
     * an object that has the following functions - init(), get(), set(), 
     * html(), and json().
     *
     * Example:
     *
     *      creator = CL.field({
     *          last_name: "Jones",
     *          first_name: "Henry",
     *          birth_date: "July 1, 1899"
     *          },
     *          '<div>' +
     *          '   <label>Last Name:</label>' +
     *          '   <input name="last_name" value="{{last_name}}">' +
     *          '</div>' +
     *          '<div>' +
     *          '  <label>First Name:</label>' +
     *          '  <input name="first_name" value="{{first_name}}">' +
     *          '</div>' +
     *          '<div>' +
     *          '  <label>Birth Date:</label>' +
     *          '  <input name="birth_date" value="{{birth_date}}">' +
     *          '</div>',
     *          function(obj) {
     *             // Normalize date before rendering form.
     *             if (('birth_date' in obj) && obj.birth_date !== "") {
     *                dt = new Date(obj.birth_date);
     *                obj.birth_date = dt.toDateString()
     *             }
     *          });
     *          
     *     // Render as HTML
     *     element.innerHTML = creator.html();
     */
    CL.field = function(attributes, template_string, init_function = undefined, sep = "") {
        let obj = new Object();
        // Shallow copy of object attributes
        for (let key in attributes) {
            obj[key] = attributes[key];
        }
        // Attach our init function
        if (init_function === undefined) {
            //NOTE: Our default init function always succeeds.
            obj.init = function () { return true };
        } else {
            //NOTE: User supplied init_function should return
            // true on success, false otherwise.
            obj.init = init_function;
        }
        // Add our get(), set(), html(), and json functions()
        obj.get = function(key, error_value) {
            let self = this;

            if (key in self) {
                return self[key];
            }
            if (error_value == undefined) {
                return null;
            }
            return error_value;
        }
        obj.set = function(key, value){
            let self = this;
            obj[key] = value;
        }
        obj.html = function() {
            let obj = this;
            return __template(template_string, obj, sep);
        }
        obj.json = function() {
            let self = this;
            return JSON.stringify(self);
        }
        return obj;
    }

    /**
     * assembleFields takes a DOM element and appends new
     * DOM elements from the html() rendering of the individual
     * fields passed to CL.assembleFields().
     *
     * Example:
     *
     *     let book = {},
     *         books = [],
     *         creators = [],
     *         steinbeck = {
     *            last_name: "Steinbeck",
     *            first_name: "John"
     *         },
     *         pratchett = {
     *            last_name: "Pratchett",
     *            first_name: "Terry"
     *         },
     *         gaiman = {
     *            last_name: "Gaiman",
     *            first_name: "Neil",
     *         };
     *
     *     steinbeck = CL.field(steinbeck, 
     *         '<span class="last_name">{{last_name}}</span>, ' +
     *         '<span class="first_name">{{first_name}}</span>');
     *
     *     creators = CL.field({"creators": [ steinbeck ]},
     *         '<div class="creators">By {{creators}}</div>',
     *         sep = '; ');
     *
     *     book = CL.field({
     *          "title": "Short Reign of Pippen IV"
     *          "description": "A novella length satire set in post-war Paris", 
     *          "creators": creators
     *         }, 
     *         '<div class="book">' +
     *         '   <div class="title">{{title}}</div>' +
     *         '   <div class="creators">By {{creators}}</div>' + 
     *         '   <div class="description">{{description}}</div>' +
     *         '</div>'
     *         undefined, '; ');
     *     books.push(book);
     *
     *     pratchett = CL.field(pratchett, 
     *         '<span class="last_name">{{last_name}}</span>, ' +
     *         '<span class="first_name">{{first_name}}</span>');
     *
     *     gaimen = CL.field(gaimen, 
     *         '<span class="last_name">{{last_name}}</span>, ' +
     *         '<span class="first_name">{{first_name}}</span>');
     *
     *     creators = CL.field({"creators": [ pratchett, gaimen ]},
     *         '<div class="creators">By {{creators}}</div>',
     *         sep = '; ');
     *
     *     // NOTE: We attach normalizeBookData for the init function 
     *     // which is called by assembleFields initializing the 
     *     // data before rendering.
     *     book = CL.field({
     *          "title": "Good Omens"
     *          "description": "A book about angels and demons set in London for the most part", 
     *          "creators": creators
     *         }, 
     *         '<div class="book">' +
     *         '   <div class="title">{{title}}</div>' +
     *         '   <div class="creators">By {{creators}}</div>' + 
     *         '   <div class="description">{{description}}</div>' +
     *         '</div>'
     *         normalizeBook, '; ');
     *     books.push(book);
     *
     *     let element = CL.assembleFields(
     *          document.getElementById("featured-book"), ...books);
     *
     */
    CL.assembleFields = function(element, ...field_list) {
        let fields = field_list;

        element.innerHTML = "";
        if (Array.isArray(fields)) {
            for (let key in fields) {
                if (fields[key].init !== undefined && 
                        fields[key].html !== undefined) {
                    fields[key].init();
                    element.innerHTML += fields[key].html();
                }
            }
        } else if (fields.html !== undefined) {
            element.innerHTML += fields.html();
        }
        return element;
    }

    /* 
     * NOTE: we need to update the global CL after 
     * adding our methods 
     */
    if (window.CL === undefined) {
        window.CL = {};
    }
    window.CL = Object.assign(window.CL, CL);
}(document, window));
