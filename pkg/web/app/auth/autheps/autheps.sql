{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import sqlh "github.com/0xor1/tlbx/pkg/web/app/sql" %}

{%- func qryInsert(args *sqlh.Args, auth *fullAuth) -%}
{%- collapsespace -%}
INSERT INTO auths(
    id,
    email,
    isActivated,
    registeredOn,
    newEmail,
    activateCode,
    changeEmailCode,
    lastPwdResetOn,
    salt,
    pwd,
    n,
    r,
    p
)
VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
{%- code 
    *args = *sqlh.NewArgs(13) 
    args.Append(
    auth.ID,
    auth.Email, 
    auth.IsActivated, 
    auth.RegisteredOn, 
    auth.NewEmail, 
    auth.ActivateCode, 
    auth.ChangeEmailCode,
    auth.LastPwdResetOn,
    auth.Salt,
    auth.Pwd,
    auth.N,
    auth.R,
    auth.P,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qrySelect(args *sqlh.Args, email *string, id *ID) -%}
{%- collapsespace -%}
{%- code 
    PanicIf(email == nil && id == nil, "one of email or id must not be nil")
    *args = *sqlh.NewArgs(1) -%}
SELECT
    id,
    email,
    isActivated,
    registeredOn,
    newEmail,
    activateCode,
    changeEmailCode,
    lastPwdResetOn,
    salt,
    pwd,
    n,
    r,
    p
FROM auths
WHERE
	{%- if email != nil -%}
	email
    {%- code args.AppendOne(*email) -%}
	{%- else -%}
	id
    {%- code args.AppendOne(*id) -%}
	{%- endif -%}
    =?
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryUpdate(args *sqlh.Args, auth *fullAuth) -%}
{%- collapsespace -%}
UPDATE auths
SET email=?,
    isActivated=?,
    registeredOn=?,
    newEmail=?,
    activateCode=?,
    changeEmailCode=?,
    lastPwdResetOn=?,
    salt=?,
    pwd=?,
    n=?,
    r=?,
    p=?
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(13) 
    args.Append(
    auth.Email,
    auth.IsActivated,
    auth.RegisteredOn,
    auth.NewEmail, 
    auth.ActivateCode,
    auth.ChangeEmailCode,
    auth.LastPwdResetOn,
    auth.Salt,
    auth.Pwd,
    auth.N,
    auth.R,
    auth.P,
    auth.ID,
) -%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryDelete(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
DELETE FROM auths
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(me)
-%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryOnRegister(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
INSERT INTO auths (
    id,
    isActivated,
    registeredOn
)
VALUES (
    ?,
    ?,
    ?
)
{%- code 
    *args = *sqlh.NewArgs(3) 
    args.Append(me, 0, Now())
-%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryOnActivate(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
UPDATE auths
SET isActivated=1
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(me)
-%}
{%- endcollapsespace -%}
{%- endfunc -%}

{%- func qryOnDelete(args *sqlh.Args, me ID) -%}
{%- collapsespace -%}
DELETE FROM auths
WHERE id=?
{%- code 
    *args = *sqlh.NewArgs(1) 
    args.Append(me)
-%}
{%- endcollapsespace -%}
{%- endfunc -%}