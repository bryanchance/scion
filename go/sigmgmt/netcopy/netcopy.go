// Copyright 2018 Anapaya Systems

package netcopy

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	log "github.com/inconshreveable/log15"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/sigmgmt/db"
)

const (
	TimeBetweenVerificationRetries = time.Second
	TimeForHTTPRequest             = 2 * time.Second
)

func CopyFileToSite(ctx context.Context, spath string, site *db.Site, dpath string,
	logger log.Logger) error {

	for _, host := range site.Hosts {
		if err := CopyFileToHost(ctx, spath, &host, dpath, logger); err != nil {
			return common.NewBasicError("Unable to copy configuration to host:", err)
		}
	}
	return nil
}

func defaultSSHParams(ctx context.Context, host *db.Host) []string {
	var args []string
	// Explicitly set the private key, if one is specified
	if host.Key != "" {
		args = append(args, "-i", host.Key)
		// Only used supplied key
		args = append(args, "-o", "IdentitiesOnly=yes")
	}
	// Enable batch mode so we don't get stuck on prompts
	args = append(args, "-o", "BatchMode=yes")
	args = append(args, "-o", "StrictHostKeyChecking=no")
	// Do not allow the command to run for longer than the context; this is an
	// approximation, as some time might pass before the argument is parsed by
	// ssh/scp.
	if deadline, ok := ctx.Deadline(); ok {
		seconds := int(deadline.Sub(time.Now()).Seconds())
		args = append(args, "-o", fmt.Sprintf("ConnectTimeout=%d", seconds))
	}
	return args
}
func CopyFileToHost(ctx context.Context, spath string, host *db.Host, dpath string,
	logger log.Logger) error {

	// scp [ -i *sshKey ] spath [*sshUser@]host:dpath
	args := defaultSSHParams(ctx, host)
	args = append(args, spath)

	if host.User == "" {
		args = append(args, fmt.Sprintf("%s:%s", host.Name, dpath))
	} else {
		args = append(args, fmt.Sprintf("%s@%s:%s", host.User, host.Name, dpath))
	}
	scpCommand := exec.CommandContext(ctx, "scp", args...)
	logger.Info("Copying SIG config file via scp", "args", args)
	return runCommand(scpCommand)
}

func ReloadSite(ctx context.Context, site *db.Site, logger log.Logger) error {
	errors := make(chan error, len(site.Hosts))
	for _, host := range site.Hosts {
		host := host
		go func() {
			errors <- ReloadHost(ctx, &host, logger)
		}()
	}

	// Take the result from each goroutine
	ok := false
Loop:
	for range site.Hosts {
		select {
		case err := <-errors:
			if err == nil {
				ok = true
				break Loop
			} else {
				logger.Info("SIG reload error", "err", err)
			}
		case <-ctx.Done():
			return common.NewBasicError("context expired", nil, "site", site.Name)
		}
	}
	if !ok {
		return common.NewBasicError("SIG could not be reloaded", nil, "site", site.Name)
	}
	return nil
}

func ReloadHost(ctx context.Context, host *db.Host, logger log.Logger) error {
	// ssh [ -i *sshKey ] [*sshUser@]host sudo systemctl --signal=SIGHUP kill sig.service
	args := defaultSSHParams(ctx, host)
	if host.User == "" {
		args = append(args, fmt.Sprintf("%s", host.Name))
	} else {
		args = append(args, fmt.Sprintf("%s@%s", host.User, host.Name))
	}
	args = append(args, "sudo systemctl reload sig.service")
	sshCommand := exec.CommandContext(ctx, "ssh", args...)
	logger.Info("Sending SIG reload signal via ssh", "args", args)
	return runCommand(sshCommand)
}

func runCommand(cmd *exec.Cmd) error {
	if output, err := cmd.CombinedOutput(); err != nil {
		return common.NewBasicError("command error", err, "output", string(output))
	}
	return nil
}

func VerifyConfigVersion(ctx context.Context, site *db.Site, version uint64,
	logger log.Logger) error {

	errors := make(chan error, len(site.Hosts))
	for _, host := range site.Hosts {
		host := host
		go func() {
			errors <- verifyConfigVersionHost(ctx, &host, site.MetricsPort, version, logger)
		}()
	}
	// Take the result from each goroutine
	ok := false
Loop:
	for range site.Hosts {
		select {
		case err := <-errors:
			if err == nil {
				ok = true
				break Loop
			} else {
				logger.Info("host verification error", "err", err)
			}
		case <-ctx.Done():
			return common.NewBasicError("context done", ctx.Err(), "site", site.Name)
		}
	}
	if !ok {
		return common.NewBasicError("no host could be verified", nil)
	}
	return nil
}

// verifyConfigVersionHost attempts multiple version verifications until one
// suceeds, for as long as ctx allows.
func verifyConfigVersionHost(ctx context.Context, host *db.Host, port uint16, version uint64,
	logger log.Logger) error {

	url := fmt.Sprintf("http://%s:%d/configversion", host.Name, port)
	log.Debug("Trying to fetch metric page", "url", url)
	for try := 0; ; try++ {
		if err := verifyConfigVersionURL(ctx, url, version); err != nil {
			logger.Warn("Verification attempt failed", "host", host, "attempt", try,
				"error", err)
		} else {
			return nil
		}
		select {
		case <-ctx.Done():
			return common.NewBasicError("verification failed; check warnings for more information:",
				nil)
		case <-time.After(TimeBetweenVerificationRetries):
		}
	}
}

// verifyConfigVersionURL attempts a one-shot HTTP request/response exchange,
// and checks whether the requested version is present in the response.
func verifyConfigVersionURL(ctx context.Context, url string, version uint64) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return common.NewBasicError("Unable to create HTTP Request to check version:", err)
	}
	subCtx, cancelF := context.WithTimeout(ctx, TimeForHTTPRequest)
	defer cancelF()
	request = request.WithContext(subCtx)

	switch response, err := http.DefaultClient.Do(request); {
	case err != nil:
		return common.NewBasicError("HTTP Request error:", err)
	case response.StatusCode != 200:
		return common.NewBasicError("HTTP Request returned with error code:", nil, "url", url,
			"code", response.StatusCode)
	default:
		// Correctly received HTTP response
		var remoteVersion uint64
		if _, err := fmt.Fscan(response.Body, &remoteVersion); err != nil {
			return common.NewBasicError("Unable to parse HTTP response", err)
		}
		if remoteVersion != version {
			return common.NewBasicError("version mismatch", nil, "actual", remoteVersion,
				"expected", version)
		}
		return nil
	}
}
