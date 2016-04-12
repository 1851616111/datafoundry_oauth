os::getEnv() {
    key=$1
    pair=`env | grep ${key}`
    echo ${pair#*$key=}
}

os:export::Develop::Env() {
#    export http_proxy=http://proxy.asiainfo.com:8080
    export GITHUB_REDIRECT_URL=http://oauth2-oauth.app.asiainfodata.com/v1/github-redirect
    export GITHUB_CLIENT_ID=2369ed831a59847924b4
    export GITHUB_CLIENT_SECRET=510bb29970fcd684d0e7136a5947f92710332c98
}

#export DF_ENV_OAUTH_DEVELOP=true

Dev=`os::getEnv DF_ENV_OAUTH_DEVELOP`

if [ "$Dev" == "true" ]||[ "$Dev" == "" ];then
    os:export::Develop::Env
fi

export ETCD_HTTP_ADDR=http://etcdsystem.servicebroker.dataos.io
export ETCD_HTTP_PORT=2379
export ETCD_USER=asiainfoLDP
export ETCD_PASSWORD=6ED9BA74-75FD-4D1B-8916-842CB936AC1A

./datafactory_oauth2