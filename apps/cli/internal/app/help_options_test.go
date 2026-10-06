package app

import "testing"

func TestHelpTokenizationUsesRegisteredFlagArity(t *testing.T) {
	for _, tc := range []struct {
		args      []string
		requested bool
		path      string
	}{
		{[]string{"--title", "--help", "assets", "create"}, false, ""},
		{[]string{"assets", "show", "--", "--help"}, false, ""},
		{[]string{"assets", "show", "--help=false"}, false, ""},
		{[]string{"--title=--help", "assets", "create", "--help=true"}, true, "assets create"},
		{[]string{"--help", "workflows", "revisions", "show"}, true, "workflows revisions show"},
	} {
		o, requested, err := ParseHelp(tc.args)
		if err != nil || requested != tc.requested {
			t.Fatalf("%v: %v %v", tc.args, requested, err)
		}
		if requested {
			path := ""
			for _, part := range o.Command {
				if path != "" {
					path += " "
				}
				path += part
			}
			if path != tc.path {
				t.Fatalf("path %q", path)
			}
		}
	}
}
