package freemarker_test

import (
	"testing"

	"github.com/weaweawe01/freemarker-ast/internal/parser"
)

// mustSucceedParse asserts that parser.Parse returns no error for the given template source.
func mustSucceedParse(t *testing.T, name, src string) {
	t.Helper()
	_, err := parser.Parse(src)
	if err != nil {
		t.Errorf("%s: expected parse success but got error: %v\nInput: %q", name, err, src)
	}
}

// TestParseSuccess_BasicDirectives tests basic directive usage.
func TestParseSuccess_BasicDirectives(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 1-10
		{"simple_if", `<#if true>hello</#if>`},
		{"if_else", `<#if x>a<#else>b</#if>`},
		{"if_elseif_else", `<#if x>a<#elseif y>b<#else>c</#if>`},
		{"simple_list", `<#list items as item>${item}</#list>`},
		{"list_with_sep", `<#list items as item>${item}<#sep>, </#list>`},
		{"list_with_items", `<#list items><#items as item>${item}</#items></#list>`},
		{"simple_assign", `<#assign x = 1>`},
		{"assign_multiple", `<#assign x = 1 y = 2 z = 3>`},
		{"simple_global", `<#global g = "hello">`},
		{"simple_local", `<#function f()><#local x = 1><#return x></#function>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_MacrosAndFunctions tests macro and function definitions.
func TestParseSuccess_MacrosAndFunctions(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 11-20
		{"simple_macro", `<#macro greet name>Hello ${name}</#macro>`},
		{"macro_no_params", `<#macro footer>Copyright 2024</#macro>`},
		{"macro_default_param", `<#macro greet name="World">Hello ${name}</#macro>`},
		{"macro_nested", `<#macro wrapper><#nested></#macro>`},
		{"macro_call", `<@greet name="Alice" />`},
		{"macro_call_body", `<@wrapper>content</@wrapper>`},
		{"simple_function", `<#function add(a, b)><#return a + b></#function>`},
		{"function_default_param", `<#function inc(x, step=1)><#return x + step></#function>`},
		{"function_call_in_interp", `<#function f(x)><#return x * 2></#function>${f(5)}`},
		{"macro_catch_all", `<#macro m a b...></#macro>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_Expressions tests various valid expressions.
func TestParseSuccess_Expressions(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 21-30
		{"arithmetic_add", `${a + b}`},
		{"arithmetic_sub", `${a - b}`},
		{"arithmetic_mul", `${a * b}`},
		{"arithmetic_div", `${a / b}`},
		{"arithmetic_mod", `${a % b}`},
		{"unary_minus", `${-x}`},
		{"unary_not", `${!x}`},
		{"comparison_eq", `<#if a == b>eq</#if>`},
		{"comparison_neq", `<#if a != b>neq</#if>`},
		{"comparison_lt_gt", `<#if a < b && b > c>yes</#if>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_Literals tests various literal values.
func TestParseSuccess_Literals(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 31-40
		{"string_double", `${"hello world"}`},
		{"string_single", `${'hello world'}`},
		{"string_escape", `${"line1\nline2\ttab"}`},
		{"empty_sequence", `${[]}`},
		{"number_int", `${42}`},
		{"number_decimal", `${3.14}`},
		{"boolean_true", `<#if true>yes</#if>`},
		{"boolean_false", `<#if false>no</#if>`},
		{"sequence_literal", `<#assign xs = [1, 2, 3]>`},
		{"hash_literal", `<#assign h = {"name": "Alice", "age": 30}>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_Builtins tests built-in function usage.
func TestParseSuccess_Builtins(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 41-50
		{"builtin_upper_case", `${name?upper_case}`},
		{"builtin_lower_case", `${name?lower_case}`},
		{"builtin_size", `${items?size}`},
		{"builtin_length", `${name?length}`},
		{"builtin_has_content", `<#if x?has_content>yes</#if>`},
		{"builtin_default", `${x!""}`},
		{"builtin_exists", `<#if x??>${x}</#if>`},
		{"builtin_string", `${num?string("0.00")}`},
		{"builtin_chain", `${name?trim?upper_case?length}`},
		{"builtin_first_last", `${items?first} ${items?last}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_StringInterpolation tests string interpolation inside strings.
func TestParseSuccess_StringInterpolation(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 51-60
		{"interp_simple", `${"Hello ${name}"}`},
		{"interp_expr", `${"Result: ${a + b}"}`},
		{"interp_nested", `${"${a} and ${b}"}`},
		{"interp_builtin", `${"Name: ${name?upper_case}"}`},
		{"interp_in_assign", `<#assign msg = "Hello ${name}">`},
		{"interp_in_if", `<#if greeting == "Hello ${name}">match</#if>`},
		{"interp_with_default", `${"Value: ${x!0}"}`},
		{"interp_method_call", `${"Result: ${obj.method()}"}`},
		{"interp_dot_access", `${"Name: ${user.name}"}`},
		{"interp_arithmetic", `${"Sum: ${a + b * c}"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_SwitchCase tests valid switch/case structures.
func TestParseSuccess_SwitchCase(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 61-65
		{"switch_single_case", `<#switch x><#case 1>one</#switch>`},
		{"switch_multi_case", `<#switch x><#case 1>one<#case 2>two</#switch>`},
		{"switch_default", `<#switch x><#case 1>one<#default>other</#switch>`},
		{"switch_with_break", `<#switch x><#case 1>one<#break><#case 2>two</#switch>`},
		{"switch_expr_value", `<#switch x + y><#case a * b>match</#switch>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_AttemptRecover tests valid attempt/recover structures.
func TestParseSuccess_AttemptRecover(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 66-70
		{"attempt_recover", `<#attempt>${dangerous}<#recover>fallback</#attempt>`},
		{"attempt_with_directives", `<#attempt><#if true>${x}</#if><#recover>error</#attempt>`},
		{"attempt_nested", `<#attempt><#attempt>${x}<#recover>inner</#attempt><#recover>outer</#attempt>`},
		{"attempt_in_list", `<#list xs as x><#attempt>${x}<#recover>err</#attempt></#list>`},
		{"attempt_in_if", `<#if true><#attempt>try<#recover>catch</#attempt></#if>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_ComplexNesting tests deeply nested valid structures.
func TestParseSuccess_ComplexNesting(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 71-80
		{"if_in_list", `<#list items as item><#if item?has_content>${item}</#if></#list>`},
		{"list_in_if", `<#if show><#list items as item>${item}</#list></#if>`},
		{"nested_if", `<#if a><#if b><#if c>deep</#if></#if></#if>`},
		{"macro_with_if_list", `<#macro render items><#list items as item><#if item != "">${item}</#if></#list></#macro>`},
		{"assign_complex_expr", `<#assign result = items?filter(x -> x > 0)?size>`},
		{"list_with_index", `<#list items as item>${item?index}: ${item}<#sep>, </#list>`},
		{"if_with_assign", `<#if true><#assign x = 1><#assign y = 2>${x + y}</#if>`},
		{"function_with_logic", `<#function max(a, b)><#if a gt b><#return a><#else><#return b></#if></#function>`},
		{"list_nested_list", `<#list matrix as row><#list row as cell>${cell}</#list></#list>`},
		{"switch_in_list", `<#list items as item><#switch item.type><#case "a">A<#case "b">B<#default>?</#switch></#list>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_IncludeImport tests include and import directives.
func TestParseSuccess_IncludeImport(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 81-85
		{"include_simple", `<#include "header.ftl">`},
		{"include_with_parse", `<#include "data.txt" parse=false>`},
		{"import_simple", `<#import "lib.ftl" as lib>`},
		{"import_and_use", `<#import "utils.ftl" as u>${u.helper()}`},
		{"include_variable", `<#include path>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_SpecialDirectives tests autoesc, outputformat, compress, etc.
func TestParseSuccess_SpecialDirectives(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 86-90
		{"autoesc", `<#autoesc>${html}</#autoesc>`},
		{"noautoesc", `<#noautoesc>${raw}</#noautoesc>`},
		{"outputformat_html", `<#outputformat "HTML">${content}</#outputformat>`},
		{"compress", `<#compress>  lots   of   spaces  </#compress>`},
		{"noparse", `<#noparse>${not_evaluated}</#noparse>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}

// TestParseSuccess_ComplexExpressions tests complex but valid expressions.
func TestParseSuccess_ComplexExpressions(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// 91-100
		{"range_exclusive", `<#list 0..<10 as i>${i}</#list>`},
		{"range_inclusive", `<#list 0..10 as i>${i}</#list>`},
		{"default_operator", `${x!"default"}`},
		{"hash_access_dynamic", `${map[key]}`},
		{"method_call_chain", `${obj.method1().method2().value}`},
		{"lambda_filter", `<#assign filtered = items?filter(x -> x > 0)>`},
		{"lambda_map", `<#assign mapped = items?map(x -> x * 2)>`},
		{"complex_boolean", `<#if (a > b) && (c < d) || !(e == f)>yes</#if>`},
		{"nested_hash_access", `${config["section"]["key"]}`},
		{"assign_plus_eq", `<#assign count += 1>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustSucceedParse(t, tc.name, tc.src)
		})
	}
}
