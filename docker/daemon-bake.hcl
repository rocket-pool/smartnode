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
  ]
  target = "smartnode_dependencies"
  platforms = [ "linux/amd64" ]
  output = [{ "type": "docker" }]
}

target "smartnode" {
  dockerfile = "docker/rocketpool-dockerfile"
  args = {
    BUILDPLATFORM = "linux/amd64"
    VERSION = "${VERSION}"
  }
  tags = [
    "rocketpool/smartnode:${VERSION}",
    "localhost/rocketpool/smartnode:${VERSION}"
  ]
  target = "smartnode"
  platforms = ["linux/amd64", "linux/arm64"]
  output = [{ "type": "docker" }]
}
