package pctk

import "sync"

// Command is a command that can be executed by the application. Every action requires to the
// application that has a side effect should be encapsulated in a command in order to ensure
// thread safety.
type Command interface {
	Execute(*App, *Promise)
}

// CommandFunc is a sync function that can be used as a command.
type CommandFunc func(*App) (any, error)

// Execute implements the Command interface.
func (f CommandFunc) Execute(a *App, done *Promise) {
	v, err := f(a)
	done.CompleteWith(v, err)
}

// CommandAsyncFunc is an async function that can be used as a command.
type CommandAsyncFunc func(*App) Future

// Execute implements the Command interface.
func (f CommandAsyncFunc) Execute(a *App, done *Promise) {
	done.Bind(f(a))
}

// CommandQueue is a queue of commands that will be executed by the application during the next
// frame.
type CommandQueue struct {
	mutex    sync.Mutex
	commands []func(*App)
}

// PushCommand will put the given command in the queue to be executed by the application during the
// next frame. This function is thread safe and can be called from any goroutine.
func (q *CommandQueue) PushCommand(c Command) Future {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	prom := NewPromise()
	q.commands = append(q.commands, func(app *App) {
		c.Execute(app, prom)
	})

	return prom
}

// Execute will execute all the commands in the queue. This function should be called by the
// application during the frame update.
func (q *CommandQueue) Execute(app *App) {
	q.mutex.Lock()
	commands := q.commands
	q.commands = nil
	q.mutex.Unlock()

	for _, c := range commands {
		c(app)
	}
}

// RunCommand will put the given command in the queue to be executed by the application during
// the next frame. This function is thread safe and can be called from any goroutine.
func (a *App) RunCommand(c Command) Future {
	return a.commands.PushCommand(c)
}

// RunCommandSequence will run a sequence of commands that will be executed one after the other.
func (a *App) RunCommandSequence(cmd Command, rest ...Command) Future {
	fut := a.RunCommand(cmd)
	for _, c := range rest {
		fut = Continue(fut, func(_ any) Future {
			return a.RunCommand(c)
		})
	}
	return fut
}
