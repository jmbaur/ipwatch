# IPWatch

Runs a script on IP address changes for a given network interface(s).

A common use case for this is for dynamically updating DNS records (e.g.
Cloudflare). This program will pass down its environment into the spawned
executable so that any secrets needed for updating external services can be
present.
