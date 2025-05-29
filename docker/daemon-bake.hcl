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
  output = [{ "type": "docker" }]
}

target "smartnode" {
  name = "smartnode-${arch}"
  dockerfile = "docker/rocketpool-dockerfile"
  args = {
    BUILDPLATFORM = "linux/amd64"
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
  platforms = ["linux/${arch}"]
  output = [{ "type": "docker" }]
}
