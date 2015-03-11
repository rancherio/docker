package main

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRestartStoppedContainer(t *testing.T) {
	defer deleteAllContainers()

	runCmd := exec.Command(dockerBinary, "run", "-d", "busybox", "echo", "foobar")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	cleanedContainerID := stripTrailingCharacters(out)

	runCmd = exec.Command(dockerBinary, "wait", cleanedContainerID)
	if out, _, err = runCommandWithOutput(runCmd); err != nil {
		t.Fatal(out, err)
	}

	runCmd = exec.Command(dockerBinary, "logs", cleanedContainerID)
	out, _, err = runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	if out != "foobar\n" {
		t.Errorf("container should've printed 'foobar'")
	}

	runCmd = exec.Command(dockerBinary, "restart", cleanedContainerID)
	if out, _, err = runCommandWithOutput(runCmd); err != nil {
		t.Fatal(out, err)
	}

	runCmd = exec.Command(dockerBinary, "logs", cleanedContainerID)
	out, _, err = runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	if out != "foobar\nfoobar\n" {
		t.Errorf("container should've printed 'foobar' twice")
	}

	logDone("restart - echo foobar for stopped container")
}

func TestRestartRunningContainer(t *testing.T) {
	defer deleteAllContainers()

	runCmd := exec.Command(dockerBinary, "run", "-d", "busybox", "sh", "-c", "echo foobar && sleep 30 && echo 'should not print this'")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	cleanedContainerID := stripTrailingCharacters(out)

	time.Sleep(1 * time.Second)

	runCmd = exec.Command(dockerBinary, "logs", cleanedContainerID)
	out, _, err = runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	if out != "foobar\n" {
		t.Errorf("container should've printed 'foobar'")
	}

	runCmd = exec.Command(dockerBinary, "restart", "-t", "1", cleanedContainerID)
	if out, _, err = runCommandWithOutput(runCmd); err != nil {
		t.Fatal(out, err)
	}

	runCmd = exec.Command(dockerBinary, "logs", cleanedContainerID)
	out, _, err = runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	time.Sleep(1 * time.Second)

	if out != "foobar\nfoobar\n" {
		t.Errorf("container should've printed 'foobar' twice")
	}

	logDone("restart - echo foobar for running container")
}

// Test that restarting a container with a volume does not create a new volume on restart. Regression test for #819.
func TestRestartWithVolumes(t *testing.T) {
	defer deleteAllContainers()

	runCmd := exec.Command(dockerBinary, "run", "-d", "-v", "/test", "busybox", "top")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	cleanedContainerID := stripTrailingCharacters(out)

	runCmd = exec.Command(dockerBinary, "inspect", "--format", "{{ len .Volumes }}", cleanedContainerID)
	out, _, err = runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	if out = strings.Trim(out, " \n\r"); out != "1" {
		t.Errorf("expect 1 volume received %s", out)
	}

	runCmd = exec.Command(dockerBinary, "inspect", "--format", "{{ .Volumes }}", cleanedContainerID)
	volumes, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(volumes, err)
	}

	runCmd = exec.Command(dockerBinary, "restart", cleanedContainerID)
	if out, _, err = runCommandWithOutput(runCmd); err != nil {
		t.Fatal(out, err)
	}

	runCmd = exec.Command(dockerBinary, "inspect", "--format", "{{ len .Volumes }}", cleanedContainerID)
	out, _, err = runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(out, err)
	}

	if out = strings.Trim(out, " \n\r"); out != "1" {
		t.Errorf("expect 1 volume after restart received %s", out)
	}

	runCmd = exec.Command(dockerBinary, "inspect", "--format", "{{ .Volumes }}", cleanedContainerID)
	volumesAfterRestart, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		t.Fatal(volumesAfterRestart, err)
	}

	if volumes != volumesAfterRestart {
		volumes = strings.Trim(volumes, " \n\r")
		volumesAfterRestart = strings.Trim(volumesAfterRestart, " \n\r")
		t.Errorf("expected volume path: %s Actual path: %s", volumes, volumesAfterRestart)
	}

	logDone("restart - does not create a new volume on restart")
}

