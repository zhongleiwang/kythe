load("//tools:build_rules/shims.bzl", "go_test", "go_library")

package(default_visibility = ["//kythe:default_visibility"])

go_library(
    name = "identifiers",
    srcs = ["identifiers.go"],
    deps = [
        "//kythe/go/services/web",
        "//kythe/go/services/xrefs",
        "//kythe/go/storage/table",
        "//kythe/go/util/kytheuri",
        "//kythe/go/util/schema/tickets",
        "//kythe/proto:common_go_proto",
        "//kythe/proto:identifier_go_proto",
        "//kythe/proto:internal_go_proto",
        "//kythe/proto:serving_go_proto",
        "@com_github_golang_protobuf//proto:go_default_library",
    ],
)

go_test(
    name = "identifiers_test",
    size = "small",
    srcs = ["identifiers_test.go"],
    library = "identifiers",
    visibility = ["//visibility:private"],
    deps = [
        "//kythe/go/test/testutil",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@go_x_text//:encoding",
        "@go_x_text//:encoding/unicode",
        "@go_x_text//:transform",
    ],
)
