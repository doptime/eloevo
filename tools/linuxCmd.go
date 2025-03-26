package tools

import (
	"log"
	"os/exec"

	"github.com/doptime/eloevo/tool"
)

type LinuxCommand struct {
	Command string   `json:"command,omitempty" description:"The Linux command to execute"`
	Args    []string `json:"args,omitempty" description:"Arguments for the command"`
}

func executeCommand(param *LinuxCommand) {
	if param.Command == "" {
		log.Println("No command provided to execute.")
		return
	}

	// Create the command with arguments
	cmd := exec.Command(param.Command, param.Args...)

	// Combine stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing command '%s': %v\nOutput: %s", param.Command, err, string(output))
		return
	}

	// Log the successful output
	log.Printf("Command '%s' executed successfully.\nOutput:\n%s", param.Command, string(output))
}

var CmdTool = tool.NewTool("linuxcmd", "After all revisions are Ready, commit the file to the directory using this tool", executeCommand)
