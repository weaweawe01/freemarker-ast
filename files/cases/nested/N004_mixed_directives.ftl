<#assign total = 0><#list items as item><#if item.price??><#assign total += item.price>${item.name}:${item.price};</#if></#list>Total=${total}
