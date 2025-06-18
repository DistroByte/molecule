# molecule

Molecule is a tool that scrapes a nomad server for allocations with exposed ports and then displays those URLs in a web interface.

It is intended to be used in a development environment to quickly find the URLs of services running in nomad.

It automatically fetches favicons from named services and displays them in the UI. It also supports restarting services from the web interface.

## See it in action

[molecule.dbyte.xyz](https://molecule.dbyte.xyz)

## Installation

See the sample [nomad job](https://github.com/DistroByte/nomad/blob/master/jobs/molecule.hcl) for deployment into a nomad cluster.
