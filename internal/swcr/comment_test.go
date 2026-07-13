package swcr

import "testing"

func processAll(cs *CommentStripper, input []string) []string {
	cs.Reset()
	var out []string
	for _, s := range input {
		if line, ok := cs.ProcessLine(s); ok {
			out = append(out, line)
		}
	}
	return out
}

func TestCommentStripper_LineComments(t *testing.T) {
	cs := NewCommentStripper([]string{"#", "//"}, nil)
	input := []string{
		`package main`,
		`// this is a comment`,
		`  // indented comment`,
		`# python comment`,
		`code // not a comment because // is mid-line`,
	}
	want := []string{
		`package main`,
		`code // not a comment because // is mid-line`,
	}
	got := processAll(cs, input)
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCommentStripper_BlankLines(t *testing.T) {
	cs := NewCommentStripper([]string{"#", "//"}, nil)
	input := []string{``, `   `, "\t\t", `code`}
	got := processAll(cs, input)
	if len(got) != 1 || got[0] != `code` {
		t.Errorf("got %v, want [code]", got)
	}
}

func TestCommentStripper_BlockCommentC(t *testing.T) {
	cs := NewCommentStripper([]string{"//"}, []string{"/*:*/"})
	input := []string{
		`package main`,
		`/* single-line block */`,
		`func main() {`,
		`  /* start of`,
		`     multi-line`,
		`     block */` + ` fmt.Println("hi")`,
		`}`,
	}
	want := []string{
		`package main`,
		`func main() {`,
		`fmt.Println("hi")`,
		`}`,
	}
	got := processAll(cs, input)
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCommentStripper_BlockCommentPython(t *testing.T) {
	cs := NewCommentStripper([]string{"#"}, []string{`""":"""`})
	input := []string{
		`def foo():`,
		`    """Single-line docstring."""`,
		`    return 42`,
		``,
		`def bar():`,
		`    """`,
		`    Multi-line docstring.`,
		`    """`,
		`    return 43`,
	}
	want := []string{
		`def foo():`,
		`    return 42`,
		`def bar():`,
		`    return 43`,
	}
	got := processAll(cs, input)
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCommentStripper_ResetBetweenFiles(t *testing.T) {
	cs := NewCommentStripper([]string{"//"}, []string{"/*:*/"})

	// File 1: opens block comment but never closes.
	_ = processAll(cs, []string{"/* unfinished"})
	if cs.inBlock == "" {
		t.Fatal("expected to be inside block comment after file 1")
	}

	// Reset simulates starting a new file.
	cs.Reset()
	if cs.inBlock != "" {
		t.Fatal("expected reset to clear block state")
	}

	// File 2 should start fresh.
	got := processAll(cs, []string{"code", "// comment"})
	if len(got) != 1 || got[0] != "code" {
		t.Errorf("got %v, want [code]", got)
	}
}

func TestCommentStripper_HTMLComment(t *testing.T) {
	cs := NewCommentStripper(nil, []string{"<!--:-->"})
	input := []string{
		`<html>`,
		`<!-- header -->`,
		`<body>`,
		`<!--`,
		`  multi-line`,
		`  comment`,
		`-->`,
		`<p>hello</p>`,
	}
	want := []string{
		`<html>`,
		`<body>`,
		`<p>hello</p>`,
	}
	got := processAll(cs, input)
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, got[i], want[i])
		}
	}
}
