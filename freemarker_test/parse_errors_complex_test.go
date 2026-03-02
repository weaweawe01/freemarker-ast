package freemarker_test

import "testing"

// TestParseErrorsComplex_NestedUnclosed tests deeply nested structures left unclosed.
func TestParseErrorsComplex_NestedUnclosed(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 1-10
		{"nested_if_list_unclosed", `<#if true><#list xs as x></#if>`},
		{"nested_if_if_if_unclosed", `<#if a><#if b><#if c></#if></#if>`},
		{"list_in_macro_unclosed", `<#macro m><#list xs as x></#macro>`},
		{"switch_in_if_unclosed", `<#if true><#switch x><#case 1>hello`},
		{"attempt_in_list_unclosed", `<#list xs as x><#attempt>bad`},
		{"nested_attempt_unclosed_in_fn", `<#function f()><#attempt></#function>`},
		{"if_else_list_unclosed", `<#if true><#else><#list xs as x></#if>`},
		{"nested_list_list_unclosed", `<#list xs as x><#list ys as y></#list>`},
		{"autoesc_in_if_unclosed", `<#if true><#autoesc></#if>`},
		{"outputformat_in_list_unclosed", `<#list xs as x><#outputformat "HTML"></#list>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_MalformedInterpolations tests complex malformed interpolations.
func TestParseErrorsComplex_MalformedInterpolations(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 11-20
		{"interp_unclosed_bracket_chain", `${a[0][1}`},
		{"double_nested_interp_unclosed", `${"a ${b + "c ${d"}"}`},
		{"interp_missing_close_nested", `${a + b + (c * d}`},
		{"interp_triple_plus", `${a +++ b}`},
		{"interp_unclosed_nested_paren", `${(a + (b + (c)}`},
		{"interp_double_star", `${a ** b}`},
		{"interp_colon_outside_hash", `${a : b}`},
		{"interp_percent_op", `${a % }`},
		{"interp_hash_sign_expr", `${#foo}`},
		{"interp_curly_in_interp", `${{a: 1}.b.}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_DirectiveMixups tests directives used in wrong contexts.
func TestParseErrorsComplex_DirectiveMixups(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 21-30
		{"else_in_switch", `<#switch x><#else></#switch>`},
		{"elseif_in_list", `<#list xs as x><#elseif true></#list>`},
		{"items_in_if", `<#if true><#items as x></#items></#if>`},
		{"recover_in_if", `<#if true><#recover></#if>`},
		{"case_in_if", `<#case 1></#if>`},
		{"list_close_with_if", `<#list xs as x></#if>`},
		{"function_close_with_macro", `<#function f()></#macro>`},
		{"macro_close_with_function", `<#macro m></#function>`},
		{"attempt_close_with_if", `<#attempt><#recover></#if>`},
		{"if_close_with_attempt", `<#if true></#attempt>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_MalformedAssignments tests complex assignment errors.
func TestParseErrorsComplex_MalformedAssignments(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 31-40
		{"assign_double_eq", `<#assign x == 1>`},
		{"assign_expr_as_name", `<#assign (x + 1) = 2>`},
		{"assign_unclosed_value_expr", `<#assign x = (1 + 2>`},
		{"assign_missing_eq_complex", `<#assign x 1 + 2>`},
		{"global_double_eq", `<#global x == 1>`},
		{"local_expr_name", `<#local (a) = 1>`},
		{"assign_nested_unclosed", `<#assign x = [1, 2>`},
		{"assign_hash_unclosed", `<#assign x = {"a": 1>`},
		{"assign_lambda_broken", `<#assign f = (x, y) ->>`},
		{"local_missing_eq_value", `<#local x 1>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_MalformedMacros tests complex macro/function definition errors.
func TestParseErrorsComplex_MalformedMacros(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 41-50
		{"macro_param_eq_no_default", `<#macro m a=></#macro>`},
		{"macro_param_unclosed_default", `<#macro m a=(1 + 2></#macro>`},
		{"macro_param_double_dots", `<#macro m a....></#macro>`},
		{"function_param_eq_no_default", `<#function f(a=)></#function>`},
		{"function_param_unclosed", `<#function f(a, b></#function>`},
		{"macro_number_as_name", `<#macro 123></#macro>`},
		{"function_number_name", `<#function 456()></#function>`},
		{"macro_dot_in_name", `<#macro a.b></#macro>`},
		{"macro_nested_unclosed_default", `<#macro m a=[1, 2></#macro>`},
		{"function_extra_close", `<#function f()></#function></#function>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_MalformedListDirectives tests complex list-related errors.
func TestParseErrorsComplex_MalformedListDirectives(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 51-60
		{"list_as_missing_var", `<#list xs as></#list>`},
		{"list_as_number", `<#list xs as 123></#list>`},
		{"list_double_as", `<#list xs as x as y></#list>`},
		{"list_unclosed_source_expr", `<#list (xs + ys as x></#list>`},
		{"list_no_close_with_sep", `<#list xs as x>${x}<#sep>, `},
		{"list_items_double_as", `<#list xs><#items as x as y></#items></#list>`},
		{"list_items_no_var", `<#list xs><#items as></#items></#list>`},
		{"list_nested_break_in_if", `<#list xs as x><#if true><#break></#list>`},
		{"list_source_unclosed_bracket", `<#list xs[0 as x></#list>`},
		{"list_source_unclosed_paren", `<#list (xs as x></#list>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_MalformedSwitchCase tests complex switch/case errors.
func TestParseErrorsComplex_MalformedSwitchCase(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 61-70
		{"switch_case_no_value", `<#switch x><#case></#switch>`},
		{"switch_case_unclosed_nested_list", `<#switch x><#case 1><#list xs as x></#switch>`},
		{"switch_case_unclosed_expr", `<#switch x><#case (1 + 2>a</#switch>`},
		{"switch_nested_unclosed_if", `<#switch x><#case 1><#if true></#switch>`},
		{"switch_value_unclosed", `<#switch (x + y><#case 1>a</#switch>`},
		{"switch_unclosed_nested_if", `<#switch x><#case 1><#if true>a</#switch>`},
		{"switch_empty_case_value", `<#switch x><#case >a</#switch>`},
		{"switch_unclosed_with_cases", `<#switch x><#case 1>a<#case 2>b`},
		{"switch_case_unclosed_string", `<#switch x><#case "abc></#switch>`},
		{"switch_nested_switch_unclosed", `<#switch x><#case 1><#switch y><#case 2>a</#switch>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_ComplexExpressions tests deeply nested or chained expression errors.
func TestParseErrorsComplex_ComplexExpressions(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 71-80
		{"deeply_nested_paren_unclosed", `${((((a + b) * c) - d)`},
		{"chained_builtin_unclosed", `${x?string?upper_case(}`},
		{"bracket_access_unclosed_nested", `${a[b + (c * d]}`},
		{"hash_concat_unclosed", `${{} + {"a": 1}`},
		{"seq_concat_unclosed", `${[] + [1, 2}`},
		{"complex_dot_chain_broken", `${a.b.c.d.}`},
		{"nested_dynamic_key_broken", `${a[b[c[d]]}`},
		{"method_chain_unclosed", `${a.b().c(d.}`},
		{"complex_arithmetic_unclosed", `${a + b * (c - d / (e + f)}`},
		{"mixed_ops_incomplete", `${a > b && c < }`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_StringAndRawErrors tests complex string literal errors.
func TestParseErrorsComplex_StringAndRawErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 81-90
		{"string_interp_nested_unclosed", `${"a${b + "c${d}"}}`},
		{"string_interp_missing_close", `${"hello ${name + " world"}`},
		{"string_concat_missing_rhs", `${"abc" + }`},
		{"string_concat_unclosed", `${"abc" + "def}`},
		{"assign_string_interp_broken", `<#assign x = "hello ${>`},
		{"if_condition_string_unclosed", `<#if x == "hello></#if>`},
		{"list_source_string_unclosed", `<#list "abc as x></#list>`},
		{"switch_string_unclosed", `<#switch "abc></#switch>`},
		{"raw_string_in_hash_unclosed", `${{r"key: "val"}}`},
		{"string_backslash_at_end", `${"hello\`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}

// TestParseErrorsComplex_MultiDirectiveErrors tests templates with multiple interacting errors.
func TestParseErrorsComplex_MultiDirectiveErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 91-100
		{"if_in_interp", `${<#if true>x</#if>}`},
		{"assign_directive_nested_broken", `<#assign x = 1 + >`},
		{"assign_with_directive_value", `<#assign x = <#if true>1</#if>>`},
		{"macro_call_unclosed_nested", `<@foo bar=(1 + 2 />`},
		{"unified_call_unclosed_param", `<@widget name="test>`},
		{"nested_directives_all_unclosed", `<#if true><#list xs as x><#attempt>`},
		{"double_close_different", `</#if></#list>`},
		{"directive_in_expr_context", `${1 + <#assign x = 2> + 3}`},
		{"function_call_with_directive", `${foo(<#if true>1</#if>)}`},
		{"hash_value_directive", `${{"a": <#if true>1</#if>}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustFailParse(t, tc.name, tc.src)
		})
	}
}
