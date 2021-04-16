package listtest

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/todo/pkg/config"
	"github.com/0xor1/tlbx/cmd/todo/pkg/list"
	"github.com/0xor1/tlbx/cmd/todo/pkg/list/listeps"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/auth/autheps"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(
		config.Get(),
		listeps.Eps,
		ratelimit.MeMware,
		nil,
		true,
		nil,
		func(c *autheps.Config) {
			c.OnDelete = listeps.OnDelete
		})
	defer r.CleanUp()

	name1 := "Test list 1"
	testList1 := (&list.Create{
		Name: name1,
	}).MustDo(r.Ali().Client())
	a.Equal(name1, testList1.Name)

	name2 := "Test list 2"
	testList2 := (&list.Create{
		Name: name2,
	}).MustDo(r.Ali().Client())
	a.Equal(name2, testList2.Name)

	get := (&list.One{
		ID: testList1.ID,
	}).MustDo(r.Ali().Client())
	a.Equal(testList1, get)

	getNil := (&list.One{
		ID: app.ExampleID(),
	}).MustDo(r.Ali().Client())
	a.Nil(getNil)

	getSet := (&list.Get{}).MustDo(r.Ali().Client())
	a.Equal(testList1, getSet.Set[0])
	a.Equal(testList2, getSet.Set[1])
	a.False(getSet.More)

	getSet = (&list.Get{
		IDs: IDs{testList2.ID, testList1.ID},
	}).MustDo(r.Ali().Client())
	a.Equal(testList2, getSet.Set[0])
	a.Equal(testList1, getSet.Set[1])
	a.False(getSet.More)

	getSet = (&list.Get{
		NamePrefix:            ptr.String("Test l"),
		CreatedOnMin:          ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax:          ptr.Time(Now()),
		TodoItemCountMin:      ptr.Int(0),
		TodoItemCountMax:      ptr.Int(1),
		CompletedItemCountMin: ptr.Int(0),
		CompletedItemCountMax: ptr.Int(1),
		Asc:                   ptr.Bool(false),
		Limit:                 2,
	}).MustDo(r.Ali().Client())
	a.Equal(testList2, getSet.Set[0])
	a.Equal(testList1, getSet.Set[1])
	a.False(getSet.More)

	getSet = (&list.Get{
		NamePrefix:            ptr.String("Test l"),
		CreatedOnMin:          ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax:          ptr.Time(Now()),
		TodoItemCountMin:      ptr.Int(0),
		TodoItemCountMax:      ptr.Int(1),
		CompletedItemCountMin: ptr.Int(0),
		CompletedItemCountMax: ptr.Int(1),
		After:                 ptr.ID(testList1.ID),
		Sort:                  list.SortTodoItemCount,
		Asc:                   ptr.Bool(true),
		Limit:                 2,
	}).MustDo(r.Ali().Client())
	a.Equal(testList2, getSet.Set[0])
	a.False(getSet.More)

	getSet = (&list.Get{
		NamePrefix:            ptr.String("Test l"),
		CreatedOnMin:          ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax:          ptr.Time(Now()),
		TodoItemCountMin:      ptr.Int(0),
		TodoItemCountMax:      ptr.Int(1),
		CompletedItemCountMin: ptr.Int(0),
		CompletedItemCountMax: ptr.Int(1),
		Asc:                   ptr.Bool(true),
		Limit:                 1,
	}).MustDo(r.Ali().Client())
	a.Equal(testList1, getSet.Set[0])
	a.True(getSet.More)

	newName := "New Name"
	updatedList := (&list.Update{
		ID:   testList1.ID,
		Name: field.String{V: newName},
	}).MustDo(r.Ali().Client())
	testList1.Name = newName
	a.Equal(testList1, updatedList)

	(&list.Delete{}).MustDo(r.Ali().Client())
	(&list.Delete{IDs: IDs{testList1.ID}}).MustDo(r.Ali().Client())
}
