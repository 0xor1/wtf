{% import "time" %}
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

{%- func qryDelete(args *sqlh.Args, me ID, client *ID, createdOn *time.Time) -%}
{%- collapsespace -%}
{%- code 
    *args = *sqlh.NewArgs(3) 
    args.Append(
    me,
) -%}
DELETE FROM fcmTokens
WHERE id=?
{%- if client != nil -%}
AND client=?
{%- code 
    args.Append(
    *client,
) -%}
{%- endif -%}
{%- if createdOn != nil -%}
AND createdOn<=?
{%- code 
    args.Append(
    createdOn,
) -%}
{%- endif -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryGetEnabled(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
SELECT enabled
FROM fcms 
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(
    me,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qrySetEnabled(args *sqlh.Args, me ID, enabled bool) -%}
{%- collapsespace -%}
INSERT INTO fcms (
    id,
    enabled
) VALUES (
    ?,
    ?
) ON DUPLICATE KEY UPDATE 
    id=VALUES(id),
    enabled=VALUES(enabled)
{%- code 
    *args = *sqlh.NewArgs(2) 
    args.Append(
    me,
    enabled,
 ) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryFifthYoungestToken(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
SELECT createdOn 
FROM fcmTokens 
WHERE id=? 
ORDER BY createdOn DESC 
LIMIT 4, 1
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(
    me,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryInsert(args *sqlh.Args, topic, token string, me ID, client ID, start time.Time) -%}
{%- collapsespace -%}
INSERT INTO fcmTokens (
    topic,
    token,
    id,
    client,
    createdOn
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
) ON DUPLICATE KEY UPDATE 
    topic=VALUES(topic),
    token=VALUES(token),
    id=VALUES(id),
    client=VALUES(client),
    createdOn=VALUES(createdOn)
{%- code 
    *args = *sqlh.NewArgs(5) 
    args.Append(
        topic,
        token,
        me,
        client,
        start,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}