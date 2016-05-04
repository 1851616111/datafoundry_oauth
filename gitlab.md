# Overview

This document is a infrastructure description.
It describe the relationships between private GitLab and DataFoundry.

## Use Cases
 1.  I want to provide GitLab as DataFoundry user's private code repository.
 1.  I want each DataFoundry user could config his private code repository.
 1.  I want each DataFoundry user could add multi GitLab code repository.
 1.  I want to use GitLab Host as a Identifier to distinguish GitLabs, and use user and private token to offer GitLab authentication.
 1.  I want a DataFoundry user could bind just one GitLab account on one GitLab.
 
## API
 
# Provide a GitLib
 
    [POST] /v1/gitlab 

**Param**

|     Name      |     Type      |  Description                     |  Must  |
| ------------- | ------------- | -------------------------------  | ------ |
| host          |     string    |  private GitLab Host addr        |  true  |
| user          |     string    |  private GitLab account          |  true  |
| private_token |     string    |  private GitLab account token    |  true  |

**Header**

|     Key         |     Value      |  Description                     |  Must  |
| --------------- | -------------- | -------------------------------  | ------ |
| Authorization   | bearer TOKEN   |  TOKEN is DataFoundry Token      |  true  |

**Curl**

curl http://127.0.0.1:9443/v1/gitlab  -d '{"host":"https://code.dataos.io", "user":"mengjing","private_token":"fXYznpUCTQQe5sjM4FWm"}' -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8"

**Description**

 1. use request header to auth for DataFoundry. 
 2. use request param to auth for GitLab.
 3. if two auth paas, gitLabInfo will store into etcd. 
 
    
|     Key       |     Value      | 
| ------------- | -------------- |
| /df_service/DATAFOUNDRYHOST/df_user/DATAFOUNDRYUSER/oauth/gitlabs/info    |     gitLabInfo    |
  
    DATAFOUNDRYHOST:  DATAFOUNDRYHOST is a DataFoundry Host Address like https://lab.dataos.io:8443 
                      DATAFOUNDRYHOST is a variable  
    
    DATAFOUNDRYUSER:  DATAFOUNDRYUSER is a DataFoundry Account Username
                      DATAFOUNDRYUSER is transferred from Param TOKEN 

        gitLabInfo is used to store GitLab information.
        type gitLabInfo struct {
    	    Host         string `json:"host"`
    	    User         string `json:"user"`
    	    PrivateToken string `json:"private_token"`
        }

# Bind a GitLib to DataFoundry User
 
    [POST] /v1/gitlab/authorize/deploy

**Param**

|     Name      |     Type      |  Description                                          |  Must  |
| ------------- | ------------- | ----------------------------------------------------  | ------ |
| host          |     string    |  host a GitLab Host                                   |  true  |
| project_id    |     int       |  project_id is a project identifier in GitLab Host    |  true  |

**Header**

|     Key         |     Value              |  Description                                                                             |  Must  |
| --------------- | ---------------------- | ---------------------------------------------------------------------------------------  | ------ |
| Authorization   | bearer TOKEN           |  TOKEN is DataFoundry Token                                                              |  true  |
| namespace       | Namespace              |  Namespace is a DataFoundry project name used to asign which project is being binding    |  true  |

**Curl**

curl http://127.0.0.1:9443/v1/gitlab/authorize/deploy -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8" -H "namespace:oauth" -d '{"host":"https://code.dataos.io","project_id":43}'

**Description**

 1. 向DataFoundry 验证Token是否有效,若有效,返回当前Token的用户信息
 2. 查询当前用户注册的主机信息列表包括(host, user, private token), 如果需要绑定的主机列表不在用户之前提供的列表中,则返回
 3. 查询需要绑定的主机下的project的deploy key, 若key的title属性满足(df_host---DATAFOUNDRYHOST---df_user---DATAFOUNDRYUSER). 
    这个设计为同时区分不同的DataFoundry主机及用户.如,一个DataFoundry地址为https://lab.dataos.io:8443,存在用户panxy3. 若该用户在一个GitLab(host=https://code.dataos.io)的代码库下的所有project下面生成的DeployKey相同,
 4. 如果(title=df_host---https://lab.dataos.io:8443---df_user---存在用户panxy3)不存在deploykey,则需要为这个GitLab的project创建一个相应的DeployKey,同时向存储保存这个用户DeployKey的全部信息
    如果(title=df_host---https://lab.dataos.io:8443---df_user---存在用户panxy3)存在deploykey,则从存储中查询出之前第一次创建时候的信息.
 5. 将上一步的DeployKey信息的privateKey,创建一个Secret.