package deployer

import (
	"log"
	"os/exec"
)

func Deploy(arg string) {
	log.Println("arg:", arg)
	switch arg {
	case "":
		cmd := exec.Command("./script/deploy.sh")
		if err := cmd.Run(); err != nil {
			log.Printf("Command finished with error: %v", err)
		}
	}
}
