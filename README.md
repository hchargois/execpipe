# execpipe

execpipe:

* launches two commands, the output (stdout) of the first being piped into the input (stdin) of the second
* the second command is the child of the first
* there is no execpipe process that stays around; you only get the two processes corresponding to the commands

## Usage

    execpipe cmd1 [args...] | cmd2 [args...]

This is a literal pipe, so you must escape it if you launch execpipe from a shell. In bash, you can surround it by quotes (single or double), or put a backslash before it.

## Why?

This is useful to run a pipe of processes under a process manager (like supervisord). The process manager, when launching a command, remembers its PID to later send signals to it. If it's something unrelated to what you actually want to launch (like a bash process that manages the pipe), the signals sent by the process manager won't do what they are meant to.

For example, if you configure in supervisord:

    command=bash -c 'sleep 1000 | sleep 2000'

It will launch OK, but supervisor will "see" bash's PID as the PID of the command. If you try to stop the command, supervisor will send a signal to bash and bash will exit, but the sleeps will continue running.

execpipe fixes that by making the first command execute with the PID that supervisor sees. So if you tell supervisor to stop the command, it will send a signal to the right process (sleep 1000) that will stop.

If the second process is well-behaved and stops itself when its stdin is closed (remember it is linked to the stdout of the first process) then it should also stop when the first process stops.

## Example

    $ execpipe sleep 1000 \| sleep 2000 &
    [1] 5345
    $ ps f -C sleep
      PID TTY      STAT   TIME COMMAND
     5345 pts/12   S      0:00 sleep 1000
     5350 pts/12   S      0:00  \_ sleep 2000

Notice that:

* execpipe executes as PID 5345 that becomes the PID of our first command (sleep 1000)
* the first command is the parent of the second

## Competitors

### Naive bash

    $ bash -c 'sleep 1000 | sleep 2000' &
    [1] 8982
    $ ps f -C sleep,bash
      PID TTY      STAT   TIME COMMAND
     [...]
     8982 pts/12   S      0:00  bash -c sleep 1000 | sleep 2000
     8983 pts/12   S      0:00   \_ sleep 1000
     8984 pts/12   S      0:00   \_ sleep 2000

We get the PID of the bash process that just stays there. If we kill it, our sleeps get reparented but continue to live.

### Bash exec ninja hack

    $ bash -c 'exec 1> >(sleep 2000); exec sleep 1000' &
    [1] 5971
    $ ps f -C sleep,bash
      PID TTY      STAT   TIME COMMAND
     [...]
     5971 pts/12   S      0:00  sleep 1000
     5972 pts/12   S      0:00   \_ bash -c exec 1> >(sleep 2000); exec sleep 1000
     5973 pts/12   S      0:00       \_ sleep 2000

That's much better, but we have a bash process that stays around for nothing.

### [pipexec](https://github.com/flonatel/pipexec)

    $ pipexec [ A /usr/bin/sleep 1000 ] [ B /usr/bin/sleep 2000 ] "{A:1>B:0}" &
    [3] 9259
    $ ps f -C sleep,pipexec
      PID TTY      STAT   TIME COMMAND
     9259 pts/12   S      0:00 pipexec [ A /usr/bin/sleep 1000 ] [ B /usr/bin/sleep 2000 ] {A 1>B 0}
     9260 pts/12   S      0:00  \_ /usr/bin/sleep 1000
     9261 pts/12   S      0:00  \_ /usr/bin/sleep 2000

Of course pipexec is an awesome tool to do complex things, but for our simple use-case, it's no good. The PID we get from running the command is pipexec's PID, that's not what we want. And the command line is more complicated than it should be.

### Other?

If you think there is a more readily available solution (more popular tool? even better bash-fu?), please let me know!
