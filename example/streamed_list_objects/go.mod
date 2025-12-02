module streamed_list_objects

go 1.24.0

toolchain go1.25.4

// To reference published build, comment below and run `go mod tidy`
replace github.com/openfga/go-sdk v0.7.3 => ../../

replace github.com/openfga/go-sdk => ../.. // added this to point to local module
