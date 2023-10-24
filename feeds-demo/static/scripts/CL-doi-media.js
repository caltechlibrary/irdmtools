/**
 * CL-doi-media.js adds an embedded video player generated from
 * DOI media metadata.
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
    'use strict';

    let CL = {};
    if (window.CL === undefined) {
        window.CL = {};
    } else {
        CL = Object.assign({}, window.CL);
    }

    function getDoiMediaType(obj) {
        if ('attributes' in obj && 'mediaType' in obj.attributes) {
            return obj.attributes.mediaType;
        }
        return '';
    }

    function getDoiMediaURL(obj) {
        if ('attributes' in obj && 'url' in obj.attributes) {
            return obj.attributes.url;
        }
        return '';
    }


    CL.doi_media = function (doi, item_no, fnRenderCallback) {
        let self = this,
            doi_url = 'https://api.datacite.org/dois/' + doi + '/media';
        self.httpGet(doi_url, "application/json", function(data, err) {
            if (err) {
                return fnRenderCallback({}, err);
            }
            if ('data' in data && Array.isArray(data.data) && data.data.length > item_no) {
                let media_url = getDoiMediaURL(data.data[item_no]),
                    media_type = getDoiMediaType(data.data[item_no]),
                    err = '';
                if (media_url === '') {
                    err += ' missing url';
                }
                if (media_type === '') {
                    err += ' missing media type';
                }
                return fnRenderCallback({"media_url": media_url, "media_type": media_type}, err);
            }
            return fnRenderCallback({}, 'No media found for ' + doi);
        });
    };

    CL.doi_video_player = function(elem, doi, item_no, width = 640, height = 480) {
        let self = this;
        if (item_no === undefined) {
            item_no = 0;
        }
        self.doi_media(doi, item_no, function(obj, err) {
            if (err) {
                elem.innerHTML = `Could not render ${doi}, ${err}`;
                return;
            }
            elem.innerHTML = `<link href="https://vjs.zencdn.net/7.5.5/video-js.css" rel="stylesheet">
<script src="https://vjs.zencdn.net/7.5.5/video.js"></script>
<!-- If you'd like to support IE8 -->
<script src="https://vjs.zencdn.net/ie8/1.1.2/videojs-ie8.min.js"></script>

<video class="video-js" controls preload="auto" width="${width}" height="${height}" data-setup="{}">
    <source src="${obj.media_url}" type='${obj.media_type}'>
    <p class="vjs-no-js">
      To view this video please enable JavaScript, and consider upgrading to a web browser that
      <a href="http://videojs.com/html5-video-support/" target="_blank">supports HTML5 video</a>
    </p>
</video>`;
        });
    };

    /* NOTE: we need to update the global CL after adding our methods */
    if (window.CL === undefined) {
        window.CL = {};
    }
    window.CL = Object.assign(window.CL, CL);
}(document, window));
