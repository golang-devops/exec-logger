# Purpose of exec-logger

This project is purely a "wrapper" around a normal command-line application. It adds functionality like timeouts, abort requests, time-stamp logging, etc

# Quick Start

## Execute the command

Specify the `-logfile`, `-task` and then lastly the actual command-line (like `ping` in the below example).

### Windows:

```
go get -u github.com/golang-devops/exec-logger
cd %Temp% & mkdir test-exec-logger & cd test-exec-logger
exec-logger -logfile tmp.log -task exec ping 127.0.0.1 -n 3
```

### Linux:

```
go get -u github.com/golang-devops/exec-logger
cd $Temp && mkdir test-exec-logger && cd ./test-exec-logger
exec-logger -logfile tmp.log -task exec ping 127.0.0.1 -c 3
```

## Inspect created files

There should now be three files in this temp dir, namely:

- `alive.txt` - gets written out every 2 seconds to inform "external observers" that the process is alive and responding.
- `exited.json` - gets written out when the process finished and contains the `ExitCode`, `Time` (of exit) and `Error` (if any)
- `tmp.log` - contains the stdout/stderr of the "wrapped command" which is that of the `ping` command in the above example

Note that the `tmp.log` file added time-stamp prefixes to each line received from `ping`. It also has additional information like the last line that should read **"Command exited with code 0"**.

The `exited.json` file should contain a zero `ExitCode`.

## Failure example

Delete the above three files if they already exist.

Change the above command to a "broken" one (for example by removing the number passed to the `-n` flag for windows or `-c` flag for linux. Broken windows command `exec-logger -logfile tmp.log -task exec ping 127.0.0.1 -n`. Broken linux command `exec-logger -logfile tmp.log -task exec ping 127.0.0.1 -c`.

Now the three files will be created again. This time the `exited.json` file should contain a non-zero `ExitCode` with an `Error`. The `tmp.log` file will contain a line or two with `EASY_EXEC_ERROR:` prefix after the timestamp. Those lines are the `Stderr` lines received by the `ping` command. The last of those error lines would also contain `Unable to run command` which is generated by the exec-logger itself.

## Abort the process prematurely

Delete the above three files if they already exist.

If an "external observer" of this process wants to abort it, it could just create a file called `must-abort.txt` inside this folder.

Run the same example from above but change the number flag (`-n` for windows or `-c` for linux) to something huge, like 3000. This means it will ping 3000 times. But we will abort this process using this mentioned file.

After running that 3000 command open another console window in the same temp directory and call `echo "" > must-abort.txt` to create the empty file. Now watch the first console windows should shortly after that abort the ping command. After aborting the `tmp.log` file will contain a line with `Got ABORT message` and another with `Successfully killed process with PID`. The `exited.json` will also contain a non-zero `ExitCode`.

## Auto time-out using a duration

Delete the above three files if they already exist.

Now to only allow the command to run for a certain time we can use the `-timeout-kill` flag. Change the above command to something like:

- windows: `exec-logger -logfile tmp.log -timeout-kill 5s -task exec ping 127.0.0.1 -n 3000`
- linux: `exec-logger -logfile tmp.log -timeout-kill 5s -task exec ping 127.0.0.1 -c 3000`

Note the added flag `-timeout-kill 5s`. This will attempt to run ping for 3000 seconds but we tell it to abort after `5s`. This timeout flag is in the format of native golang [time.Duration](https://golang.org/pkg/time/#Duration).

After 5 seconds it should automatically abort and our usual three files should be there. The `exited.json` file should again have a non-zero `ExitCode` with its error being something like "The command timed out after '5s'". The `tmp.log` will also contain a line reading `Timeout of 5s reached, now aborting` as well as `Successfully killed process with PID`.

# Contribution

Pull requests are welcome.