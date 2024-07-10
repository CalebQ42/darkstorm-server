package db

/*
TODO
Currently there isn't a easy and clean way to implement this (as far as I can tell).
valkey-go relies on an internal library for it's command builder, which makes it impossible to
use properly for generics without manually writing out the Index command. I could probably do this, but
it's a pain.
valkey-go does have a Generic Object Mapping library (valkey-go/om), but it requires a Version field
on every struct which would be confusing if I did add it to all my structs and Go doesn't allow anonymous generics inside structs
*/