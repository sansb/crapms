# CrapMS

CrapMS (Crapfiguration Management Shytstem) is a configuration management system inspired by Ansible and Sean's desire to write more Go.

To build:
    go get gopkg.in/yaml.v2 "github.com/pkg/sftp"
    go build -o bin/crapms

To run:
    `bin/crapms -h hosts.yaml -u user -p pass config.yaml`
where hosts.yaml is a list of hosts to target
and config.yaml is a config file

To run Hello World example:
    `bin/crapms -h hello-world/hosts.yaml -p foobarbaz hello-world/config.yaml`

CrapMS uses YAML for configuration manifest files because it's more human-readable/writable than XML and more-commentable than JSON.

Currently only supports password-based auth because Go's ssh library only supports password auth.


TODO:
    - copy file source depth 1
    - invisible password prompt
    - parallel execution
    - more helpful logging
    - stop trying to implement rsync

