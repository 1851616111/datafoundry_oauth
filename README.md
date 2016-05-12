# datafoundry_oauth2

# Overview

datafoundry_oauth2 is a service based on oauth2.

It is designed for datafoundry web to

1. authorize datafoundry user to be able to access github.com api

2. reduce data from github api for datafoundry web

datafoundry_oauth2 need etcd storage

# Env Dependent

|     Name                |   Description                   |  Must  |
| ----------------------- | ------------------------------- | ------ |
| GITHUB_REDIRECT_URL     |  oauth2 redirect url on github  |  true  |
| GITHUB_CLIENT_ID        |  oauth2 cliend id on github     |  true  |
| GITHUB_CLIENT_SECRET    |  oauth2 cliend secret on github |  true  |
| DATAFOUNDRY_HOST_ADDR   |  datafoundry api server addr    |  true  |
| ETCD_HTTP_ADDR          |  storage addr                   |  true  |
| ETCD_HTTP_PORT          |  storage port                   |  true  |
| ETCD_USER               |  storage user                   |  true  |
| ETCD_PASSWORD           |  storage password               |  true  |
    
    
     export GITHUB_REDIRECT_URL=http://oauth2-oauth.app.asiainfodata.com/v1/github-redirect  // oauth2 is a router name, oauth is a namespace name
     export GITHUB_CLIENT_ID=2369ed831a59847924b4
     export GITHUB_CLIENT_SECRET=510bb29970fcd684d0e7136a5947f92710332c98
     export DATAFOUNDRY_HOST_ADDR=https://lab.asiainfodata.com:8443
        
     export ETCD_HTTP_ADDR=http://etcdsystem.servicebroker.dataos.io
     export ETCD_HTTP_PORT=2379
     export ETCD_USER=asiainfoLDP
     export ETCD_PASSWORD=6ED9BA74-75FD-4D1B-8916-842CB936AC1A
    
# Running datafoundry_oauth2
start.sh contains a default config to quickly run this service

    GO15VENDOREXPERIMENT=1 go build && ./start.sh
    
# API

# Oauth Callback 

datafoundry web  --request--> Third Oauth --redirect--> datafoundry_oauth2
If success, it will generate a secret in datafoundry namespcae with a name {namespace}-{user}-{third-oauth-name} like oauth-panxy-github.

    > oc get secret 
    
    NAME                            TYPE      DATA      AGE
    oauth-panxy-github   Opaque    1         4m
    

oauth2 document [https://developer.github.com/v3/](https://developer.github.com/v3/)

    
    GET     /v1/github-redirect
    
**Param**
  
|     Name      |     Type      |  Description               |  Must  |
| ------------- | ------------- | -------------------------  | ------ |
| namespace     |     string    |  datafoundry namespace     |  true  |
| bearer        |     string    |  datafoundry bearer token  |  true  |
| code          |     string    |  github callback code      |  true  |
| state         |     string    |  github callback code      |  true  |

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
    
    
# Authorize gitlab deploy 

    POST /v1/gitlab/authorize/deploy

**Description**

    It is designed for datafoundry(DF) user to access private gitlab service.
    It needs to identify different DF Address(dev, release, product, public), if they use the same db storage.
    In each DF Address, it is designed than DF user can add several private gitlab, but only one gitlab user are allowed each gitlab.
        
**Authorization**

    bearer TOKEN 
    
    