package project

import (
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/user/social"
)

type Project struct {
	task.Task
	Base
	Host       ID     `json:"host"`
	IsArchived bool   `json:"isArchived"`
	FileLimit  uint64 `json:"fileLimit"`
}

type Base struct {
	CurrencyCode string     `json:"currencyCode"`
	HoursPerDay  *uint8     `json:"hoursPerDay"`
	DaysPerWeek  *uint8     `json:"daysPerWeek"`
	StartOn      *time.Time `json:"startOn"`
	EndOn        *time.Time `json:"endOn"`
	IsPublic     bool       `json:"isPublic"`
}

type Create struct {
	CurrencyCode string     `json:"currencyCode,omitempty"`
	HoursPerDay  *uint8     `json:"hoursPerDay,omitempty"`
	DaysPerWeek  *uint8     `json:"daysPerWeek,omitempty"`
	StartOn      *time.Time `json:"startOn"`
	EndOn        *time.Time `json:"endOn"`
	IsPublic     bool       `json:"isPublic"`
	Name         string     `json:"name"`
}

func (_ *Create) Path() string {
	return "/project/create"
}

func (a *Create) Do(c *app.Client) (*Project, error) {
	res := &Project{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *Project {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type One struct {
	Host ID `json:"host"`
	ID   ID `json:"id"`
}

func (a *One) Do(c *app.Client) (*Project, error) {
	res, err := (&Get{Host: a.Host, IDs: IDs{a.ID}}).Do(c)
	if res != nil && len(res.Set) == 1 {
		return res.Set[0], err
	}
	return nil, err
}

func (a *One) MustDo(c *app.Client) *Project {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	Host ID `json:"host,omitempty"`
	// get other users projects that host is a member of
	Others       bool       `json:"others,omitempty"`
	IsArchived   bool       `json:"isArchived"`
	IsPublic     *bool      `json:"isPublic,omitempty"`
	IDs          IDs        `json:"ids,omitempty"`
	NamePrefix   *string    `json:"namePrefix,omitempty"`
	CreatedOnMin *time.Time `json:"createdOnMin,omitempty"`
	CreatedOnMax *time.Time `json:"createdOnMax,omitempty"`
	StartOnMin   *time.Time `json:"startOnMin,omitempty"`
	StartOnMax   *time.Time `json:"startOnMax,omitempty"`
	EndOnMin     *time.Time `json:"endOnMin,omitempty"`
	EndOnMax     *time.Time `json:"endOnMax,omitempty"`
	After        *ID        `json:"after,omitempty"`
	Sort         cnsts.Sort `json:"sort,omitempty"`
	Asc          *bool      `json:"asc,omitempty"`
	Limit        uint16     `json:"limit,omitempty"`
}

type GetRes struct {
	Set  []*Project `json:"set"`
	More bool       `json:"more"`
}

func (_ *Get) Path() string {
	return "/project/get"
}

func (a *Get) Do(c *app.Client) (*GetRes, error) {
	res := &GetRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *GetRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetLatestPublic struct{}

type GetLatestPublicRes struct {
	Set []*Project `json:"set"`
}

func (_ *GetLatestPublic) Path() string {
	return "/project/getLatestPublic"
}

func (a *GetLatestPublic) Do(c *app.Client) (*GetLatestPublicRes, error) {
	res := &GetLatestPublicRes{}
	err := app.Call(c, a.Path(), nil, &res)
	return res, err
}

func (a *GetLatestPublic) MustDo(c *app.Client) *GetLatestPublicRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Update struct {
	ID           ID              `json:"id,omitempty"`
	Name         *field.String   `json:"name,omitempty"`
	CurrencyCode *field.String   `json:"currencyCode,omitempty"`
	HoursPerDay  *field.UInt8Ptr `json:"hoursPerDay,omitempty"`
	DaysPerWeek  *field.UInt8Ptr `json:"daysPerWeek,omitempty"`
	StartOn      *field.TimePtr  `json:"startOn,omitempty"`
	EndOn        *field.TimePtr  `json:"endOn,omitempty"`
	IsArchived   *field.Bool     `json:"isArchived,omitempty"`
	IsPublic     *field.Bool     `json:"isPublic,omitempty"`
}

func (a *Update) Do(c *app.Client) (*Project, error) {
	res, err := (&Updates{a}).Do(c)
	if len(res) == 1 {
		return res[0], err
	}
	return nil, err
}

func (a *Update) MustDo(c *app.Client) *Project {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Updates []*Update

func (_ *Updates) Path() string {
	return "/project/update"
}

func (a *Updates) Do(c *app.Client) ([]*Project, error) {
	res := &[]*Project{}
	err := app.Call(c, a.Path(), a, &res)
	if res == nil {
		return nil, err
	}
	return *res, err
}

func (a *Updates) MustDo(c *app.Client) []*Project {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Delete IDs

func (_ *Delete) Path() string {
	return "/project/delete"
}

func (a *Delete) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type User struct {
	social.Social
	Role     cnsts.Role `json:"role"`
	IsActive bool       `json:"isActive"`
	TimeEst  uint64     `json:"timeEst"`
	TimeInc  uint64     `json:"timeInc"`
	CostEst  uint64     `json:"costEst"`
	CostInc  uint64     `json:"costInc"`
	FileN    uint64     `json:"fileN"`
	FileSize uint64     `json:"fileSize"`
	TaskN    uint64     `json:"taskN"`
}

type AddUsers struct {
	Host    ID          `json:"host"`
	Project ID          `json:"project"`
	Users   []*SendUser `json:"users"`
}

type SendUser struct {
	ID   ID         `json:"id"`
	Role cnsts.Role `json:"role"`
}

func (_ *AddUsers) Path() string {
	return "/project/addUsers"
}

func (a *AddUsers) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *AddUsers) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type GetMe struct {
	Host    ID `json:"host"`
	Project ID `json:"project"`
}

func (_ *GetMe) Path() string {
	return "/project/getMe"
}

func (a *GetMe) Do(c *app.Client) (*User, error) {
	res := &User{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *GetMe) MustDo(c *app.Client) *User {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetUsers struct {
	Host         ID          `json:"host"`
	Project      ID          `json:"project"`
	IDs          IDs         `json:"ids,omitempty"`
	Role         *cnsts.Role `json:"role,omitempty"`
	HandlePrefix *string     `json:"handlePrefix,omitempty"`
	After        *ID         `json:"after,omitempty"`
	Limit        uint16      `json:"limit,omitempty"`
}

type GetUsersRes struct {
	Set  []*User `json:"set"`
	More bool    `json:"more"`
}

func (_ *GetUsers) Path() string {
	return "/project/getUsers"
}

func (a *GetUsers) Do(c *app.Client) (*GetUsersRes, error) {
	res := &GetUsersRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *GetUsers) MustDo(c *app.Client) *GetUsersRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type SetUserRoles AddUsers

func (_ *SetUserRoles) Path() string {
	return "/project/setUserRoles"
}

func (a *SetUserRoles) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *SetUserRoles) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type RemoveUsers struct {
	Host    ID  `json:"host"`
	Project ID  `json:"project"`
	Users   IDs `json:"users"`
}

func (_ *RemoveUsers) Path() string {
	return "/project/removeUsers"
}

func (a *RemoveUsers) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *RemoveUsers) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}

type GetActivities struct {
	Host                ID         `json:"host"`
	Project             ID         `json:"project"`
	ExcludeDeletedItems bool       `json:"excludeDeletedItems,omitempty"`
	Task                *ID        `json:"task,omitempty"`
	Item                *ID        `json:"item,omitempty"`
	User                *ID        `json:"user,omitempty"`
	OccuredAfter        *time.Time `json:"occurredAfter,omitempty"`
	OccuredBefore       *time.Time `json:"occurredBefore,omitempty"`
	Limit               uint16     `json:"limit,omitempty"`
}

func (_ *GetActivities) Path() string {
	return "/project/getActivities"
}

func (a *GetActivities) Do(c *app.Client) (*GetActivitiesRes, error) {
	res := &GetActivitiesRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *GetActivities) MustDo(c *app.Client) *GetActivitiesRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetActivitiesRes struct {
	Set  []*Activity `json:"set"`
	More bool        `json:"more"`
}

type Activity struct {
	Task        *ID          `json:"task,omitempty"`
	OccurredOn  time.Time    `json:"occurredOn"`
	User        ID           `json:"user"`
	Item        ID           `json:"item"`
	ItemType    cnsts.Type   `json:"itemType"`
	TaskDeleted bool         `json:"taskDeleted"`
	ItemDeleted bool         `json:"itemDeleted"`
	Action      cnsts.Action `json:"action"`
	TaskName    *string      `json:"taskName,omitempty"`
	ItemName    *string      `json:"itemName,omitempty"`
	ExtraInfo   *json.Json   `json:"extraInfo,omitempty"`
}
