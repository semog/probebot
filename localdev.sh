#!/bin/bash
echo Setting local development dependencies...
go mod edit -replace github.com/semog/go-bot-api/v4=../go-bot-api
go mod edit -replace github.com/semog/go-common=../go-common
go mod edit -replace github.com/semog/go-sqldb=../go-sqldb
