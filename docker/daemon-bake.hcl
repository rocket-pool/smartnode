variable "VERSION" {
  default = "$VERSION"
}

group "default" {
  targets = ["builder", "daemon"]
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

target "daemon" {
  name = "daemon-${arch}"
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
  target = "daemon"
  platform = "linux/${arch}"
}
