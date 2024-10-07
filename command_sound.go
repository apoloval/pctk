package pctk

// SoundPlay is a command that will play the sound with the given resource reference.
type SoundPlay struct {
	Sound *Sound
}

func (cmd SoundPlay) Execute(app *App, done *Promise) {
	cmd.Sound.Play(app)
	done.Complete()
}

// SoundStop is a command that will stop the sound with the given resource reference.
type SoundStop struct {
	Sound *Sound
}

func (cmd SoundStop) Execute(app *App, done *Promise) {
	cmd.Sound.Stop(app)
	done.Complete()
}
