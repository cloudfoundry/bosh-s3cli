package fakes

type FakeRunner struct {
	RunCmdName string
	RunCmdArgs []string
}

func (run *FakeRunner) Run(cmdName string, args []string) (err error) {
	run.RunCmdName = cmdName
	run.RunCmdArgs = args
	return
}
