//go:build e2e
// +build e2e

package upload

import (
	"context"
	"encoding/csv"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/simulot/immich-go/cmd"
	"github.com/simulot/immich-go/helpers/fileevent"
	"github.com/simulot/immich-go/internal/fakefs"
)

// Simulate a takeout archive with the list of zipped files
func simulate_upload(t *testing.T, zipList string, dateFormat string, forceMissingJSON bool) {
	ic := &icCatchUploadsAssets{
		albums: map[string][]string{},
	}
	ctx := context.Background()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	serv := cmd.SharedFlags{
		Immich:   ic,
		LogLevel: "INFO",
		Jnl:      fileevent.NewRecorder(log, false),
		Log:      log,
	}

	fsOpener := func() ([]fs.FS, error) {
		return fakefs.ScanFileList(zipList, dateFormat)
	}
	os.Remove(filepath.Dir(zipList) + "/debug.log")
	args := []string{"-google-photos", "-no-ui", "-debug-counters", "-log-file=" + filepath.Dir(zipList) + "/debug.log"}

	app, err := newCommand(ctx, &serv, args, fsOpener)
	if err != nil {
		t.Errorf("can't instantiate the UploadCmd: %s", err)
		return
	}
	app.ForceUploadWhenNoJSON = forceMissingJSON
	err = app.run(ctx)
	if err != nil {
		t.Errorf("can't run the UploadCmd: %s", err)
		return
	}
	os.Remove(zipList + ".csv")
	f, err := os.Create(zipList + ".csv")
	if err != nil {
		t.Errorf("can't create CSV: %s", err)
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	counts := app.Jnl.GetCounts()
	w.Write([]string{"indicator", "num indicator", "count"})
	for k, c := range counts {
		w.Write([]string{fileevent.Code(k).String(), strconv.Itoa(k), strconv.FormatInt(c, 10)})
	}
}

func TestPixilTakeOut(t *testing.T) {
	initMyEnv(t)

	simulate_upload(t, myEnv["IMMICH_TESTFILES"]+"/User Files/pixil/list.lst", "01-02-2006 15:04", false)
}

func TestPhyl404TakeOut(t *testing.T) {
	initMyEnv(t)

	simulate_upload(t, myEnv["IMMICH_TESTFILES"]+"/User Files/Phyl404/list.lst", "2006-01-02 15:04", false)
}

func TestPhyl404_2TakeOut(t *testing.T) {
	initMyEnv(t)

	simulate_upload(t, myEnv["IMMICH_TESTFILES"]+"/User Files/Phyl404#2/list.lst", "2006-01-02 15:04", false)
}

func TestSteve81TakeOut(t *testing.T) {
	initMyEnv(t)

	simulate_upload(t, myEnv["IMMICH_TESTFILES"]+"/User Files/Steve81/list.list", "2006-01-02 15:04", false)
}

func TestMuetyTakeOut(t *testing.T) {
	initMyEnv(t)

	simulate_upload(t, myEnv["IMMICH_TESTFILES"]+"/User Files/muety/list.lst", "01-02-2006 15:04", false)
}

func TestMissingJSONTakeOut(t *testing.T) {
	initMyEnv(t)

	simulate_upload(t, myEnv["IMMICH_TESTFILES"]+"/User Files/MissingJSON/list.lst", "01-02-2006 15:04", true)
}
