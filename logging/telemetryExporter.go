package logging

import (
	"os"
	"path"
	"path/filepath"
)

const (
	redirectTelemetryOutput = false
)

type TelemetryExporter struct {
	exporterProcess string
	dataDir         string
	process         *os.Process
	telemetryConfig *TelemetryConfig
}

func MakeTelemetryExporter(executable string, dataDir string, telemetryConfig *TelemetryConfig) *TelemetryExporter {
	return &TelemetryExporter{
		exporterProcess: executable,
		dataDir:         dataDir,
		telemetryConfig: telemetryConfig,
		process:         nil,
	}
}

func (te *TelemetryExporter) EnsureRunning() error {
	if ! te.isRunning() {
		configFile := path.Join(te.dataDir, "telemetry.yml")
		err := te.ensureConfig(configFile)
		if err != nil {
			return err
		}
		te.process = nil

		telemDir := path.Join(te.dataDir, "telem")
		os.Mkdir(telemDir, os.ModePerm)

		args := make([]string, 0)
		args = append(args, te.exporterProcess)
		args = append(args, "-c")
		args = append(args, configFile)
		args = append(args, "--path.home")
		args = append(args, telemDir)

		attributes := os.ProcAttr{
			Dir: filepath.Dir(os.Args[0]),
			Env: os.Environ(),
		}

		if redirectTelemetryOutput {
			attributes.Files = []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			}
		}

		te.process, err = os.StartProcess(te.exporterProcess, args, &attributes)
		if err != nil {
			te.process = nil
			return err
		}

		// wait for the process to complete on a separate goroutine, clear the process variable to nil once it's done.
		go func(proc **os.Process) {
			(*proc).Wait()
			(*proc) = nil
		}(&te.process)
	}
	return nil
}

func (te *TelemetryExporter) isRunning() bool {
	return te.process != nil
}

func (te *TelemetryExporter) ensureConfig(filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	config :=
		"filebeat.inputs:\n" +
		"- type: log\n" +
		"  enabled: true\n" +
		"  paths:\n" +
		"  - " + te.dataDir + "*.log\n" +
		"output.elasticsearch:\n" +
		"  protocol: \"https\"\n" +
		"  hosts: [\"" + te.telemetryConfig.URI + "\"]\n" +
		"  username: \""+ te.telemetryConfig.UserName + "\"\n" +
		"  password: \""+ te.telemetryConfig.Password +"\"\n" +
		"\n" +
		"processors:\n" +
		"  - add_host_metadata: ~\n" +
		"  - add_cloud_metadata: ~\n" +
		"  - add_log_history: ~\n"

	_, err = f.Write([]byte(config))
	return err
}
