package gcs

import (
	"fmt"
	"github.com/onepanelio/templates/sidecars/filesyncer/util"
	"os/exec"
)

func Sync() {
	// Make sure we don't run more than once sync at a time.
	util.Mux.Lock()
	if util.Syncing {
		util.Mux.Unlock()
		return
	} else {
		util.Syncing = true
		util.Mux.Unlock()
	}

	var cmd *exec.Cmd

	// Activate service account
	cmd = util.Command("gcloud", "auth", "activate-service-account", "--key-file", util.Config.GCS.ServiceAccountKeyPath)
	cmd.Run()

	// Sync to or from bucket
	uri := fmt.Sprintf("gs://%v/%v", util.Bucket, util.Prefix)
	if util.Action == util.ActionDownload {
		util.Status.IsDownloading = true
		cmd = util.Command("gsutil", "-m", "rsync", "-d", "-r", uri, util.Path)
	}
	if util.Action == util.ActionUpload {
		util.Status.IsUploading = true
		cmd = util.Command("gsutil", "-m", "rsync", "-d", "-r", util.Path, uri)
	}

	util.Status.ClearError()
	if err := util.RunCommand(cmd); err != nil {
		util.Status.ReportError(err)
		util.Mux.Lock()
		util.Syncing = false
		util.Mux.Unlock()
		return
	}

	if util.Action == util.ActionDownload {
		util.Status.MarkLastDownload()
	}
	if util.Action == util.ActionUpload {
		util.Status.MarkLastUpload()
	}
	if err := util.SaveSyncStatus(); err != nil {
		fmt.Printf("[error] save sync status: Message %v\n", err)
	}

	util.Mux.Lock()
	util.Syncing = false
	util.Mux.Unlock()
}
