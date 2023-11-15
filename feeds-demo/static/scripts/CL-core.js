/**
 * CL-core.js provides browser side JavaScript access to 
 * Caltech Library resources (e.g. feeds.library.caltech.edu).
 * It also provides common functions and objects used in various
 * Caltech Library projects.
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
(function (document, window) {
    "use strict";
    /* CL is our root object */
    let CL = {},
        Version = 'v0.2.2';


    if (window.CL === undefined) {
        window.CL = CL;
    } else {
        CL = Object.assign(CL, window.CL);
    }
    CL.Version = Version;

    /**
     * CL.pipeline() takes a data, error and list of functions as 
     * parameters.  It then creates a clone of the current state 
     * adds adds the list of functions passed to .pipelineFns as 
     * well as a .nextCalbackFn() method that is will take the 
     * data and error parameters after shifting the first function 
     * in .pipelineFns and envoking it with the data and error
     * parameters. The shifted out callback can then envoke 
     * this.nextCallbackFn(data, err) as needed to continue the 
     * pipe line.
     *
     * @params data (any valid JavaScript type)
     * @params err (string) holds an error message from calling function
     * @params ...pipelineFns (split of functions), functions to run the pipeline with.
     */
    CL.pipeline = function(data, err, ...pipelineFns) {
        let self = this;
        if (pipelineFns.length < 1) {
            console.log("ERORR: empty pipeline");
            return;
        }
        self.pipelineFns = pipelineFns;
        self.nextCallbackFn = function (data, err) {
            let obj = Object.assign({}, self);
            obj.callbackFn = self.pipelineFns.shift();
            if (obj.callbackFn === undefined) {
                return;
            }
            obj.callbackFn(data, err);
        };
        if (self.pipelineFns.length > 0) {
            self.nextCallbackFn(data, err);
        }
    };

    /**
     * setAttribute is used to populate a collection of attributes
     * used by various functions (e.g. recentN and viewer). It is
     * a simple map of function name to JavaScript value.
     *
     * @param name the name of the attribute (e.g. recentN, viewer)
     * @param value the value to associate with the name
     */
    CL.setAttribute = function (name, value) {
        let self = this;
        if (self._attributes === undefined) {
            self._attributes = new Map();
        }
        self._attributes.set(name, value);
    };
    
    /**
     * CL.getAttribute() returns the value if the attribute or undefined
     * if not found.
     *
     * @param name the name of the attribute to retrieve
     * @return the value of attribute or undefined if not found.
     */
    CL.getAttribute = function(name) {
        let self = this;
        if (self._attributes !== undefined && 
            self._attributes.has(name)) {
            return self._attributes.get(name);
        }
    };

    /**
     * CL.hasAttribute() returns true if attribute exists, false otherwise
     *
     * @param name the name of the attribute to check
     * @return true if found, false otherwise
     */
    CL.hasAttribute = function(name) {
        let self = this;
        if (self._attributes !== undefined) {
            return self._attributes.has(name);
        }
        return false;
    };


    /**
     * CL.httpGet() - makes an HTTP get request and returns the results 
     * via callbackFn.
     *
     * @param url (a URL object) the assembled URL (including any GET args)
     * @param contentType - string of indicating mime type 
     *        (e.g. text/html, text/plain, application/json)
     * @param callbackFn - an function to handle the callback, 
     *        function takes two args data (an object) and 
     *        error (a string)
     */
    CL.httpGet = function (url, contentType, callbackFn) {
        let self = this,
            xhr = new XMLHttpRequest(),
            page_url = new URL(window.location.href);
        xhr.onreadystatechange = function () {
            // process response
            if (xhr.readyState === XMLHttpRequest.DONE) {
                if (xhr.status == 200) {
                    let data = xhr.responseText;
                    if (contentType === "application/json" && data !== "") {
                        data = JSON.parse(xhr.responseText);
                    }
                    callbackFn(data, "");
                } else {
                    callbackFn("", xhr.status);
                }
            }
        };

        /* NOTE: Check to see if we should turn a string version of URL 
         * into a URL object. Handle case of applying a BaseURL prefix
         * if protocol/host is missing */
        if (typeof url === "string") {
            if (url.startsWith("/") && self.BaseURL !== undefined) {
                /* NOTE: combined our BaseURL string with url as 
                 * root relative pathname, then re-cast to URL object */
                url = new URL(self.BaseURL + url);
            } else {
                url = new URL(url);
            }
        } 
        if (page_url.username !== undefined && url.username === undefined) {
            url.username = page_url.username;
        }
        if (page_url.password !== undefined && url.password == undefined) {
            url.password = page_url.password;
        }

        /* we always want JSON data */
        xhr.open('GET', url, true);
        if (url.pathname.includes(".json.gz") || url.pathname.includes(".js.gz")) {
            xhr.setRequestHeader('Content-Encoding', 'gzip');
        }
        if (contentType !== "" ) {
            xhr.setRequestHeader('Content-Type', contentType);
        }
        if (self.hasAttribute("progress_bar")) {
            let progress_bar = self.getAttribute("progress_bar");
            xhr.onprogress = function(pe) {
                if (pe.lengthComputable) {
                    progress_bar.max = pe.total;
                    progress_bar.value = pe.loaded;
                }
            };
            xhr.onloadend = function(pe) {
                progress_bar.value = pe.loaded;
            };
        }
        xhr.send();
    };

    /**
     * CL.httpPost() - makes an HTTP POST request and returns the results 
     * via callbackFn.
     *
     * @param url (string) the assembled URL (including any GET args)
     * @param contentType - string of indicating mime type 
     *        (e.g. text/html, text/plain, application/json)
     * @param payload - the text you want to POST
     * @param callbackFn - an function to handle the callback, 
     *        function takes two args data (an object) and 
     *        error (a string)
     */
    CL.httpPost = function (url, contentType, payload, callbackFn) {
        let self = this,
            xhr = new XMLHttpRequest(),
            page_url = new URL(window.location.href);
        xhr.onreadystatechange = function () {
            // process response
            if (xhr.readyState === XMLHttpRequest.DONE) {
                if (xhr.status == 200) {
                    let data = xhr.responseText;
                    if (contentType === "application/json" && data !== "") {
                        data = JSON.parse(xhr.responseText);
                    }
                    callbackFn(data, "");
                } else {
                    callbackFn("", xhr.status);
                }
            }
        };

        /* NOTE: Check to see if we should turn a string version of URL 
         * into a URL object. Handle case of applying a BaseURL prefix
         * if protocol/host is missing */
        if (typeof url == "string") {
            if ( url.startsWith("/") && self.BaseURL !== undefined) {
                url = new URL(self.BaseURL + url);
            } else {
                url = new URL(url);
            }
        }
        if (page_url.username !== undefined && url.username === undefined) {
            url.username = page_url.username;
        }
        if (page_url.password !== undefined && url.password == undefined) {
            url.password = page_url.password;
        }

        /* we always want JSON data */
        xhr.open('POST', url, true);
        if (contentType !== "" ) {
            xhr.setRequestHeader('Content-Type', contentType);
        }
        if (self.hasAttribute("progress_bar")) {
            let progress_bar = self.getAttribute("progress_bar");
            xhr.onprogress = function(pe) {
                if (pe.lengthComputable) {
                    progress_bar.max = pe.total;
                    progress_bar.value = pe.loaded;
                }
            };
            xhr.onloadend = function(pe) {
                progress_bar.value = pe.loaded;
            };
        }
        xhr.send(payload);
    };

    window.CL = Object.assign(window.CL, CL);
}(document, window));
