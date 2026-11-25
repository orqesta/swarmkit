variable "IMAGE_NAME" {
  default = "orqesta/swarmkit"
}

target "dev" {
  tags = [IMAGE_NAME]
  output = ["type=docker"]
}

group "default" {
  targets = ["validate"]
}

group "validate" {
  targets = ["lint", "module-validate", "generate-validate"]
}

group "lint" {
  targets = ["golangci-lint", "fmt-proto"]
}

target "golangci-lint" {
  target = "lint"
  output = ["type=cacheonly"]
}

target "fmt-proto" {
  target = "fmt-proto"
  output = ["type=cacheonly"]
}

target "module-validate" {
  target = "module-validate"
  output = ["type=cacheonly"]
}

target "generate-validate" {
  target = "generate-validate"
  output = ["type=cacheonly"]
}

target "generate" {
  target = "generate-update"
  output = ["."]
}
