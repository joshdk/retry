[![License][license-badge]][license-link]

# Retry

⏰ Rerun a command until it eventually succeeds, or doesn't!

## Installing

### Release binary

A prebuilt [release](https://github.com/joshdk/retry/releases) binary can be downloaded by running:

```bash
$ wget -q https://github.com/joshdk/retry/releases/download/v1.0.0/retry-linux-amd64.tar.gz
$ tar -xf retry-linux-amd64.tar.gz
$ sudo install retry /usr/bin/retry
```

### From source

Alternatively, a development version of this tool by running:

```bash
$ go get -u github.com/joshdk/retry
```

## Motivations

I kept seeing folks write bespoke code to retry commands that were either flaky, or took time to succeed. This usually manifested as some sort of loop in bash, with a counter, and a return code check.

Searching around, this doesn't seem to be an isolated problem, which has an even larger number of bespoke solutions. Take for example this handful of Stack Overflow threads:

- [How to retry a command in Bash?](https://stackoverflow.com/questions/7449772/how-to-retry-a-command-in-bash)
- [Retry a Bash command with timeout](https://stackoverflow.com/questions/12321469/retry-a-bash-command-with-timeout)
- [How do I write a retry logic in script to keep retrying to run it upto 5 times?](https://unix.stackexchange.com/questions/82598/how-do-i-write-a-retry-logic-in-script-to-keep-retrying-to-run-it-upto-5-times)

These are perfectly legitimate questions, with many reasonable answers. The downside is that the solutions were usually specific to the question asked, and not always applicable to the broader problem.

This tool is an attempt to solve this broader problem. ⏰

## Usage

### Help!

```bash
$ retry -help
Usage of retry:
  -attempts int
    	maximum number of attempts (default 3)
  -backoff
    	use exponential backoff when sleeping
  -consecutive int
    	required number of back to back successes
  -max-time duration
    	maximum total time (default 1m0s)
  -quiet
    	silence all output
  -sleep duration
    	time to sleep between attempts (default 5s)
  -task-time duration
    	maximum time for a single attempt
  -version
    	print the version "1.0.0" and exit
```

### Running a command

Retry will run a given command repeatedly, until it is deemed an overall success of failure. The conditions and limits for what determine success/failure can be tuned with command line flags.

As a special case, if a URL is given, retry will GET that URL and check for a 200 OK to be returned.

### Limit attempts

The `-attempts` flag limits the maximum number of times a command can be run.

> Run `cat kubeconfig.yml` a maximum of 3 times, or less if the command succeeds earlier:
>
> ```bash
> $ retry -attempts=3 cat kubeconfig.yml
> ```

### Limit task time

The `-task-time` flag limits the maximum time that a command can run for. 

> Run `wget https://example.com`, but limit the command to only run for a maximum of 15 seconds.
>
> ```bash
> $ retry -task-time=15s wget https://example.com
> ```

### Limit overall time

The `-max-time` flag limits the maximum total time that `retry` will run for.

> GET `https://example.com` repeatedly, but stop running after a total of 60 seconds.
>
> ```bash
> $ retry -max-time=60s https://example.com
> ```

### Sleep between attempts

The `-sleep` flag inserts a timed delay between command runs.

> Run `cat kubeconfig.yml`, but sleep for 15 seconds between runs.
>
> ```bash
> $ retry -sleep=15s cat kubeconfig.yml
> ```

### Exponential backoff

The `-backoff` flag is used with `-sleep`, and will double the time delay between failed runs. Delay is reset after a successful run.

> Run `wget https://example.com`, sleeping for 15 seconds after the first failure, 30 seconds after the second failure, 1 minute after the third failure, etc...
>
> ```bash
> $ retry -sleep=15s -backoff wget https://example.com
> ```

### Consecutive successes

The `-consecutive` flag requires a number of successful command runs to occur in a row in order to be considered successful. Useful for health checking a service that is inconsistent until if if fully started.

> GET `https://example.com`, requiring the command to be successful 3 times in a row.
>
> ```bash
> $ retry -consecutive=3 wget https://example.com
> ```

### Be quiet!

Lastly, the `-quiet` flag silences all output (STDOUT and STDERR) from the command. Useful when running `retry` inside an `if`.

> Run `ls -R`, but swallow all output.
>
> ```bash
> $ retry -quiet ls -R
> ```

### Altogether now

> Run `wget https://example.com` a maximum of **10** times. Each run can take a maximum of **15 seconds**, and a total of **2 minutes**. Sleep for **5 seconds** between failures with exponential **backoff**. Lastly, require that the command succeeds **3 times** in a row.
>
> ```bash
> $ retry -attempts=10 -task-time=15s -max-time=2m -sleep=5s -backoff -consecutive=3 wget https://example.com
>```

## License

This code is distributed under the [MIT License][license-link], see [LICENSE.txt][license-file] for more information.

[license-badge]:    https://img.shields.io/badge/license-MIT-green.svg
[license-file]:     https://github.com/joshdk/retry/blob/master/LICENSE.txt
[license-link]:     https://opensource.org/licenses/MIT
