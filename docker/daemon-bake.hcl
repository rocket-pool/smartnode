variable "VERSION" {
  default = "$VERSION"
}

group "default" {
  targets = ["builder", "smartnode"]
}

target "builder" {
  dockerfile = "docker/rocketpool-dockerfile"
  tags = [ 
    "rocketpool/smartnode-builder:${VERSION}",
    "rocketpool/smartnode-builder:local"
  ]
  target = "smartnode_dependencies"
  platforms = [ "linux/amd64" ]
}

target "smartnode" {
  name = "smartnode-${arch}"
  dockerfile = "docker/rocketpool-dockerfile"
  args = {
    VERSION = "${VERSION}"
  }
  tags = [
    "rocketpool/smartnode:${VERSION}-${arch}",
    "localhost/rocketpool/smartnode:${VERSION}-${arch}"
  ]
  matrix = {
    arch = [ "amd64", "arm64" ]
  }
  target = "smartnode"
  platform = "linux/${arch}"
  output = [{ "type": "docker" }]
}
