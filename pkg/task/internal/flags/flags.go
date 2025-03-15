package flags

import (
	"cmp"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"

	"github.com/hypershift-community/dev-console/pkg/task"
	"github.com/hypershift-community/dev-console/pkg/task/errors"
	"github.com/hypershift-community/dev-console/pkg/task/internal/env"
	"github.com/hypershift-community/dev-console/pkg/task/internal/experiments"
	"github.com/hypershift-community/dev-console/pkg/task/internal/sort"
	"github.com/hypershift-community/dev-console/pkg/task/taskfile/ast"
)

const usage = `Usage: task [flags...] [task...]

Runs the specified task(s). Falls back to the "default" task if no task name
was specified, or lists all tasks if an unknown task name was specified.

Example: 'task hello' with the following 'Taskfile.yml' file will generate an
'output.txt' file with the content "hello".

'''
version: '3'
tasks:
  hello:
    cmds:
      - echo "I am going to write a file named 'output.txt' now."
      - echo "hello" > output.txt
    generates:
      - output.txt
'''

Options:
`

var (
	Version     bool
	Help        bool
	Init        bool
	Completion  string
	List        bool
	ListAll     bool
	ListJson    bool
	TaskSort    string
	Status      bool
	NoStatus    bool
	Insecure    bool
	Force       bool
	ForceAll    bool
	Watch       bool
	Verbose     bool
	Silent      bool
	AssumeYes   bool
	Dry         bool
	Summary     bool
	ExitCode    bool
	Parallel    bool
	Concurrency int
	Dir         string
	Entrypoint  string
	Output      ast.Output
	Color       bool
	Interval    time.Duration
	Global      bool
	Experiments bool
	Download    bool
	Offline     bool
	ClearCache  bool
	Timeout     time.Duration
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
	pflag.Usage = func() {
		log.Print(usage)
		pflag.PrintDefaults()
	}
	offline, err := strconv.ParseBool(cmp.Or(env.GetTaskEnv("OFFLINE"), "false"))
	if err != nil {
		offline = false
	}
	pflag.BoolVar(&Version, "version", false, "Show Task version.")
	pflag.BoolVarP(&Help, "help", "h", false, "Shows Task usage.")
	pflag.BoolVarP(&Init, "init", "i", false, "Creates a new Taskfile.yml in the current folder.")
	pflag.StringVar(&Completion, "completion", "", "Generates shell completion script.")
	pflag.BoolVarP(&List, "list", "l", false, "Lists tasks with description of current Taskfile.")
	pflag.BoolVarP(&ListAll, "list-all", "a", false, "Lists tasks with or without a description.")
	pflag.BoolVarP(&ListJson, "json", "j", false, "Formats task list as JSON.")
	pflag.StringVar(&TaskSort, "sort", "", "Changes the order of the tasks when listed. [default|alphanumeric|none].")
	pflag.BoolVar(&Status, "status", false, "Exits with non-zero exit code if any of the given tasks is not up-to-date.")
	pflag.BoolVar(&NoStatus, "no-status", false, "Ignore status when listing tasks as JSON")
	pflag.BoolVar(&Insecure, "insecure", false, "Forces Task to download Taskfiles over insecure connections.")
	pflag.BoolVarP(&Watch, "watch", "w", false, "Enables watch of the given task.")
	pflag.BoolVarP(&Verbose, "verbose", "v", false, "Enables verbose mode.")
	pflag.BoolVarP(&Silent, "silent", "s", false, "Disables echoing.")
	pflag.BoolVarP(&AssumeYes, "yes", "y", false, "Assume \"yes\" as answer to all prompts.")
	pflag.BoolVarP(&Parallel, "parallel", "p", false, "Executes tasks provided on command line in parallel.")
	pflag.BoolVarP(&Dry, "dry", "n", false, "Compiles and prints tasks in the order that they would be run, without executing them.")
	pflag.BoolVar(&Summary, "summary", false, "Show summary about a task.")
	pflag.BoolVarP(&ExitCode, "exit-code", "x", false, "Pass-through the exit code of the task command.")
	pflag.StringVarP(&Dir, "dir", "d", "", "Sets directory of execution.")
	pflag.StringVarP(&Entrypoint, "taskfile", "t", "", `Choose which Taskfile to run. Defaults to "Taskfile.yml".`)
	pflag.StringVarP(&Output.Name, "output", "o", "", "Sets output style: [interleaved|group|prefixed].")
	pflag.StringVar(&Output.Group.Begin, "output-group-begin", "", "Message template to print before a task's grouped output.")
	pflag.StringVar(&Output.Group.End, "output-group-end", "", "Message template to print after a task's grouped output.")
	pflag.BoolVar(&Output.Group.ErrorOnly, "output-group-error-only", false, "Swallow output from successful tasks.")
	pflag.BoolVarP(&Color, "color", "c", true, "Colored output. Enabled by default. Set flag to false or use NO_COLOR=1 to disable.")
	pflag.IntVarP(&Concurrency, "concurrency", "C", 0, "Limit number of tasks to run concurrently.")
	pflag.DurationVarP(&Interval, "interval", "I", 0, "Interval to watch for changes.")
	pflag.BoolVarP(&Global, "global", "g", false, "Runs global Taskfile, from $HOME/{T,t}askfile.{yml,yaml}.")
	pflag.BoolVar(&Experiments, "experiments", false, "Lists all the available experiments and whether or not they are enabled.")

	// Gentle force experiment will override the force flag and add a new force-all flag
	if experiments.GentleForce.Enabled() {
		pflag.BoolVarP(&Force, "force", "f", false, "Forces execution of the directly called task.")
		pflag.BoolVar(&ForceAll, "force-all", false, "Forces execution of the called task and all its dependant tasks.")
	} else {
		pflag.BoolVarP(&ForceAll, "force", "f", false, "Forces execution even when the task is up-to-date.")
	}

	// Remote Taskfiles experiment will adds the "download" and "offline" flags
	if experiments.RemoteTaskfiles.Enabled() {
		pflag.BoolVar(&Download, "download", false, "Downloads a cached version of a remote Taskfile.")
		pflag.BoolVar(&Offline, "offline", offline, "Forces Task to only use local or cached Taskfiles.")
		pflag.DurationVar(&Timeout, "timeout", time.Second*10, "Timeout for downloading remote Taskfiles.")
		pflag.BoolVar(&ClearCache, "clear-cache", false, "Clear the remote cache.")
	}

	pflag.Parse()
}

