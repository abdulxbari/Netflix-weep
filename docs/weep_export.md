## weep export

Retrieve credentials to be exported as environment variables

### Synopsis

The export command retrieves credentials for a role and prints a shell command to export 
the credentials to environment variables.

More information: https://hawkins.gitbook.io/consoleme/weep-cli/commands/credential-export


```
weep export [role_name] [flags]
```

### Options

```
  -h, --help   help for export
```

### Options inherited from parent commands

```
  -A, --assume-role strings        one or more roles to assume after retrieving credentials
  -c, --config string              config file (default is $HOME/.weep.yaml)
      --extra-config-file string   extra-config-file <yaml_file>
      --log-file string            log file path (default "/tmp/weep.log")
      --log-format string          log format (json or tty)
      --log-level string           log level (debug, info, warn)
  -n, --no-ip                      remove IP restrictions
  -r, --region string              AWS region (default "us-east-1")
```

### SEE ALSO

* [weep](weep.md)	 - weep helps you get the most out of ConsoleMe credentials

