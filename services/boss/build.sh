#!/usr/bin/env bash
GO=~build/builds/go/prod/bin/go
ln -s ~build/builds/boss-ui/prod/boss-ui boss-ui
cd boss
${GO} build
