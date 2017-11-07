package main

import (
	"fmt"
	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
)

type ResourceProvisioner struct {}

func (r *ResourceProvisioner) Apply(
	o terraform.UIOutput,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig) error {

	provisioner, err := r.decodeConfig(c)
	if err != nil {
		o.Output("erred out here")
		return err
	}

	err = provisioner.Validate()
	if err != nil {
		o.Output("Invalid provisioner configuration settings")
		return err
	}
	provisioner.useSudo = true
	// ensure that this is a linux machine
	if s.Ephemeral.ConnInfo["type"] != "ssh" {
		return fmt.Errorf("Unsupported connection type: %s. This provisioner currently only supports linux", s.Ephemeral.ConnInfo["type"])
	}

	// build a communicator for the provisioner to use
	comm, err := communicator.New(s)
	if err != nil {
		o.Output("erred out here 3")
		return err
	}

	err = retryFunc(comm.Timeout(), func() error {
		err := comm.Connect(o)
		return err
	})
	if err != nil {
		return err
	}
	defer comm.Disconnect()

	if err :=provisioner.Run(o, comm); err != nil {
		return err
	}
	if err := provisioner.fixPerm(o, comm); err != nil {
		return err
	}
	if err := provisioner.AddPuppetAgentPath(o, comm); err != nil {
		return err
	}
	if err := provisioner.installPuppetAgent(o, comm); err != nil {
		return err
	}
	if err := provisioner.RunPuppetAgent(o, comm); err != nil {
		return err
	}
	return nil
}

func (r *ResourceProvisioner) Validate(c *terraform.ResourceConfig) (ws []string, es []error) {
	provisioner, err := r.decodeConfig(c)
	if err != nil {
		es = append(es, err)
		return ws, es
	}

	err = provisioner.Validate()
	if err != nil {
		es = append(es, err)
		return ws, es
	}

	return ws, es
}

func (r *ResourceProvisioner) decodeConfig(c *terraform.ResourceConfig) (*Provisioner, error) {
	// decodes configuration from terraform and builds out a provisioner
	p := new(Provisioner)
	decoderConfig := &mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		Result:           p,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	for k, v := range c.Raw {
		m[k] = v
	}
	for k, v := range c.Config {
		m[k] = v
	}

	err = decoder.Decode(m)
	if err != nil {
		return nil, err
	}

	return p, nil
}
func (p *ResourceProvisioner) Stop() error {
	return nil
}
func retryFunc(timeout time.Duration, f func() error) error {
	finish := time.After(timeout)

	for {
		err := f()
		if err == nil {
			return nil
		}
		log.Printf("Retryable error: %v", err)

		select {
		case <-finish:
			return err
		case <-time.After(3 * time.Second):
		}
	}
}