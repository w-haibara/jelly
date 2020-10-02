package deployer

import (
	"log"
	"os/exec"
	"strings"
)

func Deploy(arg string) (string, error) {
	log.Println("Deploy...")
	pwd := ""
	if out, err := exec.Command("pwd").Output(); err != nil {
		log.Printf("Command finished with error: %v", err)
		return "", err
	} else {
		pwd = strings.TrimRight(string(out), "\n")
	}
	if arg == "latest" {
		cmd := exec.Command(pwd+"/script/deploy.sh", "latest")
		log.Println("exec:", cmd)
		if out, err := cmd.Output(); err != nil {
			log.Printf("Command finished with error: %v", err)
			return "", err
		} else {
			result := string(out)
			log.Println(result)
			return result, nil
		}
	} else if len(arg) < 20 {
		cmd := exec.Command(pwd+"/script/deploy.sh", arg)
		log.Println("exec:", cmd)
		if out, err := cmd.Output(); err != nil {
			log.Printf("Command finished with error: %v", err)
			return "", err
		} else {
			result := string(out)
			log.Println(result)
			return result, nil
		}
	}

	return "", nil
}
