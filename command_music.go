package pctk

// MusicPlay is a command that will play the music with the given resource reference.
type MusicPlay struct {
	Music *Music
}

func (cmd MusicPlay) Execute(app *App, done *Promise) {
	app.PlayMusic(cmd.Music)
	done.Complete()
}

// MusicStop is a command that will stop the music.
type MusicStop struct{}

func (cmd MusicStop) Execute(app *App, done *Promise) {
	app.StopMusic()
	done.Complete()
}
