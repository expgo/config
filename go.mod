module github.com/expgo/config

go 1.20

replace (
	github.com/expgo/structure => /home/mind/peace/expgo/structure
)

require (
	github.com/expgo/structure v0.0.0-20240430162200-172eb0aa3964
	github.com/fsnotify/fsnotify v1.7.0
	github.com/stretchr/testify v1.9.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/expgo/sync v0.0.0-20240416034417-7c4de7477076 // indirect
	github.com/petermattis/goid v0.0.0-20240503122002-4b96552b8156 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sasha-s/go-deadlock v0.3.1 // indirect
	golang.org/x/sys v0.19.0 // indirect
)
