import sys
import os
import copy
import json
import requests

import s3fs
from requests import session
from json.decoder import JSONDecodeError
from .files import write_files_rdm, send_to_community
from .customize_schema import customize_schema
from .utils import humanbytes
from .fixups import fixup_record
#from .get_metadata import get_metadata
#from .download_file import download_file, download_url



class IRDM_Client:
    def __init__(self, api_url = None, token = None, schema ="43", s3 = None, repo = None, community = None, production = False, publish = False):
        """
        __init__ configures the client for working with the Invenio-RDM
        API for Caltech Library's Invenio-RDM repositories. The object's
        attributes contain specifics for Caltech Library's development
        and production deployments. E.g. the URLs can be set by leaving
        the api_url parameter unset and passing one of three repository
        names - "data", "authors" and "thesis" and setting the production
        value to True or False. The production parameter will also set
        The DOI prefixes that are used by the client object.
        """
        self.token = token
        self.repo = repo
        self.production = production
        self.publish = publish
        self.schema = schema
        self.s3 = s3
        self.community = community
        self.url = api_url
        if not self.url and "RDM_URL" in os.environ:
            self.url = os.environ("RDM_URL")
        if not repo and "RDM_REPO" in os.environ:
            self.repo = os.environ("RDM_REPO")
        if not token and "RDMTOK" in os.environ:
            self.token = os.environ("RDMTOK")

        if not self.url:
            if not self.repo:
                raise Exception(f'repository not set in environment (i.e. RDM_REPO) or client initialization')
            if self.repo == 'data':
                if self.production == True:
                    self.url = "https://data.caltech.edu/"
                else:
                    self.url = "https://data.caltechlibrary.dev/"
            elif self.repo == 'authors':
                if production == True:
                    self.url = "https://authors.caltech.edu/"
                else:
                    self.url = "https://authors.caltechlibrary.dev/"
        if not self.url:
           raise Exception(f'Unable to determine IRDM API URL')

    def query(self, query_string, sort = None, size = None, page = None, allversions = False):
        """
        Send a record query to Invenio-RDM's Elasticsearch
        """
        token = self.token
        repo = self.repo
        url = self.url
        params = { "q": query_string }
        if sort != None:
            params["sort"] = sort
        if size != None:
            params["size"] = size
        if page != None:
            params["page"] = page
        if allversions != None:
            params["allversions"] = allversions

        headers = {
            "Authorization": "Bearer %s" % token,
            "Content-type": "application/json",
        }
        f_headers = {
            "Authorization": "Bearer %s" % token,
            "Content-type": "application/octet-stream",
        }
    
        # Check status
        result = requests.get(
            url + "/api/records",
            headers = headers,
            params = params,
        )
        if result.status_code != 200:
            raise Exception(result.text)
        if not result.text:
            raise Exception(f'expected result.text for result.status_code {result.status_code}')
        obj = json.loads(result.text)
        return obj 
        

    def create(self, metadata, files = [], file_links = [], doi = None):
        """
        File links are links to files existing in external systems that
        will be added directly in a CaltechDATA record, instead of
        uploading the file.

        S3 is a s3sf object for directly opening files
        """
        #community = self.community # DEBUG May want to skip community for debugging migration
        community = None
        repo = self.repo
        token = self.token
        url = self.url
        s3 = self.s3
        publish = self.publish
        production = self.production

        # Pull out pid information
        if production == True:
            repo_prefix = "10.22002"
        else:
            repo_prefix = "10.33569"

        # If files is a string - change to single value array
        if isinstance(files, str) == True:
            files = [files]

        if file_links:
            metadata = add_file_links(metadata, file_links)

        pids = {}
        if doi != None:
            prefix = doi.split("/")[0]
            if prefix == repo_prefix:
                pids["doi"] = {
                        "identifier": doi,
                        "provider": "datacite",
                        "client": "datacite",
                    }
            else:
                pids["doi"] = {
                        "identifier": doi,
                        "provider": "external",
                    }
    
        metadata["pids"] = pids 

        # See if we're working with a schema like DataCite 43 or
        # raw Invenio records.
        if not self.schema:
            data = metadata
        else:
            data = customize_schema(copy.deepcopy(metadata), schema=self.schema)
    
        headers = {
            "Authorization": "Bearer %s" % token,
            "Content-type": "application/json",
        }
        f_headers = {
            "Authorization": "Bearer %s" % token,
            "Content-type": "application/octet-stream",
        }

        # NOTE: fixup_record takes the simple record making final changes
        # suitable for importing into Invenio-RDM.  This include things
        # like crosswalking vocabularies to map from an existing 
        # Caltech Library EPrints repository to Invenio-RDM.
        data = fixup_record(data, files)

        # Make draft and publish
        result = requests.post(url + "/api/records", headers=headers, json=data)
        if result.status_code != 201:
            raise Exception(result.text)
        idv = result.json()["id"]
        publish_link = result.json()["links"]["publish"]
    
        if files:
            file_link = result.json()["links"]["files"]
            write_files_rdm(files, file_link, headers, f_headers, s3)
    
        if community:
            review_link = result.json()["links"]["review"]
            send_to_community(review_link, data, headers, publish, community)
        else:
            if publish:
                result = requests.post(publish_link, headers=headers)
                if result.status_code != 202:
                    raise Exception(result.text)
        return idv
    

    def read(self, idv):
        token = self.token
        repo = self.repo
        url = self.url
    
        headers = {
            "Authorization": "Bearer %s" % token,
            "Content-type": "application/json",
        }
        f_headers = {
            "Authorization": "Bearer %s" % token,
            "Content-type": "application/octet-stream",
        }
    
        # Check status
        result = requests.get(
            url + "/api/records/" + idv,
            headers=headers,
        )
        if result.status_code != 200:
            # Might have a draft
            result = requests.get(
                url + "/api/records/" + idv + "/draft",
                headers=headers,
            )
            if result.status_code != 200:
                raise Exception(result.text)
        if not result.text:
            raise Exception(f'expected result.text for result.status_code {result.status_code}')
        obj = json.loads(result.text)
        return obj 
    
