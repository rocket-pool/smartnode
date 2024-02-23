#!/usr/bin/env bash
this_dir=$(cd $(dirname "$0"); pwd) # this script's directory


# see:
#   - https://github.com/prometheus/alertmanager?tab=readme-ov-file#api
#   - https://goswagger.io/generate/client.html (and https://goswagger.io/generate/requirements.html)

# command from https://goswagger.io/install.html#for-mac-and-linux-users
swagger_cmd="docker run --rm -it --user $(id -u):$(id -g) -e GOPATH=$(go env GOPATH):/go -v $HOME:$HOME -w $(pwd) quay.io/goswagger/swagger"

echo "Swagger version:"
$swagger_cmd version

# Using the v0.26 of alertmanager openapi spec: 
alertmanager_release_tag=release-0.26
open_api_spec_url=https://raw.githubusercontent.com/prometheus/alertmanager/${alertmanager_release_tag}/api/v2/openapi.yaml

[ -d "${this_dir}/alertmanager" ] || mkdir -p "${this_dir}/alertmanager"
echo "Generating client for alertmanager ${alertmanager_release_tag}..."
$swagger_cmd generate client -f $open_api_spec_url -A alertmanager -t "${this_dir}/alertmanager"
