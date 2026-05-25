// SPDX-License-Identifier: Apache-2.0

package helpers_test

import (
	"testing"
	"time"

	vstar "hop.top/vstar"
	"hop.top/vstar/helpers"
)

func TestRelatedTo_emptyWhenAbsent(t *testing.T) {
	c := vstar.Component{Type: vstar.CompTodo}
	if got := helpers.RelatedTo(c); len(got) != 0 {
		t.Errorf("RelatedTo = %v, want empty", got)
	}
}

func TestAddRelatedTo_appendsWithRELTYPE(t *testing.T) {
	c, _ := helpers.NewTodo("child-uid", time.Now().UTC())
	pre, _ := c.Get("X-VSTAR-HASH")
	helpers.AddRelatedTo(&c, "parent-uid", "PARENT")
	post, _ := c.Get("X-VSTAR-HASH")
	if pre.Value == post.Value {
		t.Errorf("X-VSTAR-HASH unchanged")
	}
	got := helpers.RelatedTo(c)
	if len(got) != 1 {
		t.Fatalf("RelatedTo len = %d, want 1", len(got))
	}
	if got[0].UID != "parent-uid" || got[0].RelType != "PARENT" {
		t.Errorf("ref = %+v, want {parent-uid PARENT}", got[0])
	}
}

func TestAddRelatedTo_multipleRoundTrip(t *testing.T) {
	c, _ := helpers.NewTodo("u", time.Now().UTC())
	helpers.AddRelatedTo(&c, "p", "PARENT")
	helpers.AddRelatedTo(&c, "s1", "SIBLING")
	helpers.AddRelatedTo(&c, "s2", "SIBLING")
	helpers.AddRelatedTo(&c, "k", "CHILD")
	got := helpers.RelatedTo(c)
	if len(got) != 4 {
		t.Fatalf("RelatedTo len = %d, want 4: %+v", len(got), got)
	}
	wantPairs := []helpers.RelatedRef{
		{UID: "p", RelType: "PARENT"},
		{UID: "s1", RelType: "SIBLING"},
		{UID: "s2", RelType: "SIBLING"},
		{UID: "k", RelType: "CHILD"},
	}
	for i, w := range wantPairs {
		if got[i] != w {
			t.Errorf("ref[%d] = %+v, want %+v", i, got[i], w)
		}
	}
}

func TestAddRelatedTo_emptyRelTypeOmitsParam(t *testing.T) {
	c, _ := helpers.NewTodo("u", time.Now().UTC())
	helpers.AddRelatedTo(&c, "parent-uid", "")
	got := helpers.RelatedTo(c)
	if len(got) != 1 {
		t.Fatalf("RelatedTo len = %d, want 1", len(got))
	}
	// RFC 5545 §3.2.15 default RELTYPE is PARENT when omitted.
	if got[0].UID != "parent-uid" || got[0].RelType != "PARENT" {
		t.Errorf("ref = %+v, want {parent-uid PARENT (default)}", got[0])
	}
}

func TestAddRelatedTo_nilIsNoOp(t *testing.T) {
	helpers.AddRelatedTo(nil, "p", "PARENT")
}

func TestAddRelatedTo_emptyUIDIsNoOp(t *testing.T) {
	c, _ := helpers.NewTodo("u", time.Now().UTC())
	helpers.AddRelatedTo(&c, "", "PARENT")
	if got := helpers.RelatedTo(c); len(got) != 0 {
		t.Errorf("RelatedTo len = %d, want 0 (empty UID skipped)", len(got))
	}
}
