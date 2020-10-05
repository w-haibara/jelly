package deployer

import (
	"log"
	"os/exec"
	"strings"
)

// Deploy projects
func Deploy(arg string) (string, error) {
	log.Println("Deploy...")
	pwd := ""
	if out, err := exec.Command("pwd").Output(); err == nil {
		pwd = strings.TrimRight(string(out), "\n")
	} else {
		log.Printf("Command finished with error: %v", err)
		return "", err
	}
	if arg == "latest" {
		cmd := exec.Command(pwd+"/script/deploy.sh", "latest")
		log.Println("exec:", cmd)
		out, err := cmd.Output()
		if err != nil {
			log.Printf("Command finished with error: %v", err)
			return "", err
		}
		result := string(out)
		log.Println(result)
		return result, nil
	} else if len(arg) < 20 {
		cmd := exec.Command(pwd+"/script/deploy.sh", arg)
		log.Println("exec:", cmd)
		out, err := cmd.Output()
		if err != nil {
			log.Printf("Command finished with error: %v", err)
			return "", err
		}
		result := string(out)
		log.Println(result)
		return result, nil
	}

	return "", nil
}
