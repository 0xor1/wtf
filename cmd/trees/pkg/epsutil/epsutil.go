package epsutil

import (
	"net/http"

	"github.com/0xor1/tlbx/cmd/trees/pkg/cnsts"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/sql"
)

func MustHaveAccess(tlbx app.Tlbx, host, project ID, user *ID, role cnsts.Role) {
	userExist := user != nil
	if userExist && user.Equal(host) {
		return
	}

	srv := service.Get(tlbx)
	if !userExist || role == cnsts.RoleReader {
		row := srv.Data().QueryRow(`SELECT isPublic FROM projects WHERE host=? AND id=?`, host, project)
		isPublic := false
		sql.PanicIfIsntNoRows(row.Scan(&isPublic))
		if isPublic {
			return
		}
		// at this point project isnt public so if no active session return 403
		app.ReturnIf(!userExist, http.StatusForbidden, "")
	}

	row := srv.Data().QueryRow(`SELECT 1 FROM users WHERE host=? AND project=? AND id=? AND role<=? AND isActive=1`, host, project, *user, role)
	hasAccess := false
	sql.PanicIfIsntNoRows(row.Scan(&hasAccess))
	app.ReturnIf(!hasAccess, http.StatusForbidden, "")
}

func IMustHaveAccess(tlbx app.Tlbx, host, project ID, role cnsts.Role) {
	iExist := me.Exists(tlbx)
	var mePtr *ID
	if iExist {
		meID := me.Get(tlbx)
		mePtr = &meID
	}
	MustHaveAccess(tlbx, host, project, mePtr, role)
}

func MustLockProject(tlbx app.Tlbx, tx service.Tx, host, id ID) {
	projectExists := false
	row := tx.QueryRow(`SELECT COUNT(*)=1 FROM projectLocks WHERE host=? AND id=? FOR UPDATE`, host, id)
	sql.PanicIfIsntNoRows(row.Scan(&projectExists))
	app.ReturnIf(!projectExists, http.StatusNotFound, "no such project")
}

func StorePrefix(host ID, projectAndOrTask ...ID) string {
	prefix := host.String()
	if len(projectAndOrTask) > 0 {
		prefix += "/" + projectAndOrTask[0].String()
		if len(projectAndOrTask) > 1 {
			prefix += "/" + projectAndOrTask[1].String()
			PanicIf(len(projectAndOrTask) > 2, "invalid file address")
		}
	}
	return prefix
}

func LogActivity(tlbx app.Tlbx, tx service.Tx, host, project, item ID, itemType cnsts.Type, action cnsts.Action, itemName *string, extraInfo interface{}) {
	me := me.Get(tlbx)
	var ei *string
	if extraInfo != nil {
		eiStr := string(json.MustMarshal(extraInfo))
		ei = &eiStr
	}
	itemHasBeenDeleted := action == cnsts.ActionDeleted
	_, err := tx.Exec(`INSERT INTO activities(host, project, occurredOn, user, item, itemType, itemHasBeenDeleted, action, itemName, extraInfo) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, host, project, NowMilli(), me, item, itemType, itemHasBeenDeleted, action, itemName, ei)
	PanicOn(err)
	if itemHasBeenDeleted {
		// if this is deleting an item we need to update all previous activities on this item to have itemHasBeenDeleted
		_, err = tx.Exec(`UPDATE activities SET itemHasBeenDeleted=1 WHERE host=? AND project=? AND item=?`, host, project, item)
		PanicOn(err)
	}
}

func ActivityItemRename(tx service.Tx, host, project, item ID, newItemName string) {
	// keep all projectActivity entries itemName values up to date
	_, err := tx.Exec(`UPDATE activities SET itemName=? WHERE host=? AND project=? AND item=?`, newItemName, host, project, item)
	PanicOn(err)
}