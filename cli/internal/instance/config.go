package instance

import (
	"fmt"
	"os"
	"text/template"
)

const configTemplate = `# PMocker 生成的实例配置
jwt:
  signing-key: pmocker-instance-{{.ID}}
  expires-time: 7d
  buffer-time: 1d
  issuer: pmocker

zap:
  level: info
  prefix: '[pmocker]'
  format: console
  director: log
  encode-level: LowercaseColorLevelEncoder
  show-line: true
  log-in-console: true

redis:
  useCluster: false
  addr: 127.0.0.1:6379
  password: ""
  db: 0

system:
  addr: {{.Port}}
  db-type: sqlite
  oss-type: local
  use-redis: false
  use-mongo: false
  use-multipoint: false
  iplimit-count: 15000
  iplimit-time: 3600
  router-prefix: ""
  use-strict-auth: false
  disable-auto-migrate: false

mysql:
  path: ""
  port: ""
  config: ""
  db-name: ""
  username: ""
  password: ""
  max-idle-conns: 10
  max-open-conns: 100

sqlite:
  path: "{{.VolumePath}}"
  port: ""
  config: ""
  db-name: "system.db"
  username: ""
  password: ""
  max-idle-conns: 10
  max-open-conns: 100

pgsql:
  path: ""
  port: ""
  config: ""
  db-name: ""
  username: ""
  password: ""
  max-idle-conns: 10
  max-open-conns: 100

db-list: []

local:
  path: "{{.UploadsPath}}"

autocode:
  transfer-restart: true
  root: ""

captcha:
  open-captcha: 0
  open-captcha-timeout: 3600
  captcha-long: 4
  captcha-open: false
  type: string
  driver: math

timer:
  start: true
  spec: "@daily"
  detail:
    - tableName: sys
      compareField: created_at
      interval: 2160h
`

type configParams struct {
	ID          string
	Port        int
	VolumePath  string
	UploadsPath string
}

func GenerateConfig(vm *VolumeManager, inst *Instance) error {
	params := configParams{
		ID:          inst.ID,
		Port:        inst.Port,
		VolumePath:  vm.Path(inst.VolumeID),
		UploadsPath: vm.UploadsPath(inst.VolumeID),
	}
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("parse config template: %w", err)
	}
	cfgPath := vm.ConfigPath(inst.VolumeID)
	f, err := os.Create(cfgPath)
	if err != nil {
		return fmt.Errorf("create config.yaml: %w", err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, params); err != nil {
		return fmt.Errorf("execute config template: %w", err)
	}
	return nil
}
