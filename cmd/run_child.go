//go:build darwin

package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mnkd/cred/internal/mask"
)

// forwardedSignals are passed on to the command. This is the set op run
// forwards. SIGTSTP and SIGCONT are left out: the terminal sends them to the
// whole process group, which includes the command.
var forwardedSignals = []os.Signal{
	syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT,
	syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGWINCH,
}

// drainTimeout bounds how long output is still read after the command exits.
// Waiting for EOF instead would let a background process that inherited the
// pipes keep cred running, and cred ignores SIGTERM while it forwards it.
const drainTimeout = time.Second

// childIO is the stdio cred hands to the command. Stdin is passed as is;
// Stdout and Stderr receive the masked output.
type childIO struct {
	Stdin, Stdout, Stderr *os.File
}

// runMasked runs the command as a child, masks secrets in its stdout and
// stderr, forwards signals and returns its exit code (128+n if it was killed
// by signal n).
func runMasked(binary string, args, environ, secrets []string, stdio childIO) (int, error) {
	// Before Start, so a signal that arrives in between is not lost.
	sigs := make(chan os.Signal, 16)
	signal.Notify(sigs, forwardedSignals...)
	defer signal.Stop(sigs)

	// One pipe for both streams when they go to the same file (a terminal,
	// or 2>&1), so their order is kept.
	shared := sameFile(stdio.Stdout, stdio.Stderr)
	streams := []*os.File{stdio.Stdout}
	if !shared {
		streams = append(streams, stdio.Stderr)
	}
	var readers, writers []*os.File
	closeAll := func(files []*os.File) {
		for _, f := range files {
			f.Close()
		}
	}
	for range streams {
		r, w, err := os.Pipe()
		if err != nil {
			closeAll(readers)
			closeAll(writers)
			return 0, fmt.Errorf("failed to create pipe: %w", err)
		}
		readers, writers = append(readers, r), append(writers, w)
	}

	cmd := &exec.Cmd{
		Path:   binary,
		Args:   args,
		Env:    environ,
		Stdin:  stdio.Stdin,
		Stdout: writers[0],
		Stderr: writers[len(writers)-1],
	}
	err := cmd.Start()
	closeAll(writers)
	if err != nil {
		closeAll(readers)
		return 0, fmt.Errorf("failed to start %s: %w", args[0], err)
	}

	var wg sync.WaitGroup
	for i, r := range readers {
		wg.Go(func() { copyMasked(streams[i], r, secrets) })
	}

	// Stdin, Stdout and Stderr are all files, so Wait returns when the
	// process exits, without waiting for the pipes.
	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()
	var waitErr error
	for exited := false; !exited; {
		select {
		case sig := <-sigs:
			// Fails only when the process has already exited.
			_ = cmd.Process.Signal(sig)
		case waitErr = <-waited:
			exited = true
		}
	}

	deadline := time.Now().Add(drainTimeout)
	for _, r := range readers {
		_ = r.SetReadDeadline(deadline)
	}
	wg.Wait()

	if cmd.ProcessState == nil {
		return 0, fmt.Errorf("failed to wait for %s: %w", args[0], waitErr)
	}
	if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal()), nil
	}
	return cmd.ProcessState.ExitCode(), nil
}

// copyMasked copies src to dst with secrets masked until src ends or its read
// deadline passes, then writes the held-back tail. On a write error it stops
// and closes src, so the command gets EPIPE on its next write.
func copyMasked(dst io.Writer, src *os.File, secrets []string) {
	defer src.Close()
	mw := mask.NewWriter(dst, secrets)
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := mw.Write(buf[:n]); werr != nil {
				return
			}
		}
		if err != nil {
			break
		}
	}
	_ = mw.Flush()
}

func sameFile(a, b *os.File) bool {
	ai, err := a.Stat()
	if err != nil {
		return false
	}
	bi, err := b.Stat()
	if err != nil {
		return false
	}
	return os.SameFile(ai, bi)
}
