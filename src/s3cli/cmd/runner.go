package cmd

type Runner interface {
	Run(cmdName string, args []string) (err error)
}

type concreteRunner struct {
	factory Factory
}

func NewRunner(factory Factory) (runner Runner) {
	return concreteRunner{
		factory: factory,
	}
}

func (run concreteRunner) Run(cmdName string, args []string) (err error) {
	cmd, err := run.factory.Create(cmdName)
	if err != nil {
		return
	}

	err = cmd.Run(args)
	return
}
