package main

const (
	appName   string = "tf-reconcile-reader"
	appAuthor string = "andrei@merlescu.net"

	envConfigFile string = "CONFIG_FILE"
	envVimMode    string = "FIGS_VIM_MODE"
	envExecDir    string = "FIGS_EXEC_DIR"
	envTfDir      string = "FIGS_TF_DIR"
	envTfS3Bucket string = "FIGS_TF_S3_BUCKET"
	envTfState    string = "FIGS_TF_STATE"
	envGitHub     string = "FIGS_GITHUB"

	argInputFile            string = "input"
	argAliasInputFile       string = "i"
	argJsonOutput           string = "json"
	argAliasJsonOutput      string = "j"
	argNonInteractive       string = "non-interactive"
	argCommandContains      string = "contains"
	argAliasCommandContains string = "c"
	argVersion              string = "v"
	argVimEnabled           string = "vi"
	argSaveDir              string = "save"
	argGitHub               string = "github"

	viewMain viewState = iota
	viewBackup
	viewBackupDetail
	viewResultsCategory
	viewResultsList
	viewResultsDetail
	viewConfig
	viewConfigEdit
	viewDeleteConfirm
	viewRunningCommand
	viewExecutionLogDetail
	viewCommandRunner
)
