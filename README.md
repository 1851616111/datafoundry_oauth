# datafactory_oauth2

# Overview

datafactory_oauth2 is a service based on oauth2.
it is designed for datafoundry web to

1. authorize datafoundry user to be able to access github.com api

2. reduce github.com api return data for datafoundry api

# API

# Oauth Callback 

datafoundry web  --request--> Third Oauth --redirect--> datafactory_oauth2

oauth2 document [https://developer.github.com/v3/](https://developer.github.com/v3/)

    
    GET     /v1/github-redirect

**Header**
  
|     Name      |     Type      |   Description               |  Must  |
| ------------- | ------------- | --------------------------- | ------ |
| namespace     |     string    |  datafoundry  namespace     |  true  |
| user          |     string    |  datafoundry  user          |  true  |
| bearer        |     string    |  datafoundry  bearer token  |  true  |

**Param**
  
|     Name      |     Type      |  Description    |  Must  |
| ------------- | ------------- | --------------- | ------ |
| code          |     string    |  callback code  |  true  |
| state         |     string    |  callback code  |  true  |


**Reponse**
    
    ok

# List owner repos on github

    GET     /v1/repos/github/owne
    
**Authorization**

    bearer TOKEN 

**Reponse**

    [
        {
            "login": "1851616111",
            "repos": [
                {
                    "name": "aerospike-server",
                    "full_name": "1851616111/aerospike-server",
                    "private": false
                },
                {
                    "name": "ql",
                    "full_name": "1851616111/ql",
                    "private": false
                }
            ]
        }
    ]

# List org repos on github
   
    GET     /v1/repos/github/orgs

**Authorization**

    bearer TOKEN 

**Reponse**

    [
        {
            "login": "asiainfoLDP",
            "repos": [
              
                {
                    "name": "datahub_docs",
                    "full_name": "asiainfoLDP/datahub_docs",
                    "private": false
                },
                {
                    "name": "datahub_bill",
                    "full_name": "asiainfoLDP/datahub_bill",
                    "private": true
                }
            ]
        }
    ]

# List user/org repo branches

    GET     /v1/repos/github/users/:user/repos/:repo

**Authorization**

    bearer TOKEN 

**Reponse**
    
    [
        {
            "commit": {
                "sha": "d008a7c3b8614ac7f6597325aa4610b60219fa9a",
                "url": "https://api.github.com/repos/asiainfoLDP/datahub_repository/commits/d008a7c3b8614ac7f6597325aa4610b60219fa9a"
            },
            "name": "develop"
        },
        {
            "commit": {
                "sha": "aa9ab6869f7a517df118e5afe6126626b1581914",
                "url": "https://api.github.com/repos/asiainfoLDP/datahub_repository/commits/aa9ab6869f7a517df118e5afe6126626b1581914"
            },
            "name": "master"
        },
        {
            "commit": {
                "sha": "d008a7c3b8614ac7f6597325aa4610b60219fa9a",
                "url": "https://api.github.com/repos/asiainfoLDP/datahub_repository/commits/d008a7c3b8614ac7f6597325aa4610b60219fa9a"
            },
            "name": "releasev1.8"
        }
    ]