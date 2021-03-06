load("@io_bazel_rules_go//go:def.bzl", _go_binary = "go_binary", _go_library = "go_library", _go_test = "go_test")

# Go importpath prefix shared by all Kythe libraries
go_prefix = "kythe.io/"

def _infer_importpath(name):
  basename = native.package_name().split('/')[-1]
  importpath = go_prefix + native.package_name()
  if basename == name:
    return importpath
  return importpath + '/' + name

def go_binary(name, importpath=None, **kwargs):
  """This macro wraps the go_binary rule provided by the Bazel Go rules to
  automatically infer the binary's importpath.  It is otherwise equivalent in
  function to a go_binary. """
  if importpath == None:
    importpath = _infer_importpath(name)
  _go_binary(
    name = name,
    importpath = importpath,
    **kwargs
  )

def go_library(name, importpath=None, **kwargs):
  """This macro wraps the go_library rule provided by the Bazel Go rules to
  automatically infer the library's importpath.  It is otherwise equivalent in
  function to a go_library. """
  if importpath == None:
    importpath = _infer_importpath(name)
  _go_library(
    name = name,
    importpath = importpath,
    **kwargs
  )

def go_test(name, library=None, **kwargs):
  """This macro wraps the go_test rule provided by the Bazel Go rules
  to silence a deprecation warning for use of the "library" attribute.
  It is otherwise equivalent in function to a go_test.
  """
  if not library:
    fail('Missing required "library" attribute')

  _go_test(
      name = name,
      embed = [library],
      **kwargs
  )
