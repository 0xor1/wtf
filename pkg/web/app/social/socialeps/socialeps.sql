{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import "github.com/0xor1/tlbx/pkg/web/app" %}
{% import sqlh "github.com/0xor1/tlbx/pkg/web/app/sql" %}

{%- func qryInsert(args *sqlh.Args, social *social.Socials) -%}
{%- collapsespace -%}
INSERT INTO socials(
    id,
    handle,
    alias,
    hasAvatar
)
VALUES (
    ?,
    ?,
    ?,
    ?
)
{%- code 
    *args = *sqlh.NewArgs(4) 
    args.Append(
    social.ID,
    social.Handle,
    social.Alias,
    social.HasAvatar,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qrySelect(args *sqlh.Args, ids IDs, handlePrefix string, limit uint16) -%}
{%- collapsespace -%}
{%- code 
    limit = sqlh.Limit100(limit)
    app.BadReqIf(len(ids) > 100, "max ids to query is 100")
    app.BadReqIf(handlePrefix != "" && StrLen(handlePrefix) < 3, "min handlePrefix len is 3")
    app.BadReqIf(len(ids) == 0 && handlePrefix == "", "no query parameters provided please")
    *args = *sqlh.NewArgs(len(ids) + 1) 
-%}
SELECT id,
    handle,
    alias,
    hasAvatar
FROM socials
WHERE 
{% switch true %}
    {%- case len(ids) > 0 -%}
        id IN (?{%- for _ := range ids -%},?{%- endfor -%})
        {%- code
            *args = *sqlh.NewArgs(len(ids)) 
            args.Append(ids.ToIs()...)
        -%}
    {%- case handlePrefix -%}
        handle LIKE ?
        {%- code
            *args = *sqlh.NewArgs(2) 
            args.Append(sqlh.LikePrefix(handlePrefix), limit)
        -%}
{% endswitch%}
{%- endcollapsespace -%}
{%- endfunc -%}