package(default_visibility = ["//visibility:public"])

licenses(["notice"])  # Apache 2.0

action_listener(
    name = "cxx_extract_kindex",
    extra_actions = [":cxx_extra_action"],
    mnemonics = ["CppCompile"],
)

extra_action(
    name = "cxx_extra_action",
    cmd = "$(location :cxx_extractor) $(EXTRA_ACTION_FILE) $(output $(ACTION_ID).c++.kindex) $(location :kythe_config)",
    data = [":kythe_config"],
    out_templates = ["$(ACTION_ID).c++.kindex"],
    tools = [":cxx_extractor"],
)

action_listener(
    name = "java_extract_kindex",
    extra_actions = [":java_extra_action"],
    mnemonics = ["Javac"],
)

extra_action(
    name = "java_extra_action",
    cmd = ("java -Xbootclasspath/p:/opt/kythe/extractors/bazel_java_extractor.jar " +
    "com.google.devtools.kythe.extractors.java.bazel.JavaExtractor " +
    "$(EXTRA_ACTION_FILE) $(output $(ACTION_ID).java.kindex) $(location :kythe_config)"),
    data = [":kythe_config"],
    out_templates = ["$(ACTION_ID).java.kindex"],
    tools = [":java_extractor"],
)

filegroup(
    name = "java_extractor",
    srcs = ["bazel_java_extractor.jar"],
)

filegroup(
    name = "cxx_extractor",
    srcs = ["bazel_cxx_extractor"],
)

filegroup(
    name = "kythe_config",
    srcs = ["config.json"],
)
        
