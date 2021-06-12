#!/bin/bash
echo Setting production dependencies...
go mod edit -dropreplace github.com/semog/go-bot-api/v4
go mod edit -dropreplace github.com/semog/go-common
go mod edit -dropreplace github.com/semog/go-sqldb