func Validate() error {
	if Download && Offline {
		return errors.New("task: You can't set both --download and --offline flags")
	}

	if Download && ClearCache {
		return errors.New("task: You can't set both --download and --clear-cache flags")
	}

	if Global && Dir != "" {
		return errors.New("task: You can't set both --global and --dir")
	}

	if Output.Name != "group" {
		if Output.Group.Begin != "" {
			return errors.New("task: You can't set --output-group-begin without --output=group")
		}
		if Output.Group.End != "" {
			return errors.New("task: You can't set --output-group-end without --output=group")
		}
		if Output.Group.ErrorOnly {
			return errors.New("task: You can't set --output-group-error-only without --output=group")
		}
	}

	if List && ListAll {
		return errors.New("task: cannot use --list and --list-all at the same time")
	}

	if ListJson && !List && !ListAll {
		return errors.New("task: --json only applies to --list or --list-all")
	}

	if NoStatus && !ListJson {
		return errors.New("task: --no-status only applies to --json with --list or --list-all")
	}

	return nil
}

// WithExecutorOptions is a special internal functional option that is used to pass flags
// from the CLI directly to the executor.
func WithExecutorOptions() task.ExecutorOption {
	return func(e *task.Executor) {
		// Set the sorter
		var sorter sort.Sorter
		switch TaskSort {
		case "none":
			sorter = nil
		case "alphanumeric":
			sorter = sort.AlphaNumeric
		}

		// Change the directory to the user's home directory if the global flag is set
		dir := Dir
		if Global {
			home, err := os.UserHomeDir()
			if err == nil {
				dir = home
			}
		}

		e.Options(
			task.ExecutorWithDir(dir),
			task.ExecutorWithEntrypoint(Entrypoint),
			task.ExecutorWithForce(Force),
			task.ExecutorWithForceAll(ForceAll),
			task.ExecutorWithInsecure(Insecure),
			task.ExecutorWithDownload(Download),
			task.ExecutorWithOffline(Offline),
			task.ExecutorWithTimeout(Timeout),
			task.ExecutorWithWatch(Watch),
			task.ExecutorWithVerbose(Verbose),
			task.ExecutorWithSilent(Silent),
			task.ExecutorWithAssumeYes(AssumeYes),
			task.ExecutorWithDry(Dry || Status),
			task.ExecutorWithSummary(Summary),
			task.ExecutorWithParallel(Parallel),
			task.ExecutorWithColor(Color),
			task.ExecutorWithConcurrency(Concurrency),
			task.ExecutorWithInterval(Interval),
			task.ExecutorWithOutputStyle(Output),
			task.ExecutorWithTaskSorter(sorter),
			task.ExecutorWithVersionCheck(true),
		)
	}
}
