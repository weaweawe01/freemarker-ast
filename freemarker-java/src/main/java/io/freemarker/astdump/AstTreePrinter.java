package io.freemarker.astdump;

import freemarker.core.TemplateElement;
import freemarker.core.TemplateObject;
import freemarker.template.Template;

import java.lang.reflect.Array;
import java.lang.reflect.Method;
import java.util.Collection;
import java.util.Enumeration;

/**
 * Prints AST in a format close to FreeMarker test resources (*.ast), for example:
 *   #mixed_content  // f.c.MixedContent
 *       #assign  // f.c.Assignment
 *           - assignment target: "x"  // String
 */
public final class AstTreePrinter {
    private static final Method GET_NODE_TYPE_SYMBOL = accessTemplateObjectMethod("getNodeTypeSymbol");
    private static final Method GET_PARAMETER_COUNT = accessTemplateObjectMethod("getParameterCount");
    private static final Method GET_PARAMETER_VALUE = accessTemplateObjectMethod("getParameterValue", int.class);
    private static final Method GET_PARAMETER_ROLE = accessTemplateObjectMethod("getParameterRole", int.class);

    public String print(Template template) {
        StringBuilder out = new StringBuilder(1024);
        TemplateObject root = template.getRootTreeNode();
        printNode(root, 0, out);
        return out.toString();
    }

    private void printNode(TemplateObject node, int indent, StringBuilder out) {
        appendIndent(out, indent);
        out.append(nodeSymbol(node))
                .append("  // ")
                .append(compactClassName(node.getClass()))
                .append('\n');

        printParameters(node, indent + 4, out);

        if (node instanceof TemplateElement) {
            Enumeration<?> children = ((TemplateElement) node).children();
            while (children.hasMoreElements()) {
                Object child = children.nextElement();
                if (child instanceof TemplateObject) {
                    printNode((TemplateObject) child, indent + 4, out);
                } else {
                    printScalarParameter("child", child, indent + 4, out);
                }
            }
        }
    }

    private void printParameters(TemplateObject node, int indent, StringBuilder out) {
        int paramCount = ((Number) invokeReflective(GET_PARAMETER_COUNT, node)).intValue();
        for (int i = 0; i < paramCount; i++) {
            String role = String.valueOf(invokeReflective(GET_PARAMETER_ROLE, node, i));
            Object value = invokeReflective(GET_PARAMETER_VALUE, node, i);
            printParameter(role, value, indent, out);
        }
    }

    private void printParameter(String role, Object value, int indent, StringBuilder out) {
        if (value == null) {
            appendIndent(out, indent);
            out.append("- ").append(role).append(": null  // Null").append('\n');
            return;
        }

        if (value instanceof TemplateObject) {
            TemplateObject node = (TemplateObject) value;
            appendIndent(out, indent);
            out.append("- ").append(role).append(": ")
                    .append(nodeSymbol(node))
                    .append("  // ")
                    .append(compactClassName(node.getClass()))
                    .append('\n');
            printParameters(node, indent + 4, out);
            if (node instanceof TemplateElement) {
                Enumeration<?> children = ((TemplateElement) node).children();
                while (children.hasMoreElements()) {
                    Object child = children.nextElement();
                    if (child instanceof TemplateObject) {
                        printNode((TemplateObject) child, indent + 4, out);
                    } else {
                        printScalarParameter("child", child, indent + 4, out);
                    }
                }
            }
            return;
        }

        Class<?> c = value.getClass();
        if (c.isArray()) {
            int len = Array.getLength(value);
            for (int i = 0; i < len; i++) {
                printParameter(role, Array.get(value, i), indent, out);
            }
            if (len == 0) {
                printScalarParameter(role, "[]", indent, out);
            }
            return;
        }

        if (value instanceof Collection) {
            Collection<?> collection = (Collection<?>) value;
            if (collection.isEmpty()) {
                printScalarParameter(role, "[]", indent, out);
                return;
            }
            for (Object item : collection) {
                printParameter(role, item, indent, out);
            }
            return;
        }

        printScalarParameter(role, value, indent, out);
    }

    private void printScalarParameter(String role, Object value, int indent, StringBuilder out) {
        appendIndent(out, indent);
        out.append("- ").append(role).append(": ");
        out.append(scalarDisplay(value));
        out.append("  // ");
        out.append(scalarTypeName(value));
        out.append('\n');
    }

    private String scalarDisplay(Object value) {
        if (value == null) {
            return "null";
        }
        if (value instanceof Number || value instanceof Boolean || value instanceof Character || value instanceof CharSequence) {
            return quote(String.valueOf(value));
        }
        return quote(String.valueOf(value));
    }

    private String scalarTypeName(Object value) {
        if (value == null) {
            return "Null";
        }
        return value.getClass().getSimpleName();
    }

    private String nodeSymbol(TemplateObject node) {
        Object symbol = invokeReflective(GET_NODE_TYPE_SYMBOL, node);
        if (symbol == null) {
            return node.getClass().getSimpleName();
        }
        String s = String.valueOf(symbol);
        return s.isEmpty() ? node.getClass().getSimpleName() : s;
    }

    private String compactClassName(Class<?> c) {
        String name = c.getName();
        if (name.startsWith("freemarker.core.")) {
            return "f.c." + c.getSimpleName();
        }
        if (name.startsWith("freemarker.template.")) {
            return "f.t." + c.getSimpleName();
        }
        if (name.startsWith("freemarker.ext.")) {
            return "f.e." + c.getSimpleName();
        }
        return name;
    }

    private void appendIndent(StringBuilder out, int indent) {
        for (int i = 0; i < indent; i++) {
            out.append(' ');
        }
    }

    private String quote(String s) {
        StringBuilder out = new StringBuilder(s.length() + 8);
        out.append('"');
        for (int i = 0; i < s.length(); i++) {
            char ch = s.charAt(i);
            switch (ch) {
                case '\\':
                    out.append("\\\\");
                    break;
                case '"':
                    out.append("\\\"");
                    break;
                case '\n':
                    out.append("\\n");
                    break;
                case '\r':
                    out.append("\\r");
                    break;
                case '\t':
                    out.append("\\t");
                    break;
                default:
                    out.append(ch);
            }
        }
        out.append('"');
        return out.toString();
    }

    private static Method accessTemplateObjectMethod(String name, Class<?>... parameterTypes) {
        try {
            Method method = TemplateObject.class.getDeclaredMethod(name, parameterTypes);
            method.setAccessible(true);
            return method;
        } catch (NoSuchMethodException e) {
            throw new IllegalStateException("TemplateObject method not found: " + name, e);
        }
    }

    private static Object invokeReflective(Method method, Object target, Object... args) {
        try {
            return method.invoke(target, args);
        } catch (Exception e) {
            throw new IllegalStateException("Failed to call method: " + method.getName(), e);
        }
    }
}
