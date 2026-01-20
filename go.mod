module github.com/kyleking/djot-fmt

go 1.23

replace github.com/sivukhin/godjot/v2 => ../godjot

require (
	github.com/sivukhin/godjot/v2 v2.0.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
