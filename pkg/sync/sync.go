package sync

import (
	"strings"
	"time"

	"github.com/koltyakov/gosip/api"
	"github.com/koltyakov/sp-time-machine/pkg/state"

	log "github.com/sirupsen/logrus"
)

// Sync runs sync processing
func Sync(sp1, sp2 *api.SP, listUri string, since time.Time, syncState state.State, fullSync bool) error {
	// Processing initial list sync state
	e := syncState.GetList(listUri)
	checkSum := state.CheckSum(listUri)
	if !(e.MD5 == "[current]" || e.MD5 == checkSum) || fullSync {
		// Run full sync if something changes in list request data model or explicitely requested
		e.LastRun = state.DefaultStartDate()
	}
	modAfter := e.LastRun

	// Iterative full sync
	if state.IsFullSync(modAfter) {
		e.FullSyncSession = &since
		fullSync = true
	}

	list := sp1.Web().GetList(listUri)

	if fullSync {
		log.Infof("Full sync (list=%s)", listUri)
	} else if e.FullSyncSession == nil {
		log.Infof("Incr sync (list=%s, since=%s)\n", listUri, modAfter.UTC().Format("2006-01-02T15:04:05.000Z"))
	} else {
		log.Infof("Cont sync (list=%s, since=%s)\n", listUri, modAfter.UTC().Format("2006-01-02T15:04:05.000Z"))
	}

	d, err := list.Select("Title").Get()
	if err != nil {
		return err
	}

	if err := ensureTargetList(sp2, d.Data().Title, listUri); err != nil {
		return err
	}

	// // Fetch and process all available items for list within the range,
	// // run data write to target and UI feedback in the callback
	// cb := func(items []*w.WorkfrontItem, pageVector *w.PageVector) error {
	// 	err := target.SyncItems(listUri, items) // Writes 1000 items chunk to target

	// 	// Persist state using pageVector.LastMod value
	// 	func() {
	// 		e.MD5 = checkSum
	// 		e.LastRun = pageVector.LastMod
	// 		_ = syncState.SaveEnt(listUri, e)
	// 	}()

	// 	return err
	// }

	// // Fetches all items for a criteria
	// if err := ent.GetAll(modAfter, cb); err != nil {
	// 	return err
	// }

	// Persisting final list state
	if err := persistState(listUri, syncState, e, checkSum, since, fullSync); err != nil {
		return nil
	}

	return nil
}

func persistState(listUri string, syncState state.State, e *state.ListState, checkSum string, since time.Time, fullSync bool) error {
	e.MD5 = checkSum
	e.LastRun = since.UTC()
	// Persist last full sync datetime
	if fullSync || e.FullSyncSession != nil {
		e.FullSync = since.UTC()
		if e.FullSyncSession != nil {
			e.FullSync = *e.FullSyncSession
		}
		e.FullSyncSession = nil
	}
	// Persisting sync state to state file (state.json)
	return syncState.SaveList(listUri, e)
}

func ensureTargetList(sp *api.SP, title, listUri string) error {
	// Skip when list already exists
	if _, err := sp.Web().GetList(listUri).Get(); err == nil {
		return nil
	}
	// Create the list
	if _, err := sp.Web().Lists().AddWithURI(title, strings.Replace(listUri, "Lists/", "", -1), nil); err != nil {
		return err
	}
	list := sp.Web().GetList(listUri)
	// Provision fields
	f1 := `<Field Type="Text" Name="SrcID" DisplayName="Source Item ID" MaxLength="255" />`
	if _, err := list.Fields().CreateFieldAsXML(f1, 12); err != nil {
		return err
	}
	f2 := `<Field Type="Note" Name="Data" DisplayName="Item Data" NumLines="6" RichText="FALSE" RichTextMode="Compatible" />`
	if _, err := list.Fields().CreateFieldAsXML(f2, 12); err != nil {
		return err
	}
	return nil
}
