<!--
SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>

SPDX-License-Identifier: BSD-3-Clause
-->

# ansua

ansua is a command line interface to help tracking time.

## SYNOPSIS

    ansua <duration> [ "<command>" ]

## DESCRIPTION

ansua run a timer on defined duration and optionally run a command when
timer finished.

When ansua timer is running, one can pause the timer by pressing p+Enter,
and resume it by pressing r+Enter, or stopping it using CTRL+c.

## PARAMETERS

The duration parameter is using the "XhYmZs" format, where "h" represent
hours, "m" represent minutes, and "s" represent seconds.
For example, "1h30m" equal to one hour 30 minutes, "10m30" equal to 10
minutes and 30 seconds.

The command parameter is optional.

## EXAMPLE

Run timer for 1 minute,

    $ ansua 1m

Run timer for 1 hour and then display notification using notify-send in
GNU/Linux,

    $ ansua 1h notify-send "ansua completed"

## LINKS

Repository: https://git.sr.ht/~shulhan/pakakeh.go

Issue: https://todo.sr.ht/~shulhan/pakakeh.go
