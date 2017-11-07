package main

import (
	"fmt"
	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-linereader"
	"io"
)

type Provisioner struct {
	useSudo            bool			   `mapstructure:"use_sudo"`
	MasterIP           string          `mapstructure:"puppetmaster_ip"`

}

const(
	agent_url="https://raw.githubusercontent.com/pyToshka/puppet-install-shell/master/install_puppet_agent.sh"
	filepath="/tmp/install.sh"
)
func (p *Provisioner) Run(o terraform.UIOutput, comm communicator.Communicator) error {



	command:=fmt.Sprintf("'curl %s -o %s -s'", agent_url,filepath)
	fmt.Sprintf(command)
	if err := p.runCommand(o, comm, command); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Validate() error {

	if p.MasterIP == "" {
		return fmt.Errorf("Invalid IP parameter: %s", p.MasterIP)
	}
	return nil
}
func (p *Provisioner) fixPerm(o terraform.UIOutput, comm communicator.Communicator) error {
	err := p.runCommand(o, comm, fmt.Sprintf("'chmod +x %s'", filepath))
	if err != nil {
		return err
	}
	return nil
}
func (p *Provisioner) installPuppetAgent(o terraform.UIOutput, comm communicator.Communicator) error {
	err := p.runCommand(o, comm, fmt.Sprintf("'%s'", filepath))
	if err != nil {
		return err
	}
	return nil
}
func (p *Provisioner) AddPuppetAgentPath(o terraform.UIOutput, comm communicator.Communicator) error {
	err := p.runCommand(o, comm, fmt.Sprintf("'export PATH=/opt/puppetlabs/bin:$PATH'"))
	if err != nil {
		return err
	}
	return nil
}
func (p *Provisioner)RunPuppetAgent(o terraform.UIOutput, comm communicator.Communicator) error {

	err := p.runCommand(o, comm, fmt.Sprintf("'/opt/puppetlabs/bin/puppet resource host puppet ensure=present ip=%s ;/opt/puppetlabs/bin/puppet agent --enable; /opt/puppetlabs/bin/puppet agent -t'",p.MasterIP))
	if err != nil {
		return err
	}
	return nil
}

func (p *Provisioner) runCommand(
	o terraform.UIOutput,
	comm communicator.Communicator,
	command string) error {
	var err error
	if p.useSudo {
		command = "sudo -i bash -c " + command
	}

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()
	outDoneCh := make(chan struct{})
	errDoneCh := make(chan struct{})

	go p.copyOutput(o, outR, outDoneCh)
	go p.copyOutput(o, errR, errDoneCh)

	cmd := &remote.Cmd{
		Command: command,
		Stdout:  outW,
		Stderr:  errW,
	}

	if err := comm.Start(cmd); err != nil {
		return fmt.Errorf("Error executing command %q: %v", cmd.Command, err)
	}
	cmd.Wait()
	if cmd.ExitStatus != 0 {
		err = fmt.Errorf(
			"Command %q exited with non-zero exit status: %d", cmd.Command, cmd.ExitStatus)
	}

	outW.Close()
	errW.Close()
	<-outDoneCh
	<-errDoneCh

	return err
}


func (p *Provisioner) copyOutput(o terraform.UIOutput, r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)
	lr := linereader.New(r)
	for line := range lr.Ch {
		o.Output(line)
	}
}
