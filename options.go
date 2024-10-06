package pctk

// AppOption is a function that can be used to configure the application.
type AppOption func(*App)

// WithScreenCaption sets the screen caption of the application.
func WithScreenCaption(caption string) AppOption {
	return func(a *App) { a.screenCaption = caption }
}

// WithScreenZoom sets the screen zoom of the application.
func WithScreenZoom(zoom int32) AppOption {
	return func(a *App) { a.screenZoom = zoom }
}

// WithFlag allows you to enable or disable specific flags in the application.
func WithFlag(flag uint32, enable bool) AppOption {
	return func(a *App) {
		if enable {
			a.flags |= flag
		} else {
			a.flags &^= flag
		}
	}
}

// IsFlagEnabled checks if a specific flag is enabled.
func (a *App) IsFlagEnabled(flag uint32) bool {
	return a.flags&flag != 0
}

var defaultAppOptions = []AppOption{
	WithScreenCaption("Point&Click Toolkit"),
	WithScreenZoom(4),
}
