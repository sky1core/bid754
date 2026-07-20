module github.com/sky1core/bid754/benchcompare-go

go 1.23

require (
	github.com/shopspring/decimal v1.4.0
	github.com/sky1core/bid754/bid754-codec-go v0.0.0
	github.com/sky1core/bid754/bid754-go v0.0.0
)

replace github.com/sky1core/bid754/bid754-go => ../bid754-go

replace github.com/sky1core/bid754/bid754-codec-go => ../bid754-codec-go
