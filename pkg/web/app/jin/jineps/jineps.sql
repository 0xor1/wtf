{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import "github.com/0xor1/tlbx/pkg/json" %}
{% import sqlh "github.com/0xor1/tlbx/pkg/web/app/sql" %}

{%- func qryInsert(args *sqlh.Args, me ID, val *json.Json) -%}
{%- collapsespace -%}
INSERT INTO jins(
    id,
    val
)
VALUES (
    ?,
    ?
)
{%- code 
    *args = *sqlh.NewArgs(2) 
    args.Append(
    me,
    val,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryDelete(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
Delete FROM jins
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(
    me,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qrySelect(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
SELECT val
FROM jins 
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(
    me,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}