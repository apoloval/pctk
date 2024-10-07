package main

import (
	"github.com/apoloval/pctk"
)

func main() {
	loader := pctk.NewResourceFileLoader("./")

	app := pctk.New(loader, pctk.WithScreenZoom(4))
	app.RunCommand(pctk.ScriptRun{ScriptRef: pctk.NewResourceRef("resources", "scripts/boot")})
	app.Run()
}
