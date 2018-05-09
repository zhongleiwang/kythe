#!/bin/bash -e
#
# Copyright 2015 Google Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Utility script to run the Kythe toolset over the Bazel repository.  See the usage() function below
# for more details.

KYTHE_RELEASE=/opt/kythe
BAZEL_ROOT="$PWD"

GRAPHSTORE=./gs.bazel
SERVING_TABLE=./serving.bazel
SERVING_ADDR=localhost:8889

EXTRACT=
ANALYZE=
POSTPROCESS=
SERVE=

cd "$BAZEL_ROOT"

echo "Analyzing compilations..."
echo "GraphStore: $GRAPHSTORE"

rm -rf "$GRAPHSTORE"

find -L "$BAZEL_ROOT"/bazel-out/ -name '*java*.kindex' | \
  parallel -L1 "java -Xbootclasspath/p:/opt/kythe/indexers/java_indexer.jar com.google.devtools.kythe.analyzers.java.JavaIndexer --ignore_empty_kindex" | \
  "$KYTHE_RELEASE"/tools/dedup_stream | \
  "$KYTHE_RELEASE"/tools/write_entries --graphstore "$GRAPHSTORE"
cd "$ROOT"

find -L "$BAZEL_ROOT"/bazel-out/ -name '*c++*.kindex' | \
  parallel -L1 "/opt/kythe/indexers/cxx_indexer" | \
  "$KYTHE_RELEASE"/tools/dedup_stream | \
  "$KYTHE_RELEASE"/tools/write_entries --graphstore "$GRAPHSTORE"
cd "$ROOT"

echo "Post-processing GraphStore..."
echo "Serving table: $SERVING_TABLE"

rm -rf "$SERVING_TABLE"

"$KYTHE_RELEASE"/tools/write_tables --graphstore "$GRAPHSTORE" --out "$SERVING_TABLE"

echo "Launching Kythe server"
echo "Address: http://$SERVING_ADDR"

"$KYTHE_RELEASE"/tools/http_server \
  --listen "$SERVING_ADDR" \
  --public_resources "$KYTHE_RELEASE"/web/ui \
  --serving_table "$SERVING_TABLE"

