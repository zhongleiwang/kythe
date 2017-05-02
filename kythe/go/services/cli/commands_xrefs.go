/*
 * Copyright 2017 Google Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"kythe.io/kythe/go/util/kytheuri"
	"kythe.io/kythe/go/util/markedsource"
	"kythe.io/kythe/go/util/schema/facts"

	xpb "kythe.io/kythe/proto/xref_proto"
)

type xrefsCommand struct {
	nodeFilters                            string
	pageToken                              string
	pageSize                               int
	defKind, declKind, refKind, callerKind string
	relatedNodes, nodeDefinitions          bool
}

func (xrefsCommand) Name() string     { return "xrefs" }
func (xrefsCommand) Synopsis() string { return "retrieve cross-references for the given node" }
func (xrefsCommand) Usage() string    { return "" }
func (c *xrefsCommand) SetFlags(flag *flag.FlagSet) {
	flag.StringVar(&c.defKind, "definitions", "all", "Kind of definitions to return (kinds: all, binding, full, or none)")
	flag.StringVar(&c.declKind, "declarations", "all", "Kind of declarations to return (kinds: all or none)")
	flag.StringVar(&c.refKind, "references", "noncall", "Kind of references to return (kinds: all, noncall, call, or none)")
	flag.StringVar(&c.callerKind, "callers", "none", "Kind of callers to return (kinds: direct, overrides, or none)")
	flag.BoolVar(&c.relatedNodes, "related_nodes", false, "Whether to request related nodes")
	flag.StringVar(&c.nodeFilters, "filters", "", "Comma-separated list of additional fact filters to use when requesting related nodes")
	flag.BoolVar(&c.nodeDefinitions, "node_definitions", false, "Whether to request definition locations for related nodes")

	flag.StringVar(&c.pageToken, "page_token", "", "CrossReferences page token")
	flag.IntVar(&c.pageSize, "page_size", 0, "Maximum number of cross-references returned (0 lets the service use a sensible default)")
}
func (c *xrefsCommand) Run(ctx context.Context, flag *flag.FlagSet, api API) error {
	req := &xpb.CrossReferencesRequest{
		Ticket:          flag.Args(),
		PageToken:       c.pageToken,
		PageSize:        int32(c.pageSize),
		NodeDefinitions: c.nodeDefinitions,
	}
	if c.relatedNodes {
		req.Filter = []string{facts.NodeKind, facts.Subkind}
		if c.nodeFilters != "" {
			req.Filter = append(req.Filter, strings.Split(c.nodeFilters, ",")...)
		}
	}
	switch c.defKind {
	case "all":
		req.DefinitionKind = xpb.CrossReferencesRequest_ALL_DEFINITIONS
	case "none":
		req.DefinitionKind = xpb.CrossReferencesRequest_NO_DEFINITIONS
	case "binding":
		req.DefinitionKind = xpb.CrossReferencesRequest_BINDING_DEFINITIONS
	case "full":
		req.DefinitionKind = xpb.CrossReferencesRequest_FULL_DEFINITIONS
	default:
		return fmt.Errorf("unknown definition kind: %q", c.defKind)
	}
	switch c.declKind {
	case "all":
		req.DeclarationKind = xpb.CrossReferencesRequest_ALL_DECLARATIONS
	case "none":
		req.DeclarationKind = xpb.CrossReferencesRequest_NO_DECLARATIONS
	default:
		return fmt.Errorf("unknown declaration kind: %q", c.declKind)
	}
	switch c.refKind {
	case "all":
		req.ReferenceKind = xpb.CrossReferencesRequest_ALL_REFERENCES
	case "noncall":
		req.ReferenceKind = xpb.CrossReferencesRequest_NON_CALL_REFERENCES
	case "call":
		req.ReferenceKind = xpb.CrossReferencesRequest_CALL_REFERENCES
	case "none":
		req.ReferenceKind = xpb.CrossReferencesRequest_NO_REFERENCES
	default:
		return fmt.Errorf("unknown reference kind: %q", c.refKind)
	}
	switch c.callerKind {
	case "direct":
		req.CallerKind = xpb.CrossReferencesRequest_DIRECT_CALLERS
	case "overrides":
		req.CallerKind = xpb.CrossReferencesRequest_OVERRIDE_CALLERS
	case "none":
		req.CallerKind = xpb.CrossReferencesRequest_NO_CALLERS
	default:
		return fmt.Errorf("unknown caller kind: %q", c.callerKind)
	}
	logRequest(req)
	reply, err := api.XRefService.CrossReferences(ctx, req)
	if err != nil {
		return err
	}
	if reply.NextPageToken != "" {
		defer log.Printf("Next page token: %s", reply.NextPageToken)
	}
	return c.displayXRefs(reply)
}

func (c *xrefsCommand) displayXRefs(reply *xpb.CrossReferencesReply) error {
	if *displayJSON {
		return json.NewEncoder(os.Stdout).Encode(reply)
	}

	for _, xr := range reply.CrossReferences {
		if _, err := fmt.Fprintln(out, "Cross-References for ", showSignature(xr.MarkedSource), xr.Ticket); err != nil {
			return err
		}
		if err := displayRelatedAnchors("Definitions", xr.Definition); err != nil {
			return err
		}
		if err := displayRelatedAnchors("Declarations", xr.Declaration); err != nil {
			return err
		}
		if err := displayRelatedAnchors("References", xr.Reference); err != nil {
			return err
		}
		if err := displayRelatedAnchors("Callers", xr.Caller); err != nil {
			return err
		}
		if len(xr.RelatedNode) > 0 {
			if _, err := fmt.Fprintln(out, "  Related Nodes:"); err != nil {
				return err
			}
			for _, n := range xr.RelatedNode {
				var nodeKind, subkind string
				if node, ok := reply.Nodes[n.Ticket]; ok {
					for name, value := range node.Facts {
						switch name {
						case facts.NodeKind:
							nodeKind = string(value)
						case facts.Subkind:
							subkind = string(value)
						}
					}
				}
				if nodeKind == "" {
					nodeKind = "UNKNOWN"
				} else if subkind != "" {
					nodeKind += "/" + subkind
				}
				if _, err := fmt.Fprintf(out, "    %s %s [%s]\n", n.Ticket, n.RelationKind, nodeKind); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func displayRelatedAnchors(kind string, anchors []*xpb.CrossReferencesReply_RelatedAnchor) error {
	if len(anchors) > 0 {
		if _, err := fmt.Fprintf(out, "  %s:\n", kind); err != nil {
			return err
		}

		for _, a := range anchors {
			pURI, err := kytheuri.Parse(a.Anchor.Parent)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "    %s\t%s\t[%d:%d-%d:%d)\n      %q\n",
				pURI.Path, showSignature(a.MarkedSource),
				a.Anchor.Span.Start.LineNumber, a.Anchor.Span.Start.ColumnOffset,
				a.Anchor.Span.End.LineNumber, a.Anchor.Span.End.ColumnOffset,
				string(a.Anchor.Snippet)); err != nil {
				return err
			}
			for _, site := range a.Site {
				if _, err := fmt.Fprintf(out, "      [%d:%d-%d-%d)\n        %q\n",
					site.Span.Start.LineNumber, site.Span.Start.ColumnOffset,
					site.Span.End.LineNumber, site.Span.End.ColumnOffset,
					string(site.Snippet)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func showSignature(signature *xpb.MarkedSource) string {
	if signature == nil {
		return "(nil)"
	}
	return markedsource.Render(signature)
}
