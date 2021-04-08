{% import . "github.com/0xor1/tlbx/pkg/core" %}
{% import . "github.com/0xor1/tlbx/pkg/web/app/sql" %}

{% func qryInsert(args *sql.Args, auth *fullAuth) %}
{% collapsespace %}

INSERT INTO auths
    (id,
    email,
    registeredOn,
    activatedOn,
    newEmail,
    activateCode,
    changeEmailCode,
    lastPwdResetOn,
    salt,
    pwd,
    n,
    r,
    p)
VALUES (?,
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
    ?)

{%code 
    *args = sql.NewArgs(13) 
    args.Append(
    auth.ID,
    auth.Email, 
    auth.RegisteredOn, 
    auth.ActivatedOn, 
    auth.NewEmail, 
    auth.ActivateCode, 
    auth.ChangeEmailCode,
    auth.LastPwdResetOn,
    auth.Salt,
    auth.Pwd,
    auth.N,
    auth.R,
    auth.P,
)%}
{% endcollapsespace %}
{% endfunc %}

{% func qryGet(args *sql.Args, email *string, id *ID) %}
{% collapsespace %}
{%code 
    PanicIf(email == nil && id == nil, "one of email or id must not be nil")
    *args = sql.NewArgs(1) %}

SELECT id,
    email,
    registeredOn,
    activatedOn,
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
	{% if email != nil %}
	email
    {% code args.AppendOne(*email)%}
	{% else %}
	id
    {% code args.AppendOne(*id) %}
	{% endif %}
    = ?

{% endcollapsespace %}
{% endfunc %}

{% func qryUpdate(args *sql.Args, auth *fullAuth) %}
{% collapsespace %}

UPDATE auths
SET email=?,
    registeredOn=?,
    activatedOn=?,
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

{%code 
    *args = sql.NewArgs(13) 
    args.Append(
    auth.Email, 
    auth.RegisteredOn, 
    auth.ActivatedOn, 
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
)%}
{% endcollapsespace %}
{% endfunc %}