func TestRestartPolicyNO(t *testing.T) {
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "-d", "--restart=no", "busybox", "false")
	out, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err, out)
	}

	id := strings.TrimSpace(string(out))
	name, err := inspectField(id, "HostConfig.RestartPolicy.Name")
	if err != nil {
		t.Fatal(err, out)
	}
	if name != "no" {
		t.Fatalf("Container restart policy name is %s, expected %s", name, "no")
	}

	logDone("restart - recording restart policy name for --restart=no")
}

func TestRestartPolicyAlways(t *testing.T) {
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "-d", "--restart=always", "busybox", "false")
	out, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err, out)
	}

	id := strings.TrimSpace(string(out))
	name, err := inspectField(id, "HostConfig.RestartPolicy.Name")
	if err != nil {
		t.Fatal(err, out)
	}
	if name != "always" {
		t.Fatalf("Container restart policy name is %s, expected %s", name, "always")
	}

	logDone("restart - recording restart policy name for --restart=always")
}

func TestRestartPolicyOnFailure(t *testing.T) {
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "-d", "--restart=on-failure:1", "busybox", "false")
	out, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err, out)
	}

	id := strings.TrimSpace(string(out))
	name, err := inspectField(id, "HostConfig.RestartPolicy.Name")
	if err != nil {
		t.Fatal(err, out)
	}
	if name != "on-failure" {
		t.Fatalf("Container restart policy name is %s, expected %s", name, "on-failure")
	}

	logDone("restart - recording restart policy name for --restart=on-failure")
}

func TestRestartOptsOverridesRestartPolicy(t *testing.T) {
	defer deleteAllContainers()

	cmd := exec.Command(dockerBinary, "run", "-d", "--restart=always", "--restart-opt", "policy=on-failure", "--restart-opt", "limit=10", "busybox", "false")
	out, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err, out)
	}

	id := strings.TrimSpace(string(out))
	name, err := inspectField(id, "HostConfig.RestartPolicy.Name")
	if err != nil {
		t.Fatal(err, out)
	}

	if name != "on-failure" {
		t.Fatalf("Container restart policy name is %s, expected %s", name, "on-failure")
	}

	limit, err := inspectField(id, "HostConfig.RestartPolicy.MaximumRetryCount")
	if err != nil {
		t.Fatal(err, out)
	}

	lt, err := strconv.Atoi(limit)
	if err != nil {
		t.Fatal(err, out)
	}

	if lt != 10 {
		t.Fatalf("Container restart policy limit is %s, expected 10", limit)
	}

	logDone("restart-opt - recording restart policy attrs for --restart=on-failure and --restart-opt policy=always")
}

func TestRestartFixedInterval(t *testing.T) {
	cmd := exec.Command(dockerBinary, "run", "-d", "--restart-opt", "policy=on-failure", "--restart-opt", "fixed-interval=10m", "busybox", "false")
	out, _, err := runCommandWithOutput(cmd)
	if err != nil {
		t.Fatal(err, out)
	}

	id := strings.TrimSpace(string(out))
	interval, err := inspectField(id, "HostConfig.RestartPolicy.FixedInterval")
	if err != nil {
		t.Fatal(err, out)
	}

	inter, err := strconv.Atoi(interval)
	if err != nil {
		t.Fatal(err, out)
	}

	if inter != 600 {
		t.Fatalf("Container restart policy fixed-interval is %s, expected 600", interval)
	}
	logDone("restart-opt - recording restart policy attrs for --restart-opt policy=on-failure and setting max-interval")
}

func TestRestartOptsEitherMaxOrFixedInterval(t *testing.T) {
	cmd := exec.Command(dockerBinary, "run", "-d", "--restart-opt", "policy=always", "--restart-opt", "fixed-interval=10s", "--restart-opt", "max-interval=10m", "busybox", "false")
	out, _, err := runCommandWithOutput(cmd)
	if err == nil {
		t.Fatal(err, out)
	}

	logDone("restart-opt - recording restart policy where max-interval and fixed-interval is set, expect error")
}

func TestRestartOptsEitherAlwaysOrMaxRetry(t *testing.T) {
	cmd := exec.Command(dockerBinary, "run", "-d", "--restart-opt", "policy=always", "--restart-opt", "limit=10", "--restart-opt", "max-interval=10m", "busybox", "false")
	out, _, err := runCommandWithOutput(cmd)
	if err == nil {
		t.Fatal(err, out)
	}

	logDone("restart-opt - recording restart policy where policy=always and limit is set, expect error")
}
