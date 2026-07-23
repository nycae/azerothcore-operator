package realm

import (
	"log"
	"text/template"
)

// TODO: Move this to a file?
const worldServerTemplate = `
[worldserver]
RealmID     = {{ .RealmID }}
PlayerLimit = {{ .PlayerLimit }}

LoginDatabaseInfo     = "{{ .AuthDatabaseConnStr }}"
WorldDatabaseInfo     = "{{ .WorldDatabaseConnStr }}"
CharacterDatabaseInfo = "{{ .CharacterDatabaseConnStr }}"
`

var (
	WorldServerConfigTemplate *template.Template
)

func init() {
	var err error
	WorldServerConfigTemplate, err = template.New("realm").Parse(worldServerTemplate)
	if err != nil {
		log.Fatalf("unable to parse world server config template: %v", err)
	}
}
