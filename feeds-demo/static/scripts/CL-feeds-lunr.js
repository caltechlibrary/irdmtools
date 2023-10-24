/**
 * CL-feeds-lunr.js holds CL feed filters wrapped for use by LunrJS.
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
     * lunr_search filters the search results based on
     * the query string if set. Otherwise it is just a pass-
     */
    CL.lunr_search = function(data, err) {
        let self = this;
        if (err !== "") {
            self.nextCallbackFn(data, err);
            return;
        }
        // get query
        let u = new URL(window.location.href),
            params = new URLSearchParams(u.search),
            q = params.get("q"),
            records = [],
            results = [];

        // Build an index if query is present. Otherwise
        // Return unfiltered data.
        if (q && q !== "") {
            let idx = lunr(function() {
                this.ref("_i");
                this.field("href");
                this.field("title");
                this.field("creators");
                this.field("description");
                this.field("pub_date");
                this.field("collection");
                this.field("doi");
                this.field("citation_info");
                this.field("resource_type");
                for (let i in data) {
                    data[i]._i = i;
                    this.add(data[i]);
                }
            });
            results = idx.search(q);
            for (let i in results) {
                let j = results[i].ref;
                records.push(data[j]);
            }
            self.nextCallbackFn(records, "");
            return;
        }
        
        // Call the next filter function 
        self.nextCallbackFn(data, err);
        return;
    };


    if (window.CL === undefined) {
        window.CL = {};
    }
    window.CL = Object.assign(window.CL, CL);
}(document, window));
