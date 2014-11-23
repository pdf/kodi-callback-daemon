# Kodi/XBMC Callback Daemon
A small Go daemon that reads notifications from Kodi/XBMC via the JSON-RPC socket, and performs actions based on those notifications.

I wrote this primarily because the Python callback interface can get blocked very easily by any add-on, which results in heavy delays getting callbacks executed (for example, using the [service.xbmc.callbacks](https://github.com/pilluli/service.xbmc.callbacks) plugin), whereas notifications are shipped over the JSON interface immediately.

This is not an issue with `service.xbmc.callbacks` (good work `pilluli`!), but with the Kodi add-on infrastructure, and with other individual add-ons.

This daemon also aims to provide more flexibility. The targeted backends are [hyperion](https://github.com/tvdzwan/hyperion), `kodi` and `shell`.

## Upgrading to v1.0.0
Due to the name change to keep in line with XBMC's new name Kodi, if you have the previous package installed as `xbmc-callback-daemon`, you should uninstall it before installing the new `kodi-callback-daemon`.  The init scripts should handle your existing `/etc/xbmc-callback-daemon.json` config file, but you're encouraged to move this to the new location `/etc/kodi-callback-daemon.json` as this feature may go away in future to avoid confusion.

## Backends
### Hyperion
The Hyperion backend submits callbacks via the JSON interface. This interface is also used by the `hyperion-remote` command-line utility. There's no end-user documentation for this interface, so when writing callbacks, your best bet is to simply read the [JSON schemas](https://github.com/tvdzwan/hyperion/tree/master/libsrc/jsonserver/schema) in the source tree.

### Kodi

The Kodi backend submits callbacks via the JSON-RPC interface. There is excellent documentation available in the [Kodi wiki](http://kodi.wiki/view/JSON-RPC_API).

### Shell
The shell backend simply executes a command on the system with specified arguments.

## Installation
Grab the [Latest Release](https://github.com/pdf/kodi-callback-daemon/releases/latest) as a compiled binary and either install it using your package manager (Debian/Ubuntu/derivs, via the `.deb` package), or extract `kodi-callback-daemon` to somewhere on your path (eg - `/usr/local/bin`) on Linux/OSX/FreeBSD, or where ever on Windows.

Alternatively, you make clone this repository and build it yourself.

## Usage

_NB: I've not actually tested Windows/OSX/FreeBSD support at all, feel free to submit bug reports_

### Configure Kodi/XBMC
You must configure Kodi to accept remote control request, by enabling the options:

- `Allow programs on this system to control Kodi`
- `Allow programs on other systems to control Kodi`

In Kodi settings at `System Settings` -> `Network` -> `Remote Control`

### Run manually
Linux/OSX/FreeBSD:

```bash
/path/to/bin/kodi-callback-daemon /path/to/configFile.json
```

Windows:

```
C:\Path\To\kodi-callback-daemon.exe C:\Path\To\configFile.json
```

### Debian/Ubuntu/derivs
The deb packages include SysV and Upstart init scripts - enable and use them in the standard fashion.  You will need to add your configuration file at:

```
/etc/kodi-callback-daemon.json
```

Alternatively, you may edit `/etc/default/kodi-callback-daemon` and set the path to your configuration file there.

### Kodi autoexec.py
You might alternatively start the daemon from Kodi's `autostart.py`.  Simply edit `userdata/autoexec.py` in your Kodi directory (ie `~/.kodi/userdata/autoexec.py` on \*nix systems), and add the following:

Linux/OSX/FreeBSD:

```python
import kodi
import subprocess

subprocess.Popen(['/path/to/bin/kodi-callback-daemon', '/path/to/configFile.json'])
```

Windows:

```python
import kodi
import subprocess

subprocess.Popen(['C:\\Path\\To\\kodi-callback-daemon.exe', 'C:\\Path\\To\\configFile.json'])
```

Note the double-slashes necessary for escaping the Windows paths in Python strings.

_Note: If you're using this method, you'll also want to make sure that the daemon is killed on Kodi exit or startup, otherwise you'll get multiple copies running and they'll fight for resources.  I'm not a Python guy, so I'm open to suggestions on how to best handle this._

## User support
If you have questions on how to use the daemon, you may post them in the [Kodi forum thread](http://forum.kodi.tv/showthread.php?tid=194910).

## Configuration
The configuration file is written in JSON (I know, JSON is awful for configuration, but since we're passing JSON messages everywhere, it makes the most sense here), and has the following top-level members:

- `kodi` connection object (required)
- `hyperion` connection object (optional, but required if you're using the Hyperion backend)
- `debug` boolean enables debug logging (optional)
- `callbacks` object (required, or nothing will be done!).

See the [config.example.json](https://github.com/pdf/kodi-callback-daemon/tree/master/contrib/config.example.json) for my Hyperion setup, which uses most of the available features.

### Kodi/XBMC Connection
Specify your Kodi/XBMC IP address and port for the JSON interface in the `kodi` property:

```json
{
  "kodi": {
    "address": "127.0.0.1",
    "port": 9090
  }
}
```

### Hyperion Connection
If you're using the Hyperion backend, specify your Hyperion address and port for the JSON interface in the `hyperion` property:

```json
{
  "kodi": {
    "address": "127.0.0.1",
    "port": 9090
  },
  "hyperion": {
    "address": "127.0.0.1",
    "port": 19444
  }
}
```

### Debug logging
You can enable debug logging by setting the debug property to `true`:

```json
{
  "kodi": {
    "address": "127.0.0.1",
    "port": 9090
  },
  "hyperion": {
    "address": "127.0.0.1",
    "port": 19444
  },
  "debug": true
}
```

### Callbacks
The callbacks object is keyed by the Kodi notification method, with each method containing an array of callback objects. There is one exception, which is the `Startup` method - any callbacks attached to this method will be executed when the daemon starts up. Each callback in the array will be executed sequentially. The callback must contain a `backend` property, the value of which is one of `["hyperion", "kodi", "shell"]`. All other properties are backend-specific.

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

(see the [Hyperion JSON schemas](https://github.com/tvdzwan/hyperion/tree/master/libsrc/jsonserver/schema) for details on the required fields).

And if we wanted to run this callback on `Startup`, and on `Player.OnStop` notifications, our full configuration might look like this:

```json
{
  "kodi": {
    "address": "127.0.0.1",
    "port": 9090
  },
  "hyperion": {
    "address": "127.0.0.1",
    "port": 19444
  },
  "debug": true,
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

#### Kodi
Callbacks using the `kodi` backend contain the `backend` property, a `method` string property containing the Kodi RPC method to call, and a `params` object property containing the parameters to for the RPC call. An example might look like:

```json
{
  "backend": "kodi",
  "method": "GUI.ShowNotification",
  "params": {
    "title": "Callback Daemon",
    "message": "Hello from the callback daemon!",
    "displaytime": 15000
  }
}
```

Full example, mixing `hyperion`, `kodi` and `shell` callbacks:

```json
{
  "kodi": {
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
      },
      {
        "backend": "kodi",
        "method": "GUI.ShowNotification",
        "params": {
          "title": "Callback Daemon",
          "message": "Hello from the callback daemon!",
          "displaytime": 15000
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

For details on the available RPC methods, see the [Kodi wiki page](http://kodi.wiki/view/JSON-RPC_API/v6).  Unfortunately, because the Kodi JSON-RPC implementation is non-conformant, you will get errors logged when using RPC methods that return strings instead of conformant JSON-RPC objects, but the calls will execute fine.

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
  "kodi": {
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
The `Player.OnPlay` notification has one additional, optional property available to callbacks: `types`. This property may contain an array of item types sent with Kodi notifications with this method. At the time of writing, these types are `["movie", "episode", "song"]`.  Callbacks with a `types` property will only execute if the played media type matches one of the listed types in the callback.  Callbacks with no `types` property are always executed on `Player.OnPlay` notifications.  The following example increases Hyperion saturation/value, and decreases gamma compensation for music so that visualizations produce punchy lighting effects, and conversely sets much more sedate values for video types.  It also executes a `clear` command on channel 86 when any media is played (`types` is omitted).

```json
{
  "kodi": {
    "address": "127.0.0.1",
    "port": 9090
  },
  "hyperion": {
    "address": "127.0.0.1",
    "port": 19444
  },
  "debug": true,
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
