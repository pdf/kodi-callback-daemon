# XBMC Callback Daemon
A small Go daemon that reads notifications from XBMC via the JSON-RPC socket, and performs actions based on those notifications.

I wrote this primarily because the Python callback interface can get blocked very easily by any add-on, which results in heavy delays getting callbacks executed (for example, using the [service.xbmc.callbacks](https://github.com/pilluli/service.xbmc.callbacks) plugin), whereas notifications are shipped over the JSON interface immediately.

This is not an issue with `service.xbmc.callbacks` (good work `pilluli`!), but with the XBMC add-on infrastructure, and with other individual add-ons.

This daemon also aims to provide more flexibility. The targeted backends are [hyperion](https://github.com/tvdzwan/hyperion), `xbmc` and `shell`. Currently, `hyperion` and `shell` are implemented, `xbmc` support will come if/when I get motivated to deal with it.

## Backends
### Hyperion
The Hyperion backend submits callbacks via the JSON interface. This interface is also used by the `hyperion-remote` command-line utility. There's no end-user documentation for this interface, so when writing callbacks, your best bet is to simply read the [JSON schemas](https://github.com/tvdzwan/hyperion/tree/master/libsrc/jsonserver/schema) in the source tree.

### XBMC
__NOTE__: _This backend is not yet implemented_

The XBMC backend submits callbacks via the JSON-RPC interface. There is excellent documentation available in the [XBMC wiki](http://wiki.xbmc.org/?title=JSON-RPC_API).

### Shell
The shell backend simply executes a command on the system with specified arguments.

## Installation
Grab the [Latest Release](https://github.com/pdf/xbmc-callback-daemon/releases/latest) as a compiled binary and either install it using your package manager (Debian/Ubuntu/derivs, via the `.deb` package), or extract `xbmc-callback-daemon` to somewhere on your path (eg - `/usr/local/bin`) on Linux/OSX/FreeBSD, or where ever on Windows.

Alternatively, you make clone this repository and build it yourself.

## Usage

_NB: I've not actually tested Windows/OSX/FreeBSD support at all, feel free to submit bug reports_

### Run manually
Linux/OSX/FreeBSD:

```bash
/path/to/bin/xbmc-callback-daemon /path/to/configFile.json
```

Windows:

```
C:\Path\To\xbmc-callback-daemon.exe C:\Path\To\configFile.json
```

### Upstart
There are Upstart init scripts available in [contrib](https://github.com/pdf/xbmc-callback-daemon/tree/master/contrib/upstart).  Just copy `xbmc-callback-daemon.conf` to `/etc/init/xbmc-callback-daemon.conf` and `default` to `/etc/default/xbmc-callback-daemon`, then place your config file at `/etc/xbmc-callback-daemon.json`.  Contributions welcome for SysV/SystemD/etc.  The Upstart job assumes you are starting XBMC via an the `xbmc` Upstart job, otherwise adapt as necessary.

### XBMC autoexec.py
You might alternatively start the daemon from XBMC's `autostart.py`.  Simply edit `userdata/autoexec.py` in your XBMC directory (ie `~/.xbmc/userdata/autoexec.py` on \*nix systems), and add the following:

Linux/OSX/FreeBSD:

```python
import xbmc
import subprocess

subprocess.Popen(['/path/to/bin/xbmc-callback-daemon', '/path/to/configFile.json'])
```

Windows:

```python
import xbmc
import subprocess

subprocess.Popen(['C:\\Path\\To\\xbmc-callback-daemon.exe', 'C:\\Path\\To\\configFile.json'])
```

Note the double-slashes necessary for escaping the Windows paths in Python strings.

## Configuration
The configuration file is written in JSON (I know, JSON is awful for configuration, but since we're passing JSON messages everywhere, it makes the most sense here), and has three top-level members: `xbmc` (required), `hyperion` (optional, required if you're using the Hyperion backend), and `callbacks` (required, or nothing will be done!).

See the [config.example.json](https://github.com/pdf/xbmc-callback-daemon/tree/master/contrib/config.example.json) for my Hyperion setup, which uses most of the available features.

### XBMC Connection
Specify your XBMC IP address and port for the JSON interface in the `xbmc` property:

```json
{
  "xbmc": {
    "address": "127.0.0.1",
    "port": 9090
  }
}
```

### Hyperion Connection
If you're using the Hyperion backend, specify your Hyperion address and port for the JSON interface in the `hyperion` property:

```json
{
  "xbmc": {
    "address": "127.0.0.1",
    "port": 9090
  },
  "hyperion": {
    "address": "127.0.0.1",
    "port": 19444
  }
}
```

### Callbacks
The callbacks object is keyed by the XBMC notification method, with each method containing an array of callback objects. There is one exception, which is the `Startup` method - any callbacks attached to this method will be executed when the daemon starts up. Each callback in the array will be executed sequentially. The callback must contain a `backend` property, the value of which is one of `["hyperion", "xbmc", "shell"]`. All other properties are backend-specific.

#### Hyperion
Callbacks using the `hyperion` backend contain the `backend` property, and all other properties are sent verbatim as the request to Hyperion.  For example, to execute the `Rainbow swirl` effect, the callback would look something like this:

```json
{
  "backend": "hyperion",
  "command": "effect",
  "effect": {
    "name": "Rainbow swirl"
  }
}
```

(see the [Hyperion JSON scemas](https://github.com/tvdzwan/hyperion/tree/master/libsrc/jsonserver/schema) for details on the required fields).

And if we wanted to run this callback on `Startup`, and on `Player.OnStop` notifications, our full configuration might look like this:

```json
{
  "xbmc": {
    "address": "127.0.0.1",
    "port": 9090
  },
  "hyperion": {
    "address": "127.0.0.1",
    "port": 19444
  },
  "callbacks": {
    "Startup": [
      {
        "backend": "hyperion",
        "command": "effect",
        "priority": 86,
        "effect": {
          "name": "Rainbow swirl"
        }
      }
    ],
    "Player.OnStop": [
      {
        "backend": "hyperion",
        "command": "effect",
        "priority": 86,
        "effect": {
          "name": "Rainbow swirl"
        }
      }
    ]
  }
}
```

#### XBMC
__TODO__

#### Shell
Callbacks using the `shell` backend contain the `backend` property, a `command` string property containing the path to the executable to be run, an optional `arguments` array containing a list of arguments to be passed to the command, and an optional `background` property to allow forking a long-running process without waiting for it to return.  An example might look like:

```json
{
  "backend": "shell",
  "background": true,
  "command": "/bin/echo",
  "arguments": [
    "-e",
    "The 'arguments' property is optional\n"
  ]
}
```

Full example, mixing `hyperion` and `shell` callbacks:

```json
{
  "xbmc": {
    "address": "127.0.0.1",
    "port": 9090
  },
  "hyperion": {
    "address": "127.0.0.1",
    "port": 19444
  },
  "callbacks": {
    "Startup": [
      {
        "backend": "hyperion",
        "command": "effect",
        "priority": 86,
        "effect": {
          "name": "Rainbow swirl"
        }
      },
      {
        "backend": "shell",
        "background": true,
        "command": "/bin/echo",
        "arguments": [
          "-e",
          "The 'arguments' property is optional\n"
        ]
      }
    ],
    "Player.OnStop": [
      {
        "backend": "hyperion",
        "command": "effect",
        "priority": 86,
        "effect": {
          "name": "Rainbow swirl"
        }
      }
    ]
  }
}
```

#### Player.OnPlay
The `Player.OnPlay` notification has one additional, optional property available to callbacks: `types`. This property may contain an array of item types sent with XBMC notifications with this method. At the time of writing, these types are `["movie", "episode", "song"]`.  Callbacks with a `types` property will only execute if the played media type matches one of the listed types in the callback.  Callbacks with no `types` property are always executed on `Player.OnPlay` notifications.  The following example increases Hyperion saturation/value, and decreases gamma compensation for music so that visualizations produce punchy lighting effects, and conversely sets much more sedate values for video types.  It also executes a `clear` command on channel 86 when any media is played (`types` is omitted).

```json
{
  "xbmc": {
    "address": "127.0.0.1",
    "port": 9090
  },
  "hyperion": {
    "address": "127.0.0.1",
    "port": 19444
  },
  "callbacks": {
    "Player.OnPlay": [
      {
        "types": ["movie", "episode"],
        "backend": "hyperion",
        "command": "transform",
        "transform": {
          "gamma": [2.2, 2.2, 2.8],
          "valueGain": 1.0,
          "saturationGain": 1.0
        }
      },
      {
        "types": ["song"],
        "backend": "hyperion",
        "command": "transform",
        "transform": {
          "gamma": [0.8, 0.8, 0.8],
          "valueGain": 2.0,
          "saturationGain": 2.0
        }
      },
      {
        "backend": "hyperion",
        "command": "clear",
        "priority": 86
      }
    ]
  }
}
```
