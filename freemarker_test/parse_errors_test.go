package freemarker_test

import (
	"testing"
	"time"

	"github.com/weaweawe01/freemarker-ast/internal/parser"
)

// mustFailParse asserts that parser.Parse returns a non-nil error for the given
// template source.
func mustFailParse(t *testing.T, name, src string) {
	t.Helper()
	_, err := parser.Parse(src)
	if err == nil {
		t.Errorf("%s: expected parse error but got nil for input: %q", name, src)
	}
}

// mustFailParseWithin asserts parse fails and does not hang beyond timeout.
func mustFailParseWithin(t *testing.T, name, src string, timeout time.Duration) {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		_, err := parser.Parse(src)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Errorf("%s: expected parse error but got nil for input: %q", name, src)
		}
	case <-time.After(timeout):
		t.Fatalf("%s: parse timeout after %s, possible infinite loop; input=%q", name, timeout, src)
	}
}

// TestParseErrors_UnclosedDirectives tests templates with unclosed directives.
func TestParseErrors_UnclosedDirectives(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 1-10: unclosed block directives
		{"unclosed_if", `<#if true>`},
		{"unclosed_list", `<#list items as item>`},
		{"unclosed_macro", `<#macro foo>`},
		{"unclosed_function", `<#function foo()>`},
		{"unclosed_switch", `<#switch x>`},
		{"unclosed_attempt", `<#attempt>`},
		{"unclosed_autoesc", `<#autoesc>`},
		{"unclosed_noautoesc", `<#noautoesc>`},
		{"unclosed_outputformat", `<#outputformat "HTML">`},
		{"unclosed_nested_if", `<#if true><#if false></#if>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrors_MissingExpressions tests directives missing required expressions.
func TestParseErrors_MissingExpressions(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 11-20: missing expressions
		{"if_no_condition", `<#if></#if>`},
		{"if_empty_space_condition", `<#if ></#if>`},
		{"list_no_source", `<#list></#list>`},
		{"assign_no_value", `<#assign x>`},
		{"assign_no_name", `<#assign = 1>`},
		{"switch_no_value", `<#switch><#default></#switch>`},
		{"outputformat_no_value", `<#outputformat></#outputformat>`},
		{"global_no_value", `<#global x>`},
		{"local_no_value", `<#local x>`},
		{"assign_eq_missing_rhs", `<#assign x =>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrors_InvalidExpressions tests malformed expressions.
func TestParseErrors_InvalidExpressions(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 21-35: invalid expressions
		{"unclosed_interpolation", `${`},
		{"unclosed_interpolation_expr", `${x +`},
		{"empty_interpolation", `${}`},
		{"unclosed_paren", `${(x}`},
		{"unclosed_bracket", `${x[0}`},
		{"double_operator", `${x ++ y}`},
		{"missing_rhs_add", `${x +}`},
		{"missing_rhs_mul", `${x *}`},
		{"missing_rhs_and", `${x &&}`},
		{"missing_rhs_or", `${x ||}`},
		{"missing_rhs_eq", `${x ==}`},
		{"invalid_number_literal", `${12.34.56}`},
		{"unclosed_string_double", `${"hello}`},
		{"unclosed_string_single", `${'hello}`},
		{"trailing_dot", `${x.}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrors_InvalidBuiltins tests malformed built-in usage.
func TestParseErrors_InvalidBuiltins(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 36-45: invalid built-in usage
		{"builtin_no_name", `${x?}`},
		{"builtin_unclosed_args", `${x?upper_case(}`},
		{"builtin_invalid_name_number", `${x?123}`},
		{"builtin_chain_broken", `${x?.}`},
		{"builtin_starts_with_dot", `${x?.foo}`},
		{"builtin_missing_close_paren", `${x?string(}`},
		{"builtin_unclosed_paren_arg", `${x?string("a"`},
		{"builtin_question_at_eof", `${x?`},
		{"builtin_nested_unclosed", `${x?string(y?}`},
		{"builtin_incomplete_chain", `${x?size?}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrors_InvalidDirectiveSyntax tests directives with wrong syntax.
func TestParseErrors_InvalidDirectiveSyntax(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 46-60: invalid directive syntax
		{"else_without_if", `<#else></#if>`},
		{"elseif_without_if", `<#elseif true></#if>`},
		{"items_without_list", `<#items as i></#items>`},
		{"sep_without_list", `<#sep>, </#sep>`},
		{"recover_without_attempt", `<#recover></#attempt>`},
		{"case_without_switch", `<#case 1></#case>`},
		{"default_without_switch", `<#default></#default>`},
		{"close_tag_mismatch", `<#if true></#list>`},
		{"double_else_if", `<#if true><#else><#else></#if>`},
		{"macro_no_name", `<#macro></#macro>`},
		{"function_no_name", `<#function></#function>`},
		{"list_no_as_keyword", `<#list items item></#list>`},
		{"elseif_after_else", `<#if true><#else><#elseif false></#if>`},
		{"switch_unclosed_case", `<#switch x><#case 1>`},
		{"if_unclosed_nested_else", `<#if true><#if false><#else></#if>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrors_InvalidTags tests malformed tag syntax.
func TestParseErrors_InvalidTags(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 61-70: bad tag syntax
		{"unclosed_tag_if", `<#if true`},
		{"unclosed_tag_assign", `<#assign x = 1`},
		{"unclosed_tag_list", `<#list items as i`},
		{"unclosed_tag_macro", `<#macro foo`},
		{"tag_with_garbage", `<#if true $$$></#if>`},
		{"hash_interpolation_unclosed", `#{x`},
		{"unclosed_unified_call", `<@foo`},
		{"unclosed_close_tag", `<#if true></#if`},
		{"unclosed_tag_switch", `<#switch x`},
		{"unclosed_tag_function", `<#function foo(`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrors_InvalidLiterals tests malformed literal values.
func TestParseErrors_InvalidLiterals(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 71-80: invalid literals
		{"unclosed_hash_literal", `${{"a": 1}`},
		{"hash_missing_colon", `${{"a" 1}}`},
		{"hash_missing_value", `${{"a":}}`},
		{"hash_trailing_comma", `${{"a": 1,}}`},
		{"list_literal_unclosed", `${[1, 2}`},
		{"list_literal_trailing_comma", `${[1, 2,]}`},
		{"raw_string_unclosed", `${r"hello}`},
		{"unclosed_hash_nested", `${{"a": {"b": 1}}`},
		{"list_literal_double_comma", `${[1,,2]}`},
		{"hash_no_key", `${{: 1}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrors_MethodCalls tests malformed method call syntax.
func TestParseErrors_MethodCalls(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 81-90: method call errors
		{"call_unclosed_paren", `${foo(1, 2}`},
		{"call_double_comma", `${foo(1,,2)}`},
		{"call_leading_comma", `${foo(,1)}`},
		{"dynamic_key_unclosed", `${foo[0}`},
		{"dynamic_key_empty", `${foo[]}`},
		{"chained_unclosed", `${foo.bar(1.}`},
		{"assign_call_unclosed", `<#assign x = foo(1, 2>`},
		{"nested_call_unclosed", `${foo(bar(1)}`},
		{"missing_rhs_lt", `${foo <}`},
		{"bracket_missing_close", `${a[b[c]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrors_Miscellaneous tests other edge cases that should fail.
func TestParseErrors_Miscellaneous(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 91-100: miscellaneous errors
		{"lambda_no_arrow", `${x -> }`},
		{"lambda_no_body", `<#assign f = (x) ->>`},
		{"range_triple_dot", `${0...5}`},
		{"assign_plus_no_value", `<#assign x += >`},
		{"function_no_return_type", `<#function f(></#function>`},
		{"missing_rhs_gte", `${x >= }`},
		{"unary_not_no_operand", `${!}`},
		{"unary_minus_no_operand", `${-}`},
		{"double_close_brace", `${{}`},
		{"interpolation_in_directive_tag", `<#if ${x}></#if>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

func TestParseErrors_InvalidEscapedIdentifier_NoHang(t *testing.T) {
	src := `${ux;\J}`
	mustFailParseWithin(t, "invalid_escaped_identifier_no_hang", src, 300*time.Millisecond)
}
