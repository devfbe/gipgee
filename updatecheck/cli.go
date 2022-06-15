package updatecheck

import "fmt"

type UpdateCheckCmd struct {
}

func (r *UpdateCheckCmd) Run() error {
	fmt.Println("UpdateCheckCmd release")
	return nil
}

func (r *UpdateCheckCmd) Help() string {
	return "Generates the update check pipeline"
}
