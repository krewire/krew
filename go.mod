module github.com/krewire/krew

go 1.27.1

require (
	github.com/krewire/auth v0.1.0
	github.com/krewire/krewire v0.1.0
)

replace github.com/krewire/auth => ../auth
replace github.com/krewire/krewire => ../krewire
