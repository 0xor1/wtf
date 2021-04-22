{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import "github.com/0xor1/tlbx/pkg/web/app" %}
{% import sqlh "github.com/0xor1/tlbx/pkg/web/app/sql" %}
{% import "github.com/0xor1/tlbx/pkg/web/app/user/social" %}

{%- func qryInsert(args *sqlh.Args, social *social.Social) -%}
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

{%- func qryUpdate(args *sqlh.Args, social *social.Social) -%}
{%- collapsespace -%}
UPDATE socials SET
    handle=?,
    alias=?,
    hasAvatar=?
WHERE
    id=?
{%- code 
    *args = *sqlh.NewArgs(4) 
    args.Append(
    social.Handle,
    social.Alias,
    social.HasAvatar,
    social.ID,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qrySelect(qryArgs *sqlh.Args, args *social.Get) -%}
{%- collapsespace -%}
{%- code 
    args.Limit = sqlh.Limit100(args.Limit)
    app.BadReqIf(len(args.IDs) > 100, "max ids to query is 100")
    app.BadReqIf(args.HandlePrefix != "" && StrLen(args.HandlePrefix) < 3, "min handlePrefix len is 3")
    app.BadReqIf(len(args.IDs) == 0 && args.HandlePrefix == "", "no query parameters provided please")
    *qryArgs = *sqlh.NewArgs(len(args.IDs) + 1) 
-%}
SELECT id,
    handle,
    alias,
    hasAvatar
FROM socials
WHERE 
{% switch true %}
    {%- case len(args.IDs) > 0 -%}
        id IN ({%- for i := range args.IDs -%}{%- if i > 0 -%},{%- endif -%}?{%- endfor -%})
        ORDER BY FIELD (id,{%- for i := range args.IDs -%}{%- if i > 0 -%},{%- endif -%}?{%- endfor -%})
        {%- code
            *qryArgs = *sqlh.NewArgs(len(args.IDs)*2) 
            qryArgs.Append(args.IDs.ToIs()...)
            qryArgs.Append(args.IDs.ToIs()...)
        -%}
    {%- case args.HandlePrefix != "" -%}
        handle LIKE ?
        ORDER BY handle ASC
        LIMIT ?
        {%- code
            *qryArgs = *sqlh.NewArgs(2) 
            qryArgs.Append(sqlh.LikePrefix(args.HandlePrefix), args.Limit)
        -%}
{% endswitch%}
{%- endcollapsespace -%}
{%- endfunc -%}