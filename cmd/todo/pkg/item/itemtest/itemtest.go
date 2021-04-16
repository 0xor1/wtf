package itemtest

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/cmd/todo/pkg/config"
	"github.com/0xor1/tlbx/cmd/todo/pkg/item"
	"github.com/0xor1/tlbx/cmd/todo/pkg/item/itemeps"
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
		app.JoinEps(listeps.Eps, itemeps.Eps),
		ratelimit.MeMware,
		nil,
		true,
		nil,
		func(c *autheps.Config) {
			c.OnDelete = listeps.OnDelete
		})
	defer r.CleanUp()

	testList1 := (&list.Create{
		Name: "Test list 1",
	}).MustDo(r.Ali().Client())

	name1 := "Test item 1"
	testItem1 := (&item.Create{
		List: testList1.ID,
		Name: name1,
	}).MustDo(r.Ali().Client())
	a.Equal(name1, testItem1.Name)

	name2 := "Test item 2"
	testItem2 := (&item.Create{
		List: testList1.ID,
		Name: name2,
	}).MustDo(r.Ali().Client())
	a.Equal(name2, testItem2.Name)

	one := (&item.One{
		List: testList1.ID,
		ID:   testItem1.ID,
	}).MustDo(r.Ali().Client())
	a.Equal(testItem1, one)

	nilOne := (&item.One{
		List: testList1.ID,
		ID:   app.ExampleID(),
	}).MustDo(r.Ali().Client())
	a.Nil(nilOne)

	get := (&item.Get{
		List: testList1.ID,
	}).MustDo(r.Ali().Client())
	a.Equal(testItem1, get.Set[0])
	a.Equal(testItem2, get.Set[1])
	a.False(get.More)

	get = (&item.Get{
		List: testList1.ID,
		IDs:  IDs{testItem2.ID, testItem1.ID},
	}).MustDo(r.Ali().Client())
	a.Equal(testItem2, get.Set[0])
	a.Equal(testItem1, get.Set[1])
	a.False(get.More)

	get = (&item.Get{
		List:         testList1.ID,
		NamePrefix:   ptr.String("Test i"),
		CreatedOnMin: ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax: ptr.Time(Now()),
		Asc:          ptr.Bool(false),
		Limit:        2,
	}).MustDo(r.Ali().Client())
	a.Equal(testItem2, get.Set[0])
	a.Equal(testItem1, get.Set[1])
	a.False(get.More)

	get = (&item.Get{
		List:         testList1.ID,
		NamePrefix:   ptr.String("Test i"),
		CreatedOnMin: ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax: ptr.Time(Now()),
		After:        ptr.ID(testItem1.ID),
		Sort:         item.SortName,
		Asc:          ptr.Bool(true),
		Limit:        2,
	}).MustDo(r.Ali().Client())
	a.Equal(testItem2, get.Set[0])
	a.False(get.More)

	get = (&item.Get{
		List:         testList1.ID,
		NamePrefix:   ptr.String("Test i"),
		CreatedOnMin: ptr.Time(Now().Add(-5 * time.Second)),
		CreatedOnMax: ptr.Time(Now()),
		Asc:          ptr.Bool(true),
		Limit:        1,
	}).MustDo(r.Ali().Client())
	a.Equal(testItem1, get.Set[0])
	a.True(get.More)

	rename := "Test item 1 rename"
	updatedItem1 := (&item.Update{
		List:     testList1.ID,
		ID:       testItem1.ID,
		Name:     &field.String{V: rename},
		Complete: &field.Bool{V: true},
	}).MustDo(r.Ali().Client())
	testItem1.Name = rename
	a.Equal(testItem1.Name, updatedItem1.Name)
	a.NotNil(updatedItem1.CompletedOn)

	updatedItem2 := (&item.Update{
		List:     testList1.ID,
		ID:       testItem2.ID,
		Complete: &field.Bool{V: true},
	}).MustDo(r.Ali().Client())
	a.NotNil(updatedItem2.CompletedOn)

	get = (&item.Get{
		List:      testList1.ID,
		Completed: ptr.Bool(true),
	}).MustDo(r.Ali().Client())
	a.Equal(updatedItem1, get.Set[0])
	a.Equal(updatedItem2, get.Set[1])
	a.False(get.More)

	get = (&item.Get{
		List:           testList1.ID,
		Completed:      ptr.Bool(true),
		CompletedOnMin: updatedItem1.CompletedOn,
		CompletedOnMax: updatedItem1.CompletedOn,
	}).MustDo(r.Ali().Client())
	a.Equal(updatedItem1, get.Set[0])
	a.Equal(1, len(get.Set))
	a.False(get.More)

	updatedItem1 = (&item.Update{
		List: testList1.ID,
		ID:   testItem1.ID,
	}).MustDo(r.Ali().Client())
	a.Equal(testItem1.Name, updatedItem1.Name)
	a.NotNil(updatedItem1.CompletedOn)

	updatedItem1 = (&item.Update{
		List:     testList1.ID,
		ID:       testItem1.ID,
		Complete: &field.Bool{V: false},
	}).MustDo(r.Ali().Client())
	a.Equal(testItem1.Name, updatedItem1.Name)
	a.Nil(updatedItem1.CompletedOn)

	(&item.Delete{List: testList1.ID}).MustDo(r.Ali().Client())
	(&item.Delete{List: testList1.ID, IDs: IDs{testItem1.ID}}).MustDo(r.Ali().Client())
}
