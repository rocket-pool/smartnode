variable "VERSION" {
  default = "$VERSION"
}

variable "GIT_BRANCH" {
  default = "$GIT_BRANCH"
}

variable "GIT_COMMIT" {
  default = "$GIT_COMMIT"
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
  labels = {
    "org.opencontainers.image.ref.name" = "${GIT_BRANCH}"
    "org.opencontainers.image.revision" = "${GIT_COMMIT}"
    "org.opencontainers.image.source" = "https://github.com/rocket-pool/smartnode"
  }
  tags = [
    "rocketpool/smartnode:${VERSION}",
  ]
  target = "smartnode"
  platforms = ["linux/amd64", "linux/arm64"]
  output = [{ "type": "docker" }]
}
