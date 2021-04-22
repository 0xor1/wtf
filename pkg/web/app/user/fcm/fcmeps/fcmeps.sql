{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import sqlh "github.com/0xor1/tlbx/pkg/web/app/sql" %}

{%- func qryDistinctTokens(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
SELECT DISTINCT token 
FROM fcmTokens 
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(
    me,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryDelete(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
DELETE FROM fcmTokens
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(
    me,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}