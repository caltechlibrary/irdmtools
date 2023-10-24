/**
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
/* jshint esversion: 6 */
(function (document, window) {
    "use strict";
    /* CL is our root object */
    let CL = {};

    if (window.CL === undefined) {
        window.CL = CL;
    } else {
        CL = Object.assign(CL, window.CL);
    }

    /********************
     * FeedsBaseURL: this normally should be 
     *
     *        https://feeds.library.caltech.edu
     *
     * It maybe changed to support testing and development 
     * versions of feeds content.
     ********************/
    CL.FeedsBaseURL = 'https://feeds.library.caltech.edu';

    /**
     * getFeed allows you to fetch the raw feed as plain text. 
     * You can use this to form your own custom queries or as 
     * a debug tool.
     *
     * @param feedURL (string, required) the URL to fetch
     * @param callbackFn (function, required) the callback function 
     *        to process the results the callaback function has two 
     *        parameters data and err where err is a string holding 
     *        an error message if something went wrong or an empty 
     *        string if everything was successful. data will hold any
     *        data returned
     */
    CL.getFeed = function (feedURL, callbackFn) {
        this.httpGet(feedURL, "text/plain", callbackFn);
    };

    /**
     * CL.getPeopleList() fetches the group list as an array of group objects.
     *
     * @param callbackFn (function, required) is the function that processes the list.
     */
    CL.getPeopleList = function (callbackFn) {
       let self = this,
            url = self.FeedsBaseURL + "/people/people_list.json";
       this.httpGet(url, "application/json", callbackFn);
    };
    
    /**
     * CL.getPeopleInfo() fetch the /people/[peopleID]/people.json
     * so you can build a list of available feed types (e.g. article, recent/article).
     * and other useful this.
     *
     * @param peopleID (string, required) e.g. Newman-D-K
     * @param callbackFn (function, required) is a function that has two parameters - data and error
     */
    CL.getPeopleInfo = function (peopleID, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + "/people/" + peopleID + "/people.json";
        this.httpGet(url, "application/json", callbackFn);
    };


    /**
     * CL.getPeopleInclude() fetches a people based HTML include feed
     * and envokes the callback function provided. 
     *
     * @param personID - an internal Caltech Library Person identifier 
     *          like the creator ID used in EPrints repository systems.
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getPeopleInclude = function(personID, feedName, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/people/' + personID + 
                '/' + feedName.toLowerCase() + '.include';
        this.httpGet(url, "text/plain", callbackFn);
    };

    /**
     * CL.getPeopleJSON() fetches person based JSON feed and envokes 
     * the callback function provided.
     *
     * @param personID - an internal Caltech Library Person identifier 
     *          like the creator ID used in EPrints repository systems.
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getPeopleJSON = function (personID, feedName, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/people/' + personID + 
            '/' + feedName.toLowerCase() + '.json';
        this.httpGet(url, "application/json", callbackFn);
    };

    /**
     * CL.getPeopleCustomJSON() fetchs the people based JSON feed and
     * filters against a list of ids before calling the callback function
     * provided.
     *
     * @param peopleID - a string identifying the people found in the people URL of feeds
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param idList is a JavaScript list of ids to filter against
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getPeopleCustomJSON = function (peopleID, feedName, idList, callbackFn) {
        let self = this, 
            url = self.FeedsBaseURL + '/people/' + peopleID + 
            '/' + feedName.toLowerCase() + '.json';
        this.httpGet(url, "application/json", function(object_list, err) {
            if (err != "") {
                callbackFn([], err);
                return;
            }
            let object_map = {},
                filtered_list = [];

            /* Build a map of id to item */
            object_list.forEach(function(obj) {
                let key = obj._Key;
                object_map[key] = obj;
            });
            /* Using our map, build our filtered list */
            idList.forEach(function (id) {
                let key = id.toString();
                if (key in object_map) {
                    filtered_list.push(object_map[key]);
                }
            });
            /* Now we're ready to pass to the callback function */
            callbackFn(filtered_list, "");
        });
    };

    /**
     * CL.getPeopleKeys() fetches a person based list of keys and
     * the callback function provided.
     *
     * @param personID - an internal Caltech Library Person identifier 
     *          like the creator ID used in EPrints repository systems.
     * @param feedName string (e.g. combined, article, book, monograph)
     */
    CL.getPeopleKeys = function(personID, feedName, callbackFN) {
        let self = this,
            url = self.FeedsBaseURL + '/people/' + personID + '/' + feedName.toLowerCase() + '.keys';
        this.httpGet(url, "text/plain", function (data, err) {
            if (err) {
                callbackFn([], err);
                return;
            }
            callbackFn(data.split("\n"), err);
        });
    };

    /**
     * CL.getGroupsList() fetches the group list as an array of group objects.
     *
     * @param callbackFn is the function that processes the list.
     */
    CL.getGroupsList = function(callbackFn) {
       let self = this,
            url = self.FeedsBaseURL + "/groups/group_list.json";
       this.httpGet(url, "application/json", callbackFn);
    };

    /**
     * CL.getGroupInfo() fetch the /groups/GROUP_ID/group.json
     * so you can build a list of available feed types (e.g. article, recent/article).
     * and other useful this.
     *
     * @param groupID (string, required) e.g. GACIT, COSMOS, Caltech-Library
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getGroupInfo = function (groupID, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + "/groups/" + groupID + "/group.json";
        this.httpGet(url, "application/json", callbackFn);
    };

    /**
     * CL.getGroupSummary() fetches the group summary data which 
     * may include description, aprox_start, aprox_end as well as
     * alternative names.
     *
     * @param groupID - a string identifying the group like that found in the group URL of feeds
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getGroupSummary = function(groupID, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/groups/' + groupID + 
            '/group.json';
        this.httpGet(url, "application/json", function(data, err) {
            if (err) {
                callbackFn({}, err);
                return;
            }
            // Prune the object to a summary
            if ('_Key' in data) {
                delete data._Key;
            }
            if ('email' in data) {
                delete data.email;
            }
            if ('CaltechTHESIS' in data) {
                delete data.CaltechTHESIS;
            }
            if ('CaltechAUTHORS' in data) {
                delete data.CaltechAUTHORS;
            }
            if ('CaltechDATA' in data) {
                delete data.CaltechDATA;
            }
            callbackFn(data, err);
        });
    };

    /**
     * CL.getGroupInclude() fetches a group based HTML include feed
     * and envokes the callback function provided. 
     *
     * @param groupID - a string identifying the group like that found in the group URL of feeds
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getGroupInclude = function(groupID, feedName, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/groups/' + groupID + 
            '/' + feedName.toLowerCase() + '.include';
        this.httpGet(url, "text/plain", callbackFn);
    };

    /**
     * CL.getGroupJSON() fetches group based JSON feed and envokes 
     * the callback function provided.
     *
     * @param groupID - a string identifying the group like that found in the group URL of feeds
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getGroupJSON = function (groupID, feedName, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/groups/' + groupID + 
            '/' + feedName.toLowerCase() + '.json';
        this.httpGet(url, "application/json", callbackFn);
    };

    /**
     * CL.getGroupCustomJSON() fetchs the group based JSON feed and
     * filters against a list of ids before calling the callback function
     * provided.
     *
     * @param groupID - a string identifying the group like that found in the group URL of feeds
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param idList is a JavaScript list of ids to filter against
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getGroupCustomJSON = function (groupID, feedName, idList, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/groups/' + groupID + 
            '/' + feedName.toLowerCase() + '.json';
        this.httpGet(url, "application/json", function(object_list, err) {
            if (err != "") {
                callbackFn([], err);
                return;
            }
            let object_map = {},
                filtered_list = [];

            /* Build a map of id to item */
            object_list.forEach(function(obj) {
                let key = obj._Key;
                object_map[key] = obj;
            });
            /* Using our map, build our filtered list */
            idList.forEach(function (id) {
                let key = id.toString();
                if (key in object_map) {
                    filtered_list.push(object_map[key]);
                }
            });
            /* Now we're ready to pass to the callback function */
            callbackFn(filtered_list, "");
        });
    };


    /**
     * CL.getGroupKeys() fetches group base key list and envokes
     * the callback function provided.
     *
     * @param groupID - a string identifying the group like that found in the group URL of feeds
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getGroupKeys = function(groupID, feedName, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/groups/' + groupID + 
            '/' + feedName.toLowerCase() + '.keys';
        this.httpGet(url, "text/plain", function (data, err) {
            if (err) {
                callbackFn([], err);
                return;
            }
            callbackFn(data.split("\n"), err);
        });
    };


    /**
     * CL.getPersonInclude() fetches person based HTML feed and envokes 
     * the callback function provided.
     *
     * @param orcid string representation of the ORCID
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getPersonInclude = function (orcid, feedName, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/person/' + orcid + '/' + feedName.toLowerCase() + '.include';
        this.httpGet(url, "text/plain", callbackFn);
    };

    /**
     * CL.getPersonJSON() fetches person based JSON feeds and envokes the callback function provided.
     *
     * @param orcid string representation of the ORCID
     * @param feedName string (e.g. combined, article, book, monograph)
     * @param callbackFn is a function that has two parameters - data and error
     */
    CL.getPersonJSON = function (orcid, feedName, callbackFn) {
        let self = this,
            url = self.FeedsBaseURL + '/person/' + orcid + '/' + feedName.toLowerCase() + '.json';
        this.httpGet(url, "application/json", callbackFn);
    };


    /* NOTE: we need to update the global CL after adding our methods */
    if (window.CL === undefined) {
        window.CL = {};
    }
    window.CL = Object.assign(window.CL, CL);
}(document, window));
