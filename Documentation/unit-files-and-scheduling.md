# Unit Files

Unit files are the primary means of interacting with fleet. They define what you want to do, and how fleet should do it.

fleet will schedule any valid service, socket, path or timer systemd unit to a machine in the cluster, taking into account a few special properties in the `[X-Fleet]` section. If you're new to using systemd unit files, check out the [Getting Started with systemd guide](https://coreos.com/docs/launching-containers/launching/getting-started-with-systemd).

## Unit Requirements

* Must be a supported unit type: `service`, `socket`, `device`, `mount`, `automount`, `timer`, `path`
* Each unit file must have a file extension corresponding to its respective unit type.

## fleet-specific Options

| Option Name | Description |
|---------------|-------------|
| `X-ConditionMachineID` | Require the unit be scheduled to the machine identified by the given string. |
| `X-ConditionMachineOf` | Limit eligible machines to the one that hosts a specific unit. |
| `X-ConditionMachineMetadata` | Limit eligible machines to those with this specific metadata. |
| `X-Conflicts` | Prevent a unit from being collocated with other units using glob-matching on the other unit names. |
| `Global` | Schedule this unit on all agents in the cluster. Should not be used with other options. | 

See [more information](#unit-scheduling) on these parameters and how they impact scheduling decisions.

Take the following as an example of how your `[X-Fleet]` section could be written:

```
[Unit]
Description=Some Monitoring Service

[Service]
ExecStart=/bin/monitorme

[X-Fleet]
X-ConditionMachineMetadata=location=chicago
X-Conflicts=monitor*
```

## systemd specifiers

When evaluating the `[X-Fleet]` section, fleet supports a subset of systemd's [specifiers][systemd specifiers] to perform variable substitution. The following specifiers are currently supported:

* `%n`
* `%N`
* `%p`
* `%i`

For the meaning of the specifiers, refer to the official [systemd documentation][systemd specifiers].

[systemd specifiers]: http://www.freedesktop.org/software/systemd/man/systemd.unit.html#Specifiers


# Unit Scheduling

## Making Scheduling Decisions

The fleet engine is responsible for deciding where units should be scheduled in the cluster. The current algorithm for making this determination is incredibly simple and only considers the number of units that each agent is running; it will attempt to schedule a unit on the agent with the lowest number of units.

Read more about [fleet's architecture and data model](https://github.com/coreos/fleet/blob/master/Documentation/architecture.md).

## User-Defined Requirements

##### Schedule unit to specific machine

The `X-ConditionMachineID` option of a unit file causes the system to schedule a unit to a machine identified by the option's value.

The ID of each machine is currently published in the `MACHINE` column in the output of `fleetctl list-machines -l`.
One must use the entire ID when setting `X-ConditionMachineID` - the shortened ID returned by `fleetctl list-machines` without the `-l` flag is not acceptable.

fleet depends on its host to generate an identifier at `/etc/machine-id`, which is handled today by systemd.
Read more about machine IDs in the [official systemd documentation][machine-id].

[machine-id]: http://www.freedesktop.org/software/systemd/man/machine-id.html

##### Schedule unit to machine with specific metadata

The `X-ConditionMachineMetadata` option of a unit file allows you to set conditional metadata required for a machine to be elegible.

```
[X-Fleet]
X-ConditionMachineMetadata="region=us-east-1" "diskType=SSD"
```

This requires an eligible machine to have at least the `region` and `diskType` keys set accordingly. A single key may also be defined multiple timess, in which case only one of the conditions needs to be met:

```
[X-Fleet]
X-ConditionMachineMetadata=region=us-east-1
X-ConditionMachineMetadata=region=us-west-1
```

This would allow a machine to match just one of the provided values to be considered eligible to run.

A machine is not automatically configured with metadata.
A deployer may define machine metadata using the `metadata` [config option](https://github.com/coreos/fleet/blob/master/Documentation/deployment-and-configuration.md).

##### Schedule unit next to another unit

In order for a unit to be scheduled to the same machine as another unit, a unit file can define `X-ConditionMachineOf`.
The value of this option is the exact name of another unit in the system, which we'll call the target unit.

If the target unit is not found in the system, the follower unit will be considered unschedulable. 
Once the target unit is scheduled somewhere, the follower unit will be scheduled there as well.

Follower units will reschedule themselves around the cluster to ensure their `X-ConditionMachineOf` options are always fulfilled.

##### Schedule unit away from other unit(s)

The value of the `X-Conflicts` option is a [glob pattern](http://golang.org/pkg/path/#Match) defining which other units next to which a given unit must not be scheduled.

If a unit is scheduled to the system without an `X-Conflicts` option, other units' conflicts still take effect and prevent the new unit from being scheduled to machines where conflicts exist.

##### Dynamic requirements

fleet supports several systemd specifiers to allow requirements to be dynamically determined based on a Unit's name. This means that the same unit can be used for multiple Units and the requirements are dynamically substituted when the Unit is scheduled.

For example, a Unit by the name `foo.service`, whose unit contains the following snippet:

```
[X-Fleet]
X-ConditionMachineOf=%p.socket
```

would result in an effective `X-ConditionMachineOf` of `foo.socket`. Using the same unit snippet with a Unit called `bar.service`, on the other hand, would result in an effective `X-ConditionMachineOf` of `bar.socket`.

For more information on the available specifiers, see the [unit file configuration](unit-files.md#systemd-specifiers) documentation.