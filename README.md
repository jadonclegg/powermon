# powermon
powermon is a utility for managing computers attached to battery backups when the power goes out, handling automatic shutdown if the power goes out, and can send WOL packets to power on the same computers.

The idea is to run the server on a computer that's NOT attached to backup power, so it goes down when the power goes out. (I'm using a Raspberry PI for this).
When the clients fail to 'ping' the server, they will shut down automatically after the specified time.
When the server comes back online, it will send WOL packets to each client listed, and turn them back on, and optionally keep sending those packets until it's verified that they are back online.

# Usage
There are three sub commands to the powermon utility, we'll go over them here.

## Global Options
These options are ignored for the `powermon mac` command.

### Generic options
`-v [--verbose]` - enable verbose output

`-l [--logfile]` - specifies a file to log to

`-n [--nickname]` - nickname for the computer used in logging, and Pushover notifications.

### Pushover options:
`-k [--pushover-token]` - specifies the Pushover API token to use

`-u [--user-token]` - specifies the user token to send Pushover notifications to. (can be specified more than once, i.e. `-u <token> -u <second token>` etc.)

## Commands

### powermon mac
`powermon mac`

The mac command just lists the mac addresses and their interface names, so you can easily add them to the -w [--wake] option.

### powermon server
`powermon [Global Options] server [Server Options]`

`-p [--port]` Port to listen on. Default is port 10101

`-w [--wake]` MAC address to send WOL packet to upon startup. Can be specified more than one time.

`--wakelist` A file with a list of one mac address per line to be woken up with WOL packets upon server startup. Empty lines are ignored, as well as lines beinning with # (comments)

`--verify` Keep sending WOL packets to MAC addresses specified with `-w [--wake]` or `--wakelist` until they have sent a ping to the server.

### powermon client
`powermon [Global Options] client [Client Options]`

`-a [--address]` Address of server. Can be an IP or a domain name.

`-p [--port]` Port to send pings to, make sure this is the same port the server is listening on. Defaults to 10101

`-t [--timeout]` Shuts the computer down X seconds after a ping fails. Cancelled if a ping succeeds. Default is 60 

`-i [--interval]` Ping the server every X seconds. Default is 60

# Compiling
Just run `go build` after cloning the project.

# Make commands
`make` - runs the default command `make powermon` - basically an alias to `go build`

`make arm` - builds for the arm (32bit) architecture, output file is powermon-arm (Raspberry Pi 0, 1? not verified)

`make arm64` - builds for the arm (64bit) architecture, output file is powermon-arm64 (Verified to work on my raspberry pi 4 running ubuntu server 20.04)

You can also just install go on your raspberry pi and run `go build`.
