<#list users as u><#if u.roles?has_content><#list u.roles as r>${u.name}:${r}</#list></#if><#-- ignored --></#list>
